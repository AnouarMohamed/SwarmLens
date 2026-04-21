import { FormEvent, useEffect, useMemo, useState } from 'react'
import { api } from '../../lib/api'
import { relativeTime } from '../../lib/utils'
import { useClusterStore } from '../../store/clusterStore'
import { useControlPlaneStore } from '../../store/controlPlaneStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useSessionStore } from '../../store/sessionStore'
import type {
  AssistantActionProposal,
  AssistantCitation,
  AssistantMessage,
  AssistantSession,
  InsightAction,
  InsightHypothesis,
  OpsInsights,
} from '../../types'

const QUICK_PROMPTS = [
  'What needs action right now?',
  'Summarize current risk and confidence.',
  'What are the top remediation steps?',
]

const INTRO_MESSAGE: AssistantMessage = {
  id: 'intro',
  sessionID: 'local',
  role: 'assistant',
  content:
    'Ops copilot is ready. Ask for cluster triage, service risk, or a remediation sequence and I will keep the conversation attached to this cluster.',
  createdAt: new Date(0).toISOString(),
}

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

function mergeSession(session: AssistantSession, sessions: AssistantSession[]) {
  const next = [session, ...sessions.filter((item) => item.id !== session.id)]
  return next.sort((a, b) => +new Date(b.updatedAt) - +new Date(a.updatedAt))
}

