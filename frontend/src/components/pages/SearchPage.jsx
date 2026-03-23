import { useState } from 'react'
import { Search, Network, Radio, Shield, Terminal } from 'lucide-react'
import { api }           from '@/api/http'
import { Button, Badge, EmptyState, Card, Spinner } from '@/components/ui'
import toast             from 'react-hot-toast'

const SCANNER_TYPES = [
  { id: 'arp',  label: 'ARP',  Icon: Network,  color: 'green'  },
  { id: 'icmp', label: 'ICMP', Icon: Radio,    color: 'blue'   },
  { id: 'nmap', label: 'Nmap', Icon: Shield,   color: 'purple' },
  { id: 'tcp',  label: 'TCP',  Icon: Terminal, color: 'yellow' },
]

export default function SearchPage() {
  const [type,    setType]    = useState('arp')
  const [results, setResults] = useState(null)
  const [loading, setLoading] = useState(false)

  const selected = SCANNER_TYPES.find((t) => t.id === type)

  const handleSearch = async (body) => {
    setLoading(true)
    setResults(null)
    try {
      const res = await api.search(type, body)
      if (!res.success) {
        toast.error(res.error ?? 'Search failed')
      } else {
        setResults(res)
      }
    } catch {
      toast.error('Search failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <h1 className="page-title"><Search size={22} /> Search</h1>
          <p className="page-subtitle">Query historical scan data by target parameters</p>
        </div>
      </div>

      <Card title="Scanner Type" style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {SCANNER_TYPES.map(({ id, label, Icon, color }) => (
            <button
              key={id}
              onClick={() => { setType(id); setResults(null) }}
              className={`scanner-type-btn${type === id ? ` active color-${color}` : ''}`}
            >
              <Icon size={14} />
              {label}
            </button>
          ))}
        </div>
      </Card>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1.5fr', gap: 20, alignItems: 'start' }}>
        <Card title={`Search ${selected?.label}`}>
          {type === 'arp'  && <ARPForm  onSearch={handleSearch} loading={loading} />}
          {type === 'icmp' && <ICMPForm onSearch={handleSearch} loading={loading} />}
          {type === 'nmap' && <NmapForm onSearch={handleSearch} loading={loading} />}
          {type === 'tcp'  && <TCPForm  onSearch={handleSearch} loading={loading} />}
        </Card>

        <Card title="Results">
          {loading && <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}><Spinner size="lg" /></div>}
          {!loading && results === null && <EmptyState title="Run a search" description="Fill in the form and press Search" />}
          {!loading && results !== null && !results.found && <EmptyState title="No results found" />}
          {!loading && results?.found && (
            <SearchResults type={type} data={results.data} count={results.count} />
          )}
        </Card>
      </div>
    </div>
  )
}

function ARPForm({ onSearch, loading }) {
  const [iface,   setIface]   = useState('')
  const [ipRange, setIpRange] = useState('')
  return (
    <div style={{ display: 'grid', gap: 12 }}>
      <div className="form-group">
        <label>Interface Name</label>
        <input value={iface} onChange={(e) => setIface(e.target.value)} placeholder="eth0" />
      </div>
      <div className="form-group">
        <label>IP Range (CIDR)</label>
        <input value={ipRange} onChange={(e) => setIpRange(e.target.value)} placeholder="192.168.1.0/24" />
      </div>
      <Button variant="primary" loading={loading} onClick={() => onSearch({ interface_name: iface, ip_range: ipRange })} disabled={!iface && !ipRange}>
        Search
      </Button>
    </div>
  )
}

function ICMPForm({ onSearch, loading }) {
  const [targets, setTargets] = useState('')
  const parsed = targets.split(/[\n,]+/).map((t) => t.trim()).filter(Boolean)
  return (
    <div style={{ display: 'grid', gap: 12 }}>
      <div className="form-group">
        <label>Targets (comma or newline)</label>
        <textarea rows={3} value={targets} onChange={(e) => setTargets(e.target.value)} placeholder="192.168.1.1, google.com" />
      </div>
      <Button variant="primary" loading={loading} onClick={() => onSearch({ targets: parsed })} disabled={!parsed.length}>
        Search
      </Button>
    </div>
  )
}

function NmapForm({ onSearch, loading }) {
  const [ip,     setIp]     = useState('')
  const [method, setMethod] = useState('tcp_udp_scan')
  return (
    <div style={{ display: 'grid', gap: 12 }}>
      <div className="form-group">
        <label>IP Address</label>
        <input value={ip} onChange={(e) => setIp(e.target.value)} placeholder="192.168.1.1" />
      </div>
      <div className="form-group">
        <label>Scan Method</label>
        <select value={method} onChange={(e) => setMethod(e.target.value)}>
          <option value="tcp_udp_scan">TCP / UDP Scan</option>
          <option value="os_detection">OS Detection</option>
          <option value="host_discovery">Host Discovery</option>
        </select>
      </div>
      <Button variant="primary" loading={loading} onClick={() => onSearch({ ip, scan_method: method })} disabled={!ip}>
        Search
      </Button>
    </div>
  )
}

function TCPForm({ onSearch, loading }) {
  const [host, setHost] = useState('')
  const [port, setPort] = useState('')
  return (
    <div style={{ display: 'grid', gap: 12 }}>
      <div className="form-group">
        <label>Host</label>
        <input value={host} onChange={(e) => setHost(e.target.value)} placeholder="192.168.1.1" />
      </div>
      <div className="form-group">
        <label>Port</label>
        <input value={port} onChange={(e) => setPort(e.target.value)} placeholder="80" />
      </div>
      <Button variant="primary" loading={loading} onClick={() => onSearch({ host, port })} disabled={!host && !port}>
        Search
      </Button>
    </div>
  )
}

function SearchResults({ type, data, count }) {
  const rows = Array.isArray(data) ? data : []
  return (
    <div>
      <div style={{ marginBottom: 12, fontSize: 12, color: 'var(--text-muted)' }}>
        Found <strong style={{ color: 'var(--text-primary)' }}>{count}</strong> record{count !== 1 ? 's' : ''}
      </div>
      {type === 'arp'  && <ARPResults  rows={rows} />}
      {type === 'icmp' && <ICMPResults rows={rows} />}
      {type === 'nmap' && <NmapResults rows={rows} />}
      {type === 'tcp'  && <TCPResults  rows={rows} />}
    </div>
  )
}

function ARPResults({ rows }) {
  return rows.map((r, i) => (
    <div key={i} style={{ marginBottom: 16 }}>
      <div style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
        <Badge>{r.status}</Badge>
        <span className="td-mono text-sm">{r.ip_range} via {r.interface_name}</span>
        <span className="text-muted text-sm">{r.online_count}/{r.total_count} online</span>
      </div>
      <div className="table-wrap">
        <table>
          <thead><tr><th>IP</th><th>MAC</th><th>Status</th></tr></thead>
          <tbody>
            {(r.devices ?? []).slice(0, 10).map((d, j) => (
              <tr key={j}><td className="td-mono">{d.ip}</td><td className="td-mono">{d.mac}</td><td><Badge>{d.status}</Badge></td></tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  ))
}

function ICMPResults({ rows }) {
  return rows.map((r, i) => (
    <div key={i} style={{ marginBottom: 16 }}>
      <div style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
        <Badge>{r.status}</Badge>
        <span className="td-mono text-sm">{(r.targets ?? []).join(', ')}</span>
      </div>
      <div className="table-wrap">
        <table>
          <thead><tr><th>Target</th><th>Sent</th><th>Recv</th><th>Loss</th></tr></thead>
          <tbody>
            {(r.results ?? []).map((res, j) => (
              <tr key={j}>
                <td className="td-mono">{res.target}</td>
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
  ))
}

function NmapResults({ rows }) {
  return rows.map((r, i) => (
    <div key={i} style={{ marginBottom: 12 }}>
      <div style={{ display: 'flex', gap: 8, marginBottom: 6 }}>
        <Badge>{r.status}</Badge>
        <span className="td-mono text-sm">{r.ip || r.host}</span>
        {r.name && <span className="text-muted text-sm">{r.name}</span>}
      </div>
    </div>
  ))
}

function TCPResults({ rows }) {
  return rows.map((r, i) => (
    <div key={i} style={{ marginBottom: 16 }}>
      <div style={{ display: 'flex', gap: 8, marginBottom: 6 }}>
        <Badge>{r.status}</Badge>
        <span className="td-mono text-sm">{r.host}:{r.port}</span>
      </div>
      {r.decoded_text && <pre className="decoded-text" style={{ maxHeight: 150 }}>{r.decoded_text}</pre>}
    </div>
  ))
}

