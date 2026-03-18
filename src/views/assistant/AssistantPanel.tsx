import { FormEvent, useMemo, useState } from 'react'
import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'

type Message = { id: string; role: 'user' | 'assistant'; text: string }

const QUICK_PROMPTS = [
  'What needs action right now?',
  'Summarize cluster health in one paragraph.',
  'Which services are risky?',
]

function makeAssistantReply({
  prompt,
  disconnected,
  riskyServices,
  criticalFindings,
  warningFindings,
  pendingTasks,
}: {
  prompt: string
  disconnected: boolean
  riskyServices: number
  criticalFindings: number
  warningFindings: number
  pendingTasks: number
}) {
  const normalized = prompt.toLowerCase()

  if (disconnected) {
    return 'SwarmLens is disconnected. Reconnect first to resume live telemetry, then run diagnostics and refresh services.'
  }

  if (normalized.includes('action') || normalized.includes('now')) {
    if (criticalFindings > 0 || riskyServices > 0) {
      return `Prioritize ${criticalFindings} critical findings and ${riskyServices} risky services. Inspect incidents, then confirm remediation with a diagnostics run.`
    }
    if (pendingTasks > 0) {
      return `${pendingTasks} tasks are pending. Inspect scheduling constraints and node availability before scaling.`
    }
    return 'No urgent actions detected. Keep a steady diagnostics cadence and review audit trail before planned changes.'
  }

  if (normalized.includes('risky') || normalized.includes('service')) {
    if (riskyServices === 0) return 'No services are currently flagged as risky.'
    return `${riskyServices} services show replica drift or task failures. Review Services, then open Incidents if impact continues.`
  }

  return `Cluster summary: ${criticalFindings} critical, ${warningFindings} warning findings, ${pendingTasks} pending tasks, ${riskyServices} risky services.`
}

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function AssistantPanel() {
  const { swarm, services, tasks, connectionState, error } = useClusterStore()
  const { findings } = useDiagnosticsStore()
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'intro',
      role: 'assistant',
      text: 'Ops copilot is running in local mode. Ask for triage, health summaries, or next steps.',
    },
  ])

  const disconnected = connectionState === 'disconnected' || Boolean(error)
  const riskyServices = services.filter(
    (service) => service.runningTasks < service.desiredReplicas || service.failedTasks > 0,
  ).length
  const pendingTasks = tasks.filter((task) =>
    ['pending', 'assigned', 'accepted', 'preparing', 'starting'].includes(task.currentState),
  ).length
  const criticalFindings = findings.filter((finding) => finding.severity === 'critical').length
  const warningFindings = findings.filter(
    (finding) => finding.severity === 'high' || finding.severity === 'medium',
  ).length

  const contextLine = useMemo(() => {
    if (!swarm) return 'No cluster context available.'
    return `cluster/${swarm.clusterID.slice(0, 12)} | ${swarm.mode.toUpperCase()} | ${disconnected ? 'disconnected' : 'connected'}`
  }, [swarm, disconnected])

  function submitPrompt(prompt: string) {
    const clean = prompt.trim()
    if (!clean) return
    const userMessage: Message = { id: `u-${Date.now()}`, role: 'user', text: clean }
    const assistantMessage: Message = {
      id: `a-${Date.now() + 1}`,
      role: 'assistant',
      text: makeAssistantReply({
        prompt: clean,
        disconnected,
        riskyServices,
        criticalFindings,
        warningFindings,
        pendingTasks,
      }),
    }
    setMessages((state) => [...state, userMessage, assistantMessage])
    setInput('')
  }

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    submitPrompt(input)
  }

  return (
    <div className="space-y-8">
      <section className="industrial-section border-b-0 pt-0">
        <p className="industrial-label">Assistant</p>
        <h2 className="mt-3 font-heading text-[2rem] uppercase leading-none tracking-[0.04em]">
          Ops Copilot
        </h2>
        <p className="mt-3 text-sm text-text-secondary">{contextLine}</p>
      </section>

      <section>
        <div className="flex flex-wrap gap-3">
          {QUICK_PROMPTS.map((prompt) => (
            <button
              key={prompt}
              type="button"
              onClick={() => submitPrompt(prompt)}
              className="industrial-action"
            >
              {prompt}
            </button>
          ))}
        </div>
      </section>

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
            placeholder="Ask about health, incidents, or remediation steps..."
            value={input}
            onChange={(event) => setInput(event.target.value)}
          />
          <button type="submit" className="industrial-action industrial-action-accent">
            Send
          </button>
        </form>
      </section>
    </div>
  )
}
