import { useState } from 'react'
import { useClusterStore } from '../../store/clusterStore'

export function AssistantPanel() {
  const swarm = useClusterStore(s => s.swarm)
  const [msg, setMsg] = useState('')

  if (!swarm) return null

  return (
    <div className="view">
      <div className="assistant-notice">
        <span className="dim">
          Assistant is disabled by default. Set <span className="mono">ASSISTANT_PROVIDER=openai_compatible</span> and
          configure <span className="mono">ASSISTANT_API_KEY</span> to enable.
          The assistant is grounded in live diagnostic findings and cluster state — not raw data.
        </span>
      </div>
      <div className="assistant-input-row">
        <input
          className="assistant-input"
          placeholder="Ask about cluster health, failures, or remediation steps…"
          value={msg}
          onChange={e => setMsg(e.target.value)}
          disabled
        />
        <button className="btn-primary" disabled>Send</button>
      </div>
    </div>
  )
}
