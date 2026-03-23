export function Card({ title, children, action, accent, className = '' }) {
  return (
    <div className={`card ${accent ? `card-accent-${accent}` : ''} ${className}`}>
      {title && (
        <div className="card-header">
          <span className="card-title">{title}</span>
          {action && <div className="card-action">{action}</div>}
        </div>
      )}
      {children}
    </div>
  )
}

