import { useState, useEffect } from 'react'
import { Clock, Trash2, RefreshCw, ChevronDown, ChevronRight, Network, Radio, Shield, Terminal } from 'lucide-react'
import { useHistory }        from '@/hooks/useHistory'
import { Button, Badge, EmptyState, Spinner } from '@/components/ui'
import { formatDistanceToNow, format } from 'date-fns'

const TABS = [
  { id: 'arp',  label: 'ARP',  Icon: Network  },
  { id: 'icmp', label: 'ICMP', Icon: Radio    },
  { id: 'nmap', label: 'Nmap', Icon: Shield   },
  { id: 'tcp',  label: 'TCP',  Icon: Terminal },
]

export default function HistoryPage() {
  const [tab, setTab] = useState('arp')

  return (
    <div>
      <div className="page-header">
        <div>
          <h1 className="page-title"><Clock size={22} /> History</h1>
          <p className="page-subtitle">View and manage saved scan results</p>
        </div>
      </div>

      <div className="tabs">
        {TABS.map(({ id, label, Icon }) => (
          <button key={id} className={`tab${tab === id ? ' active' : ''}`} onClick={() => setTab(id)}>
            <Icon size={13} style={{ marginRight: 6, verticalAlign: 'middle' }} />
            {label}
          </button>
        ))}
      </div>

      {tab === 'arp'  && <ARPHistory />}
      {tab === 'icmp' && <ICMPHistory />}
      {tab === 'nmap' && <NmapHistory />}
      {tab === 'tcp'  && <TCPHistory />}
    </div>
  )
}

function HistoryShell({ type, title, children, limit = 50 }) {
  const { records, loading, load, clear } = useHistory(type)

  useEffect(() => { load({ limit }) }, [])

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginBottom: 16 }}>
        <Button variant="ghost" size="sm" icon={<RefreshCw size={13} />} onClick={() => load({ limit })} loading={loading}>
          Refresh
        </Button>
        <Button variant="danger" size="sm" icon={<Trash2 size={13} />} onClick={clear}>
          Clear All
        </Button>
      </div>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
          <Spinner size="lg" />
        </div>
      ) : children(records)}
    </div>
  )
}

function ARPHistory() {
  return (
    <HistoryShell type="arp">
      {(records) => !records?.length
        ? <EmptyState title="No ARP history" description="Run an ARP scan to populate history" />
        : <div className="card" style={{ padding: 0 }}>
            {records.map((r) => <ARPRow key={r.id} r={r} />)}
          </div>
      }
    </HistoryShell>
  )
}

