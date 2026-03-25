import { useState, useEffect } from 'react'
import {
  Bell, AlertTriangle, AlertCircle, Info,
  Wifi, WifiOff, Trash2, RefreshCw, Filter,
} from 'lucide-react'
import { useChangesSSE }                      from '@/hooks/useChangesSSE'
import { api }                                from '@/api/http'
import { Button, Badge, EmptyState, Spinner } from '@/components/ui'
import { format }                             from 'date-fns'

// ── severity metadata ─────────────────────────────────────────────────────────
const SEV = {
  CRITICAL: { color: 'red',    Icon: AlertTriangle, label: 'CRITICAL', rank: 0 },
  HIGH:     { color: 'yellow', Icon: AlertCircle,   label: 'HIGH',     rank: 1 },
  MEDIUM:   { color: 'blue',   Icon: Info,          label: 'MEDIUM',   rank: 2 },
  LOW:      { color: 'gray',   Icon: Info,          label: 'LOW',      rank: 3 },
}

const EVENT_LABELS = {
  NEW_DEVICE:     'Новое устройство',
  DEVICE_GONE:    'Устройство пропало',
  NEW_PORT:       'Новый порт',
  PORT_CLOSED:    'Порт закрыт',
  VERSION_CHANGE: 'Изменение сервиса',
}


// ── sub-components ───────────────────────────────────────────────────────────

function SeverityBadge({ severity }) {
  const cfg = SEV[severity] ?? SEV.LOW
  return <Badge color={cfg.color}>{severity}</Badge>
}