export function AssistantPanel() {
  const { swarm, connectionState, error, fetchAll } = useClusterStore()
  const { findings, fetch: fetchDiagnostics } = useDiagnosticsStore()
  const { clusters, selectedClusterID, refreshWorkflow } = useControlPlaneStore()
  const { me } = useSessionStore()

  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [sessions, setSessions] = useState<AssistantSession[]>([])
  const [activeSessionID, setActiveSessionID] = useState('')
  const [messages, setMessages] = useState<AssistantMessage[]>([INTRO_MESSAGE])
  const [insight, setInsight] = useState<OpsInsights | null>(null)
  const [hypotheses, setHypotheses] = useState<InsightHypothesis[]>([])
  const [actions, setActions] = useState<InsightAction[]>([])
  const [citations, setCitations] = useState<AssistantCitation[]>([])
  const [proposals, setProposals] = useState<AssistantActionProposal[]>([])
  const [statusLine, setStatusLine] = useState('')
  const [executingProposal, setExecutingProposal] = useState('')

  const activeCluster =
    clusters.find((cluster) => cluster.id === selectedClusterID) ??
    clusters.find((cluster) => cluster.default) ??
    clusters[0] ??
    null

  const disconnected = connectionState === 'disconnected' || Boolean(error)
  const canOperate = me ? me.authenticated && me.role !== 'viewer' : true
  const contextLine = useMemo(() => {
    if (activeCluster) {
      return `${activeCluster.name} | ${activeCluster.connectionMode.toUpperCase()} | ${activeCluster.health.freshness.toUpperCase()}`
    }
    if (!swarm) return 'No cluster context available.'
    return `cluster/${swarm.clusterID.slice(0, 12)} | ${swarm.mode.toUpperCase()} | ${swarm.freshness.toUpperCase()}`
  }, [activeCluster, swarm])

  async function loadSession(sessionID: string) {
    const session = await api.assistant.getSession(sessionID)
    setActiveSessionID(session.id)
    setMessages(session.messages?.length ? session.messages : [INTRO_MESSAGE])
    const lastAssistant = [...(session.messages ?? [])].reverse().find((message) => message.role === 'assistant')
    setCitations(lastAssistant?.citations ?? [])
    setProposals(lastAssistant?.actionProposals ?? [])
  }

  async function loadSessions(preferredSessionID?: string) {
    const nextSessions = await api.assistant.sessions()
    setSessions(nextSessions)
    const nextID =
      preferredSessionID ??
      nextSessions.find((session) => session.id === activeSessionID)?.id ??
      nextSessions[0]?.id ??
      ''
    if (nextID) {
      await loadSession(nextID)
      return
    }
    setActiveSessionID('')
    setMessages([INTRO_MESSAGE])
    setCitations([])
    setProposals([])
  }

  async function createSession() {
    const session = await api.assistant.createSession({
      title: activeCluster ? `${activeCluster.name} triage` : 'Ops Copilot Session',
    })
    setSessions((state) => mergeSession(session, state))
    setActiveSessionID(session.id)
    setMessages([INTRO_MESSAGE])
    setCitations([])
    setProposals([])
    return session
  }

  useEffect(() => {
    let cancelled = false

    async function syncSessions() {
      const nextSessions = await api.assistant.sessions()
      if (cancelled) return
      setSessions(nextSessions)
      const nextID =
        nextSessions.find((session) => session.id === activeSessionID)?.id ??
        nextSessions[0]?.id ??
        ''
      if (!nextID) {
        setActiveSessionID('')
        setMessages([INTRO_MESSAGE])
        setCitations([])
        setProposals([])
        return
      }
      const session = await api.assistant.getSession(nextID)
      if (cancelled) return
      setActiveSessionID(session.id)
      setMessages(session.messages?.length ? session.messages : [INTRO_MESSAGE])
      const lastAssistant = [...(session.messages ?? [])].reverse().find((message) => message.role === 'assistant')
      setCitations(lastAssistant?.citations ?? [])
      setProposals(lastAssistant?.actionProposals ?? [])
    }

    void syncSessions()
    return () => {
      cancelled = true
    }
  }, [activeSessionID, selectedClusterID])

  async function submitPrompt(prompt: string) {
    const clean = prompt.trim()
    if (!clean || loading) return

    setLoading(true)
    setStatusLine('')
    setHypotheses([])
    setActions([])
    setInsight(null)
    setCitations([])
    setProposals([])

    const tempMessage: AssistantMessage = {
      id: `u-${Date.now()}`,
      sessionID: activeSessionID || 'pending',
      role: 'user',
      content: clean,
      createdAt: new Date().toISOString(),
    }
    setMessages((state) => [...state.filter((message) => message.id !== 'intro'), tempMessage])

    let sessionID = activeSessionID
    if (!sessionID) {
      const session = await createSession()
      sessionID = session.id
    }

    try {
      await api.assistant.chatStream(
        { prompt: clean, sessionID },
        {
          onEvent: (event, payload) => {
            if (event === 'session' && payload && typeof payload === 'object') {
              const session = payload as AssistantSession
              sessionID = session.id
              setSessions((state) => mergeSession(session, state))
              setActiveSessionID(session.id)
              return
            }
            if (event === 'insight' && payload && typeof payload === 'object') {
              setInsight(payload as OpsInsights)
              return
            }
            if (event === 'hypothesis' && payload && typeof payload === 'object') {
              setHypotheses((state) => [...state, payload as InsightHypothesis])
              return
            }
            if (event === 'action' && payload && typeof payload === 'object') {
              setActions((state) => [...state, payload as InsightAction])
              return
            }
            if (event === 'citation' && payload && typeof payload === 'object') {
              setCitations((state) => [...state, payload as AssistantCitation])
              return
            }
            if (event === 'action_proposal' && payload && typeof payload === 'object') {
              setProposals((state) => [...state, payload as AssistantActionProposal])
              return
            }
            if (event === 'message' && payload && typeof payload === 'object') {
              const message = payload as AssistantMessage
              setMessages((state) => {
                const withoutIntro = state.filter((item) => item.id !== 'intro')
                return [...withoutIntro.filter((item) => item.id !== message.id), message]
              })
            }
          },
          onDone: () => {
            setStatusLine('Assistant response stored with citations and ready-to-run proposals.')
          },
          onError: (message) => {
            setStatusLine(`Assistant stream failed: ${message}`)
          },
        },
      )
      await loadSessions(sessionID)
    } finally {
      setInput('')
      setLoading(false)
    }
  }

  async function executeProposal(proposal: AssistantActionProposal) {
    setExecutingProposal(`${proposal.action}:${proposal.resourceID ?? ''}`)
    try {
      const outcome = await api.actions.execute({
        action: proposal.action,
        resource: proposal.resource,
        resourceID: proposal.resourceID,
        reason: proposal.reason,
        params: proposal.params,
      })
      setStatusLine(`${proposal.title}: ${outcome.status} · ${outcome.message}`)
      await Promise.all([refreshWorkflow(), fetchAll(), fetchDiagnostics()])
    } finally {
      setExecutingProposal('')
    }
  }

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void submitPrompt(input)
  }

  return (
    <div className="space-y-8">
      <section className="industrial-section border-b-0 pt-0">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="industrial-label">Assistant</p>
            <h2 className="mt-3 font-heading text-[2rem] uppercase leading-none tracking-[0.04em]">
              Ops Copilot
            </h2>
            <p className="mt-3 text-sm text-text-secondary">{contextLine}</p>
            <p className={cn('mt-1 text-sm', disconnected ? 'text-state-danger' : 'text-text-secondary')}>
              {disconnected
                ? 'Cluster is disconnected. Responses are based on stored context and last known state.'
                : `${findings.length} findings and persistent session history are available to the copilot.`}
            </p>
            {statusLine ? <p className="mt-2 text-sm text-text-secondary">{statusLine}</p> : null}
          </div>
          <button type="button" onClick={() => { void createSession() }} className="industrial-action">
            New Session
          </button>
        </div>
      </section>

      <section>
        <div className="flex flex-wrap gap-3">
          {sessions.map((session) => (
            <button
              key={session.id}
              type="button"
              onClick={() => {
                void loadSession(session.id)
              }}
              className={cn(
                'industrial-action',
                activeSessionID === session.id && 'industrial-action-accent',
              )}
            >
              {session.title}
            </button>
          ))}
          {sessions.length === 0 ? (
            <span className="industrial-label text-text-tertiary">No saved sessions for this cluster yet.</span>
          ) : null}
        </div>
      </section>

      <section>
        <div className="flex flex-wrap gap-3">
          {QUICK_PROMPTS.map((prompt) => (
            <button
              key={prompt}
              type="button"
              onClick={() => {
                void submitPrompt(prompt)
              }}
              disabled={loading}
              className={cn('industrial-action', loading && 'cursor-not-allowed opacity-35')}
            >
              {prompt}
            </button>
          ))}
        </div>
      </section>

      {insight ? (
        <section className="border-t border-border-muted pt-5">
          <p className="industrial-label">Current Insight</p>
          <p className="mt-2 text-sm text-text-primary">{insight.summary}</p>
          <p className="mt-2 industrial-label text-text-secondary">
            Risk {insight.risk.score.toFixed(2)} · Confidence {(insight.risk.confidence * 100).toFixed(0)}% · Provider {insight.provider.toUpperCase()}
          </p>
        </section>
      ) : null}

      {hypotheses.length > 0 ? (
        <section className="border-t border-border-muted pt-5">
          <p className="industrial-label">Hypotheses</p>
          <ul className="mt-3 space-y-2">
            {hypotheses.slice(0, 5).map((hypothesis, idx) => (
              <li key={`${hypothesis.title}-${idx}`} className="text-sm text-text-secondary">
                <span className="text-text-primary">{hypothesis.title}</span>
                <span className="ml-2 industrial-data text-xs text-text-tertiary">
                  {Math.round(hypothesis.confidence * 100)}%
                </span>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {actions.length > 0 ? (
        <section className="border-t border-border-muted pt-5">
          <p className="industrial-label">Recommended Actions</p>
          <ul className="mt-3 space-y-2">
            {actions.slice(0, 5).map((action, idx) => (
              <li key={`${action.title}-${idx}`} className="text-sm text-text-secondary">
                <span className="text-text-primary">{action.title}</span>
                <span className="ml-2 industrial-label text-text-tertiary">{action.actionability}</span>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {citations.length > 0 ? (
        <section className="border-t border-border-muted pt-5">
          <p className="industrial-label">Citations</p>
          <ul className="mt-3 divide-y divide-border-muted">
            {citations.map((citation) => (
              <li key={`${citation.kind}-${citation.id}`} className="industrial-row">
                <p className="industrial-label text-text-secondary">
                  {citation.kind} · {citation.locator}
                </p>
                <p className="mt-2 text-sm text-text-primary">{citation.title}</p>
                {citation.snippet ? (
                  <p className="mt-2 text-sm text-text-secondary">{citation.snippet}</p>
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {proposals.length > 0 ? (
        <section className="border-t border-border-muted pt-5">
          <p className="industrial-label">Action Proposals</p>
          <ul className="mt-3 divide-y divide-border-muted">
            {proposals.map((proposal, idx) => {
              const proposalKey = `${proposal.action}:${proposal.resourceID ?? idx}`
              return (
                <li key={proposalKey} className="industrial-row">
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div>
                      <p className="industrial-label text-text-secondary">
                        {proposal.action}
                        {proposal.resourceID ? ` · ${proposal.resource}/${proposal.resourceID}` : ''}
                      </p>
                      <p className="mt-2 text-sm text-text-primary">{proposal.title}</p>
                      <p className="mt-2 text-sm text-text-secondary">{proposal.reason}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => {
                        void executeProposal(proposal)
                      }}
                      disabled={!canOperate || executingProposal === proposalKey}
                      className={cn(
                        'industrial-action industrial-action-accent',
                        (!canOperate || executingProposal === proposalKey) && 'cursor-not-allowed opacity-35',
                      )}
                    >
                      {executingProposal === proposalKey
                        ? 'Submitting...'
                        : proposal.requiresApproval
                          ? 'Submit For Approval'
                          : 'Run Action'}
                    </button>
                  </div>
                </li>
              )
            })}
          </ul>
        </section>
      ) : null}

      <section>
        <ul className="divide-y divide-border-muted">
          {messages.map((message) => (
            <li key={message.id} className="industrial-row">
              <p
                className={cn(
                  'industrial-label',
                  message.role === 'assistant' ? 'text-text-secondary' : 'text-state-danger',
                )}
              >
                {message.role === 'assistant' ? 'ASSISTANT' : 'YOU'}
              </p>
              <p className="mt-2 text-sm text-text-primary">{message.content}</p>
              {message.createdAt !== INTRO_MESSAGE.createdAt ? (
                <p className="mt-2 industrial-data text-xs text-text-tertiary">
                  {relativeTime(message.createdAt)}
                </p>
              ) : null}
            </li>
          ))}
        </ul>
      </section>

      <section>
        <form onSubmit={onSubmit} className="flex gap-3 border-t border-border-muted pt-4">
          <input
            className="h-10 flex-1 border-b border-border-muted bg-transparent px-0 text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none"
            placeholder="Ask about risk, incidents, approvals, and remediation plans..."
            value={input}
            onChange={(event) => setInput(event.target.value)}
            disabled={loading}
          />
          <button
            type="submit"
            disabled={loading}
            className={cn('industrial-action industrial-action-accent', loading && 'cursor-not-allowed opacity-35')}
          >
            {loading ? 'Streaming...' : 'Send'}
          </button>
        </form>
      </section>
    </div>
  )
}
