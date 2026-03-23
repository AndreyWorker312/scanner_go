import { Inbox } from 'lucide-react'

export function EmptyState({ icon: Icon = Inbox, title = 'No data', description }) {
  return (
    <div className="empty-state">
      <Icon size={40} />
      <div>
        <div style={{ fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 4 }}>{title}</div>
        {description && <div>{description}</div>}
      </div>
    </div>
  )
}

