export function StatusDot({ status = 'disconnected' }) {
  return <span className={`status-dot ${status}`} title={status} />
}

