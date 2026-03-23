const STATUS_COLOR = {
  completed: 'green', success: 'green', online: 'green', up: 'green',
  error: 'red', failed: 'red', offline: 'red', down: 'red',
  scanning: 'yellow', pending: 'yellow', running: 'yellow',
  tcp: 'blue', udp: 'blue', info: 'blue',
  nmap: 'purple', os: 'purple',
}

export function Badge({ children, color, dot = true }) {
  const c = color ?? STATUS_COLOR[String(children).toLowerCase()] ?? 'gray'
  return (
    <span className={`badge badge-${c}`}>
      {dot && <span className="badge-dot" />}
      {children}
    </span>
  )
}

