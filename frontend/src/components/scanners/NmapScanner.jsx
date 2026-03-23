import { useState, useCallback } from 'react'
import { Shield, Copy, Check } from 'lucide-react'
import { useStore }         from '@/store'
import { useSend }          from '@/hooks/useWebSocket'
import { Button, Badge, Card, EmptyState, ScanAnimation } from '@/components/ui'

const SERVICE = 'nmap_service'

const METHODS = [
  { id: 'tcp_udp_scan',   label: 'TCP / UDP Scan'   },
  { id: 'os_detection',   label: 'OS Detection'     },
  { id: 'host_discovery', label: 'Host Discovery'   },
]

const SCANNER_TYPES = ['SYN', 'TCP', 'UDP', 'ACK', 'Window', 'Maimon']

export default function NmapScanner() {
  const [method,      setMethod]      = useState('tcp_udp_scan')
  const [ip,          setIp]          = useState('')
  const [ports,       setPorts]       = useState('1-1024')
  const [scannerType, setScannerType] = useState('SYN')
  const [copied,      setCopied]      = useState(false)

  const send         = useSend()
  const activeScan   = useStore((s) => s.activeScan)
  const latestResult = useStore((s) => s.latestResult)

  const isScanning = activeScan?.scanner_service === SERVICE
  const result     = latestResult?.scanner_service === SERVICE ? latestResult : null
  const scanResult = result?.result
  const usedMethod = result?.options?.scan_method

  const handleScan = useCallback(() => {
    if (!ip || isScanning) return
    send(SERVICE, { scan_method: method, ip, ports, scanner_type: scannerType })
  }, [method, ip, ports, scannerType, isScanning, send])

  const handleKey = (e) => {
    if (e.ctrlKey && e.key === 'Enter') handleScan()
  }

  const copyJSON = () => {
    navigator.clipboard.writeText(JSON.stringify(scanResult, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div onKeyDown={handleKey}>
      <div className="page-header">
        <div>
          <h1 className="page-title"><Shield size={22} /> Nmap Scanner</h1>
          <p className="page-subtitle">Port scanning, OS detection and host discovery powered by Nmap</p>
        </div>
      </div>

      <div style={{ display: 'grid', gap: 20 }}>
        <Card title="Scan Configuration">
          <div className="tabs" style={{ marginBottom: 20 }}>
            {METHODS.map((m) => (
              <button key={m.id} className={`tab${method === m.id ? ' active' : ''}`} onClick={() => setMethod(m.id)}>
                {m.label}
              </button>
            ))}
          </div>

          <div className="grid-2">
            <div className="form-group">
              <label>Target IP / CIDR</label>
              <input value={ip} onChange={(e) => setIp(e.target.value)} placeholder="192.168.1.1 or 192.168.1.0/24" />
            </div>

            {method === 'tcp_udp_scan' && (
              <div className="form-group">
                <label>Port Range</label>
                <input value={ports} onChange={(e) => setPorts(e.target.value)} placeholder="1-1024, 80,443,8080" />
              </div>
            )}

            {method === 'tcp_udp_scan' && (
              <div className="form-group">
                <label>Scan Type</label>
                <select value={scannerType} onChange={(e) => setScannerType(e.target.value)}>
                  {SCANNER_TYPES.map((t) => <option key={t}>{t}</option>)}
                </select>
              </div>
            )}
          </div>

          <div style={{ marginTop: 16, display: 'flex', gap: 10, alignItems: 'center' }}>
            <Button variant="primary" size="lg" loading={isScanning} onClick={handleScan} disabled={!ip}>
              {isScanning ? 'Scanning…' : `Run ${METHODS.find(m => m.id === method)?.label}`}
            </Button>
            <span className="text-muted text-sm">Ctrl + Enter</span>
          </div>
        </Card>

        {isScanning && (
          <Card>
            <ScanAnimation label={`Running ${METHODS.find(m => m.id === method)?.label} on ${ip}…`} />
          </Card>
        )}

        {scanResult && !isScanning && (
          <div className="result-panel animate-in">
            <div className="result-header">
              <span className="result-title"><Shield size={14} /> Nmap Result</span>
              <div style={{ display: 'flex', gap: 8 }}>
                <Badge>{scanResult.status ?? 'completed'}</Badge>
                <button className="copy-btn" onClick={copyJSON}>
                  {copied ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            {scanResult.error ? (
              <div className="result-error">Error: {scanResult.error}</div>
            ) : (
              <>
                {usedMethod === 'tcp_udp_scan' && <TcpUdpResult r={scanResult} />}
                {usedMethod === 'os_detection' && <OsDetectionResult r={scanResult} />}
                {usedMethod === 'host_discovery' && <HostDiscoveryResult r={scanResult} />}
                {!usedMethod && <TcpUdpResult r={scanResult} />}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function TcpUdpResult({ r }) {
  if (!r.port_info?.length) return <EmptyState title="No ports found" />
  const ports = []
  r.port_info.forEach((info) => {
    info.all_ports?.forEach((port, i) => {
      ports.push({
        port,
        protocol:    info.protocols?.[i]    ?? '—',
        state:       info.state?.[i]        ?? '—',
        serviceName: info.service_name?.[i] ?? '—',
      })
    })
  })

  return (
    <>
      <div className="result-stats">
        <div className="result-stat">
          <div className="result-stat-value">{r.host || '—'}</div>
          <div className="result-stat-label">Host</div>
        </div>
        <div className="result-stat">
          <div className="result-stat-value">{ports.length}</div>
          <div className="result-stat-label">Ports</div>
        </div>
        <div className="result-stat">
          <div className="result-stat-value" style={{ color: 'var(--green)' }}>
            {ports.filter(p => p.state === 'open').length}
          </div>
          <div className="result-stat-label">Open</div>
        </div>
      </div>
      <div className="table-wrap">
        <table>
          <thead><tr>
            <th>Port</th>
            <th>Protocol</th>
            <th>State</th>
            <th>Service</th>
          </tr></thead>
          <tbody>
            {ports.map((p, i) => (
              <tr key={i}>
                <td className="td-mono" style={{ color: p.state === 'open' ? 'var(--green)' : undefined }}>{p.port}</td>
                <td className="td-mono td-muted">{p.protocol}</td>
                <td><Badge>{p.state}</Badge></td>
                <td className="td-muted">{p.serviceName}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}

function OsDetectionResult({ r }) {
  const accuracy = r.accuracy ?? 0
  return (
    <div className="os-result-grid">
      {[
        { k: 'Host',     v: r.host     || '—' },
        { k: 'OS Name',  v: r.name     || '—' },
        { k: 'Vendor',   v: r.vendor   || '—' },
        { k: 'Family',   v: r.family   || '—' },
        { k: 'Type',     v: r.type     || '—' },
        { k: 'Accuracy', v: `${accuracy}%`    },
      ].map(({ k, v }) => (
        <div key={k} className="os-result-item">
          <div className="os-result-key">{k}</div>
          <div className="os-result-value">{v}</div>
          {k === 'Accuracy' && (
            <div className="accuracy-bar">
              <div className="accuracy-fill" style={{ width: `${accuracy}%` }} />
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

function HostDiscoveryResult({ r }) {
  return (
    <div className="os-result-grid">
      {[
        { k: 'Host',        v: r.host       || '—' },
        { k: 'DNS',         v: r.dns        || '—' },
        { k: 'Hosts Up',    v: r.host_up ?? 0 },
        { k: 'Total Hosts', v: r.host_total ?? 0 },
        { k: 'Reason',      v: r.reason     || '—' },
        { k: 'Status',      v: r.status     || '—' },
      ].map(({ k, v }) => (
        <div key={k} className="os-result-item">
          <div className="os-result-key">{k}</div>
          <div className="os-result-value">{String(v)}</div>
        </div>
      ))}
    </div>
  )
}

