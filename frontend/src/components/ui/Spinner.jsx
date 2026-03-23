export function Spinner({ size = 'md' }) {
  return <div className={`spinner ${size === 'lg' ? 'spinner-lg' : ''}`} />
}

export function ScanAnimation({ label = 'Scanning…' }) {
  return (
    <div className="scan-in-progress">
      <div className="scan-ring" />
      <div className="scan-text">{label}</div>
    </div>
  )
}

