import { Link } from 'react-router-dom'
import { Network, Radio, Shield, Terminal, Clock, Activity, Zap } from 'lucide-react'
import { useStore } from '@/store'
import { StatusDot, Badge } from '@/components/ui'
import { formatDistanceToNow } from 'date-fns'

const SCANNER_TILES = [
  { to: '/arp',  label: 'ARP Scanner',  Icon: Network,  color: 'green',  desc: 'Discover LAN devices via ARP' },
  { to: '/icmp', label: 'ICMP Ping',    Icon: Radio,    color: 'blue',   desc: 'Ping hosts and measure loss'  },
  { to: '/nmap', label: 'Nmap Scanner', Icon: Shield,   color: 'purple', desc: 'Port scan, OS detection'      },
  { to: '/tcp',  label: 'TCP Banner',   Icon: Terminal, color: 'yellow', desc: 'Grab service banners'         },
]

const svcLabel = {
  arp_service:  'ARP',
  icmp_service: 'ICMP',
  nmap_service: 'Nmap',
  tcp_service:  'TCP',
}

export default function Dashboard() {
  const wsStatus     = useStore((s) => s.wsStatus)
  const activeScan   = useStore((s) => s.activeScan)
  const recentResults = useStore((s) => s.recentResults)

  return (
    <div>
      <div className="page-header">
        <div>
          <h1 className="page-title"><Activity size={22} /> Dashboard</h1>
          <p className="page-subtitle">Network scanning control center</p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <StatusDot status={wsStatus} />
          <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
            {wsStatus === 'connected' ? 'Backend ready' : 'Reconnecting…'}
          </span>
        </div>
      </div>

      {activeScan && (
        <div className="active-scan-banner animate-in">
          <Zap size={16} />
          <span>
            Scanning with <strong>{svcLabel[activeScan.scanner_service] ?? activeScan.scanner_service}</strong>
            &nbsp;— started {Math.round((Date.now() - activeScan.startedAt) / 1000)}s ago
          </span>
          <span className="banner-spinner" />
        </div>
      )}

      <div style={{ marginBottom: 28 }}>
        <div className="section-label">Quick Launch</div>
        <div className="scanner-tiles">
          {SCANNER_TILES.map(({ to, label, Icon, color, desc }) => (
            <Link key={to} to={to} className="scanner-tile">
              <div className={`stat-icon ${color}`}>
                <Icon size={20} />
              </div>
              <div>
                <div style={{ fontWeight: 600, fontSize: 14 }}>{label}</div>
                <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 2 }}>{desc}</div>
              </div>
            </Link>
          ))}
        </div>
      </div>

      <div>
        <div className="section-label" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>Recent Scans</span>
          <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>{recentResults.length} in memory</span>
        </div>

        {recentResults.length === 0 ? (
          <div className="card" style={{ padding: 32, textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
            <Clock size={32} style={{ opacity: 0.3, display: 'block', margin: '0 auto 12px' }} />
            No scans yet — run a scanner to see results here
          </div>
        ) : (
          <div className="card" style={{ padding: 0 }}>
            {recentResults.slice(0, 10).map((entry, i) => {
              const r      = entry.result ?? {}
              const isErr  = !!r.error
              const label  = svcLabel[entry.scanner_service] ?? entry.scanner_service
              const status = isErr ? 'error' : (r.status ?? 'completed')
              const meta   = buildMeta(entry)

              return (
                <div key={i} className="recent-result-item" style={{ padding: '12px 20px', borderBottom: i < recentResults.slice(0, 10).length - 1 ? '1px solid var(--border)' : 'none' }}>
                  <Badge color={isErr ? 'red' : 'green'} dot={false}>{label}</Badge>
                  <span className="td-mono" style={{ flex: 1, fontSize: 13 }}>{meta}</span>
                  <Badge>{status}</Badge>
                  <span className="text-muted text-xs">
                    {formatDistanceToNow(entry.receivedAt, { addSuffix: true })}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

function buildMeta(entry) {
  const r   = entry.result ?? {}
  const svc = entry.scanner_service
  if (svc === 'arp_service')  return `${entry.options?.ip_range ?? '—'} — ${r.online_count ?? 0}/${r.total_count ?? 0} online`
  if (svc === 'icmp_service') return `${(entry.options?.targets ?? []).join(', ')}`
  if (svc === 'nmap_service') return `${r.host ?? entry.options?.ip ?? '—'} [${entry.options?.scan_method ?? '—'}]`
  if (svc === 'tcp_service')  return `${r.host ?? entry.options?.host ?? '—'}:${entry.options?.port ?? '—'}`
  return '—'
}