function ARPRow({ r }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="history-item">
      <div className="history-item-header" onClick={() => setOpen(!open)}>
        <div className="history-item-title">
          {open ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          <Network size={14} />
          {r.ip_range} via {r.interface_name}
        </div>
        <div className="history-item-meta">
          <Badge>{r.status}</Badge>
          <span>{r.online_count}/{r.total_count} online</span>
          <span>{r.created_at ? formatDistanceToNow(new Date(r.created_at), { addSuffix: true }) : '—'}</span>
        </div>
      </div>
      {open && (
        <div className="history-item-body">
          <div className="table-wrap">
            <table>
              <thead><tr><th>IP</th><th>MAC</th><th>Vendor</th><th>Status</th></tr></thead>
              <tbody>
                {(r.devices ?? []).map((d, i) => (
                  <tr key={i}>
                    <td className="td-mono">{d.ip}</td>
                    <td className="td-mono">{d.mac || '—'}</td>
                    <td className="td-muted">{d.vendor || '—'}</td>
                    <td><Badge>{d.status}</Badge></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

function ICMPHistory() {
  return (
    <HistoryShell type="icmp">
      {(records) => !records?.length
        ? <EmptyState title="No ICMP history" description="Run a ping scan to populate history" />
        : <div className="card" style={{ padding: 0 }}>
            {records.map((r) => <ICMPRow key={r.id} r={r} />)}
          </div>
      }
    </HistoryShell>
  )
}

function ICMPRow({ r }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="history-item">
      <div className="history-item-header" onClick={() => setOpen(!open)}>
        <div className="history-item-title">
          {open ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          <Radio size={14} />
          {(r.targets ?? []).join(', ')}
        </div>
        <div className="history-item-meta">
          <Badge>{r.status}</Badge>
          <span>{r.results?.length ?? 0} targets</span>
          <span>{r.created_at ? formatDistanceToNow(new Date(r.created_at), { addSuffix: true }) : '—'}</span>
        </div>
      </div>
      {open && (
        <div className="history-item-body">
          <div className="table-wrap">
            <table>
              <thead><tr><th>Target</th><th>Address</th><th>Sent</th><th>Received</th><th>Loss</th></tr></thead>
              <tbody>
                {(r.results ?? []).map((res, i) => (
                  <tr key={i}>
                    <td className="td-mono">{res.target}</td>
                    <td className="td-mono td-muted">{res.address || '—'}</td>
                    <td className="td-mono" style={{ textAlign: 'right' }}>{res.packets_sent}</td>
                    <td className="td-mono" style={{ textAlign: 'right' }}>{res.packets_received}</td>
                    <td className="td-mono" style={{ textAlign: 'right', color: res.packet_loss_percent > 0 ? 'var(--red)' : 'var(--green)' }}>
                      {res.packet_loss_percent?.toFixed(1)}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

function NmapHistory() {
  const [subTab, setSubTab] = useState('tcp_udp')

  const SUB_TABS = [
    { id: 'tcp_udp',        label: 'TCP / UDP' },
    { id: 'os_detection',   label: 'OS Detection' },
    { id: 'host_discovery', label: 'Host Discovery' },
  ]

  return (
    <HistoryShell type="nmap">
      {(records) => {
        const data = records ?? {}
        return (
          <div>
            <div className="tabs" style={{ marginBottom: 16 }}>
              {SUB_TABS.map(({ id, label }) => (
                <button key={id} className={`tab${subTab === id ? ' active' : ''}`} onClick={() => setSubTab(id)}>
                  {label}
                </button>
              ))}
            </div>
            {subTab === 'tcp_udp'        && <NmapTcpUdpList rows={data.tcp_udp ?? []} />}
            {subTab === 'os_detection'   && <NmapOsList     rows={data.os_detection ?? []} />}
            {subTab === 'host_discovery' && <NmapHostList   rows={data.host_discovery ?? []} />}
          </div>
        )
      }}
    </HistoryShell>
  )
}

function NmapTcpUdpList({ rows }) {
  if (!rows.length) return <EmptyState title="No TCP/UDP scan history" />
  return (
    <div className="card" style={{ padding: 0 }}>
      {rows.map((r) => {
        const ports = []
        ;(r.port_info ?? []).forEach((info) => {
          info.all_ports?.forEach((p, i) => ports.push({ p, proto: info.protocols?.[i], state: info.state?.[i], svc: info.service_name?.[i] }))
        })
        return (
          <NmapExpandRow key={r.id} title={`${r.ip} [${r.scanner_type}]`} badge={r.status} meta={`${ports.filter(p => p.state === 'open').length} open ports`} time={r.created_at} icon={<Shield size={14} />}>
            <div className="table-wrap">
              <table>
                <thead><tr><th>Port</th><th>Protocol</th><th>State</th><th>Service</th></tr></thead>
                <tbody>
                  {ports.map((p, i) => (
                    <tr key={i}>
                      <td className="td-mono" style={{ color: p.state === 'open' ? 'var(--green)' : undefined }}>{p.p}</td>
                      <td className="td-mono td-muted">{p.proto}</td>
                      <td><Badge>{p.state}</Badge></td>
                      <td className="td-muted">{p.svc}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </NmapExpandRow>
        )
      })}
    </div>
  )
}

function NmapOsList({ rows }) {
  if (!rows.length) return <EmptyState title="No OS detection history" />
  return (
    <div className="card" style={{ padding: 0 }}>
      {rows.map((r) => (
        <NmapExpandRow key={r.id} title={r.ip} badge={r.status} meta={`${r.name || '—'} (${r.accuracy}%)`} time={r.created_at} icon={<Shield size={14} />}>
          <div className="os-result-grid">
            {[['Host', r.host], ['OS', r.name], ['Vendor', r.vendor], ['Family', r.family], ['Type', r.type], ['Accuracy', `${r.accuracy}%`]].map(([k, v]) => (
              <div key={k} className="os-result-item">
                <div className="os-result-key">{k}</div>
                <div className="os-result-value">{v || '—'}</div>
              </div>
            ))}
          </div>
        </NmapExpandRow>
      ))}
    </div>
  )
}

function NmapHostList({ rows }) {
  if (!rows.length) return <EmptyState title="No host discovery history" />
  return (
    <div className="card" style={{ padding: 0 }}>
      {rows.map((r) => (
        <NmapExpandRow key={r.id} title={r.ip} badge={r.status} meta={`${r.host_up}/${r.host_total} up`} time={r.created_at} icon={<Shield size={14} />}>
          <div className="os-result-grid">
            {[['Host', r.host], ['DNS', r.dns], ['Hosts Up', r.host_up], ['Total', r.host_total], ['Reason', r.reason], ['Status', r.status]].map(([k, v]) => (
              <div key={k} className="os-result-item">
                <div className="os-result-key">{k}</div>
                <div className="os-result-value">{v ?? '—'}</div>
              </div>
            ))}
          </div>
        </NmapExpandRow>
      ))}
    </div>
  )
}

function NmapExpandRow({ title, badge, meta, time, icon, children }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="history-item">
      <div className="history-item-header" onClick={() => setOpen(!open)}>
        <div className="history-item-title">
          {open ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          {icon}{title}
        </div>
        <div className="history-item-meta">
          <Badge>{badge}</Badge>
          <span>{meta}</span>
          <span>{time ? formatDistanceToNow(new Date(time), { addSuffix: true }) : '—'}</span>
        </div>
      </div>
      {open && <div className="history-item-body">{children}</div>}
    </div>
  )
}

function TCPHistory() {
  return (
    <HistoryShell type="tcp">
      {(records) => !records?.length
        ? <EmptyState title="No TCP history" description="Run a TCP banner grab to populate history" />
        : <div className="card" style={{ padding: 0 }}>
            {records.map((r) => <TCPRow key={r.id} r={r} />)}
          </div>
      }
    </HistoryShell>
  )
}

function TCPRow({ r }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="history-item">
      <div className="history-item-header" onClick={() => setOpen(!open)}>
        <div className="history-item-title">
          {open ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          <Terminal size={14} />
          {r.host}:{r.port}
        </div>
        <div className="history-item-meta">
          <Badge>{r.status}</Badge>
          <span>{r.created_at ? formatDistanceToNow(new Date(r.created_at), { addSuffix: true }) : '—'}</span>
        </div>
      </div>
      {open && (
        <div className="history-item-body">
          {r.decoded_text
            ? <pre className="decoded-text">{r.decoded_text}</pre>
            : <span className="text-muted">No decoded data</span>
          }
          {r.hex_object_key && (
            <div style={{ marginTop: 8, fontSize: 11, color: 'var(--text-muted)' }}>
              Key: <code>{r.hex_object_key}</code>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

