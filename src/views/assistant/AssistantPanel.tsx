import { FormEvent, useMemo, useState } from 'react'
import { api } from '../../lib/api'
import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import type { InsightAction, InsightHypothesis, OpsInsights } from '../../types'

type Message = { id: string; role: 'user' | 'assistant'; text: string }

const QUICK_PROMPTS = [
  'What needs action right now?',
  'Summarize current risk and confidence.',
  'What are the top remediation steps?',
]

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function AssistantPanel() {
  const { swarm, connectionState, error } = useClusterStore()
  const { findings } = useDiagnosticsStore()
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'intro',
      role: 'assistant',
      text: 'Ops copilot is connected. Ask for triage, risk narrative, and remediation sequencing.',
    },
  ])
  const [insight, setInsight] = useState<OpsInsights | null>(null)
  const [hypotheses, setHypotheses] = useState<InsightHypothesis[]>([])
  const [actions, setActions] = useState<InsightAction[]>([])

  const disconnected = connectionState === 'disconnected' || Boolean(error)
  const contextLine = useMemo(() => {
    if (!swarm) return 'No cluster context available.'
    return `cluster/${swarm.clusterID.slice(0, 12)} | ${swarm.mode.toUpperCase()} | ${swarm.freshness.toUpperCase()}`
  }, [swarm])

  async function submitPrompt(prompt: string) {
    const clean = prompt.trim()
    if (!clean || loading) return

    setLoading(true)
    setHypotheses([])
    setActions([])
    setInsight(null)
    setMessages((state) => [...state, { id: `u-${Date.now()}`, role: 'user', text: clean }])

    await api.assistant.chatStream(clean, {
      onEvent: (event, payload) => {
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
        if (event === 'message' && payload && typeof payload === 'object') {
          const content = (payload as { content?: string }).content ?? ''
          if (!content) return
          setMessages((state) => [
            ...state,
            { id: `a-${Date.now()}-${Math.random()}`, role: 'assistant', text: content },
          ])
        }
      },
      onDone: () => {
        setLoading(false)
      },
      onError: (message) => {
        setLoading(false)
        setMessages((state) => [
          ...state,
          {
            id: `err-${Date.now()}`,
            role: 'assistant',
            text: `Assistant stream failed: ${message}. Falling back to latest deterministic context.`,
          },
        ])
      },
    })
    setInput('')
    setLoading(false)
  }

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void submitPrompt(input)
  }

  return (
    <div className="space-y-8">
      <section className="industrial-section border-b-0 pt-0">
        <p className="industrial-label">Assistant</p>
        <h2 className="mt-3 font-heading text-[2rem] uppercase leading-none tracking-[0.04em]">
          Ops Copilot
        </h2>
        <p className="mt-3 text-sm text-text-secondary">{contextLine}</p>
        <p className={cn('mt-1 text-sm', disconnected ? 'text-state-danger' : 'text-text-secondary')}>
          {disconnected
            ? 'Cluster is disconnected. Responses are generated from last known state.'
            : `${findings.length} findings in current context.`}
        </p>
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
            {hypotheses.slice(0, 5).map((h, idx) => (
              <li key={`${h.title}-${idx}`} className="text-sm text-text-secondary">
                <span className="text-text-primary">{h.title}</span>
                <span className="ml-2 industrial-data text-xs text-text-tertiary">{Math.round(h.confidence * 100)}%</span>
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
              <p className="mt-2 text-sm text-text-primary">{message.text}</p>
            </li>
          ))}
        </ul>
      </section>

      <section>
        <form onSubmit={onSubmit} className="flex gap-3 border-t border-border-muted pt-4">
          <input
            className="h-10 flex-1 border-b border-border-muted bg-transparent px-0 text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none"
            placeholder="Ask about risk, incidents, and remediation plans..."
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