function ChangeCard({ event }) {
  const cfg  = SEV[event.severity] ?? SEV.LOW
  const Icon = cfg.Icon
  const date = event.created_at ? new Date(event.created_at) : null

  return (
    <div className={`change-card change-card-${cfg.color}`}>
      <div className="change-card-body">
        <div className={`change-card-icon change-icon-${cfg.color}`}>
          <Icon size={15} />
        </div>

        <div className="change-card-content">
          <div className="change-card-title">{event.title}</div>
          <div className="change-card-desc">{event.description}</div>

          <div className="change-card-meta">
            <span className="change-type-tag">
              {EVENT_LABELS[event.event_type] ?? event.event_type}
            </span>
            <span className="change-target">📍 {event.target}</span>
            <span className="change-scanner">via {event.scanner?.toUpperCase()}</span>
          </div>

          <div className="change-action">{event.action}</div>
        </div>

        <div className="change-card-side">
          <SeverityBadge severity={event.severity} />
          {date && (
            <div className="change-date">
              {format(date, 'dd.MM.yy')}<br />
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11 }}>
                {format(date, 'HH:mm:ss')}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function StatCard({ severity, label, count, active, onClick }) {
  const cfg = SEV[severity] ?? SEV.LOW
  return (
    <button className={`changes-stat-card${active ? ' active' : ''}`} onClick={onClick}>
      <div className={`changes-stat-num changes-stat-${cfg.color}`}>{count}</div>
      <div className="changes-stat-label">{label}</div>
    </button>
  )
}

// ── main page ────────────────────────────────────────────────────────────────

export default function ChangesPage() {
  const { events: sseEvents, connected, newCount, clearNewCount, clearEvents } = useChangesSSE()
  const [history,   setHistory]   = useState([])
  const [loading,   setLoading]   = useState(true)
  const [deleting,  setDeleting]  = useState(false)
  const [filter,    setFilter]    = useState('ALL')

  const loadHistory = async () => {
    setLoading(true)
    try {
      const res = await api.getChanges({ limit: 300 })
      if (res.success && Array.isArray(res.data)) setHistory(res.data)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadHistory()
    clearNewCount()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Merge SSE events with history, deduplicate by event_id, sort newest-first
  const allEvents = (() => {
    const seen   = new Set(history.map((e) => e.event_id))
    const merged = [...history]
    for (const e of sseEvents) {
      if (!seen.has(e.event_id)) {
        merged.push(e)
        seen.add(e.event_id)
      }
    }
    return merged.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
  })()

  const filtered = filter === 'ALL' ? allEvents : allEvents.filter((e) => e.severity === filter)

  const counts = {
    CRITICAL: allEvents.filter((e) => e.severity === 'CRITICAL').length,
    HIGH:     allEvents.filter((e) => e.severity === 'HIGH').length,
    MEDIUM:   allEvents.filter((e) => e.severity === 'MEDIUM').length,
    LOW:      allEvents.filter((e) => e.severity === 'LOW').length,
  }

  const handleDelete = async () => {
    if (!confirm('Удалить все события изменений?')) return
    setDeleting(true)
    try {
      await api.deleteChanges()
      setHistory([])
      clearEvents()      // also wipe the in-memory store (live WS events)
    } finally {
      setDeleting(false)
    }
  }

  const toggleFilter = (f) => setFilter((cur) => (cur === f ? 'ALL' : f))

  return (
    <div>
      {/* ── Header ─────────────────────────────────────────────────────── */}
      <div className="page-header">
        <div>
          <h1 className="page-title">
            <Bell size={22} />
            Change Detection
            {newCount > 0 && (
              <span className="changes-new-badge">{newCount} new</span>
            )}
          </h1>
          <p className="page-subtitle">
            Мониторинг изменений сети в реальном времени — порты, устройства, сервисы
          </p>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div className={`sse-pill ${connected ? 'sse-on' : 'sse-off'}`}>
            {connected ? <Wifi size={12} /> : <WifiOff size={12} />}
            <span>{connected ? 'Live' : 'Reconnecting…'}</span>
          </div>
        </div>
      </div>

      {/* ── Severity stats ─────────────────────────────────────────────── */}
      <div className="changes-stats-row">
        <StatCard
          severity="CRITICAL" label="Критических" count={counts.CRITICAL}
          active={filter === 'CRITICAL'} onClick={() => toggleFilter('CRITICAL')}
        />
        <StatCard
          severity="HIGH" label="Высоких" count={counts.HIGH}
          active={filter === 'HIGH'} onClick={() => toggleFilter('HIGH')}
        />
        <StatCard
          severity="MEDIUM" label="Средних" count={counts.MEDIUM}
          active={filter === 'MEDIUM'} onClick={() => toggleFilter('MEDIUM')}
        />
        <StatCard
          severity="LOW" label="Низких" count={counts.LOW}
          active={filter === 'LOW'} onClick={() => toggleFilter('LOW')}
        />
      </div>

      {/* ── Toolbar ────────────────────────────────────────────────────── */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 16, alignItems: 'center' }}>
        <Button
          variant="ghost" size="sm"
          icon={<RefreshCw size={13} />}
          onClick={loadHistory}
          loading={loading}
        >
          Обновить
        </Button>
        <Button
          variant="danger" size="sm"
          icon={<Trash2 size={13} />}
          onClick={handleDelete}
          loading={deleting}
        >
          Очистить всё
        </Button>

        {filter !== 'ALL' && (
          <Button variant="ghost" size="sm" icon={<Filter size={13} />} onClick={() => setFilter('ALL')}>
            Снять фильтр: {filter}
          </Button>
        )}

        <span style={{ marginLeft: 'auto', fontSize: 12, color: 'var(--text-muted)' }}>
          {filtered.length} событий
        </span>
      </div>

      {/* ── Timeline ───────────────────────────────────────────────────── */}
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
          <Spinner size="lg" />
        </div>
      ) : filtered.length === 0 ? (
        <EmptyState
          title="Изменений не обнаружено"
          description={
            allEvents.length === 0
              ? 'Запустите сканирование минимум дважды для одной цели — система начнёт сравнивать результаты автоматически'
              : `Нет событий с фильтром «${filter}»`
          }
        />
      ) : (
        <div className="changes-timeline">
          {filtered.map((evt, idx) => (
            <ChangeCard key={evt.event_id ?? idx} event={evt} />
          ))}
        </div>
      )}
    </div>
  )
}

