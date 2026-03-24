import { useState, useCallback } from 'react'
import { Network, Copy, Check } from 'lucide-react'
import { useStore }         from '@/store'
import { useSend }          from '@/hooks/useWebSocket'
import { Button, Badge, Card, EmptyState, ScanAnimation } from '@/components/ui'

const SERVICE = 'arp_service'

export default function ARPScanner() {
  const [iface,   setIface]   = useState('eth0')
  const [ipRange, setIpRange] = useState('192.168.1.0/24')
  const [copied,  setCopied]  = useState(false)

  const send         = useSend()
  const activeScan   = useStore((s) => s.activeScan)
  const latestResult = useStore((s) => s.latestResult)
  const wsStatus     = useStore((s) => s.wsStatus)

  const isScanning  = activeScan?.scanner_service === SERVICE
  const isConnected = wsStatus === 'connected'
  const result     = latestResult?.scanner_service === SERVICE ? latestResult : null
  const scanResult = result?.result

  const handleScan = useCallback(() => {
    if (!iface || !ipRange || isScanning) return
    send(SERVICE, { interface_name: iface, ip_range: ipRange })
  }, [iface, ipRange, isScanning, send])

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
          <h1 className="page-title"><Network size={22} /> ARP Scanner</h1>
          <p className="page-subtitle">Discover devices on a local network via ARP requests</p>
        </div>
      </div>

      <div style={{ display: 'grid', gap: 20 }}>
        <Card title="Scan Configuration">
          <div className="grid-2">
            <div className="form-group">
              <label>Network Interface</label>
              <input value={iface} onChange={(e) => setIface(e.target.value)} placeholder="eth0, wlan0, ens3…" />
            </div>
            <div className="form-group">
              <label>IP Range (CIDR)</label>
              <input value={ipRange} onChange={(e) => setIpRange(e.target.value)} placeholder="192.168.1.0/24" />
            </div>
          </div>
          <div style={{ marginTop: 16, display: 'flex', gap: 10, alignItems: 'center' }}>
            <Button variant="primary" size="lg" loading={isScanning} onClick={handleScan} disabled={!iface || !ipRange || !isConnected}>
              {isScanning ? 'Scanning…' : 'Start ARP Scan'}
            </Button>
            <span className="text-muted text-sm">{isConnected ? 'Ctrl + Enter' : 'Waiting for backend…'}</span>
          </div>
        </Card>

        {isScanning && (
          <Card>
            <ScanAnimation label={`Scanning ${ipRange} via ${iface}…`} />
          </Card>
        )}

        {scanResult && !isScanning && (
          <div className="result-panel animate-in">
            <div className="result-header">
              <span className="result-title"><Network size={14} /> ARP Scan Result</span>
              <div style={{ display: 'flex', gap: 8 }}>
                <Badge>{scanResult.status ?? 'completed'}</Badge>
                <button className="copy-btn" onClick={copyJSON} title="Copy JSON">
                  {copied ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            {scanResult.error ? (
              <div className="result-error">Error: {scanResult.error}</div>
            ) : (
              <>
                <div className="result-stats">
                  <div className="result-stat">
                    <div className="result-stat-value">{scanResult.total_count ?? 0}</div>
                    <div className="result-stat-label">Total</div>
                  </div>
                  <div className="result-stat">
                    <div className="result-stat-value" style={{ color: 'var(--green)' }}>{scanResult.online_count ?? 0}</div>
                    <div className="result-stat-label">Online</div>
                  </div>
                  <div className="result-stat">
                    <div className="result-stat-value" style={{ color: 'var(--red)' }}>{scanResult.offline_count ?? 0}</div>
                    <div className="result-stat-label">Offline</div>
                  </div>
                  <div className="result-stat">
                    <div className="result-stat-value" style={{ color: 'var(--text-muted)', fontSize: 13 }}>{ipRange}</div>
                    <div className="result-stat-label">Range</div>
                  </div>
                </div>

                {scanResult.devices?.length > 0 ? (
                  <div className="table-wrap">
                    <table>
                      <thead><tr>
                        <th>IP Address</th>
                        <th>MAC Address</th>
                        <th>Vendor</th>
                        <th>Status</th>
                      </tr></thead>
                      <tbody>
                        {scanResult.devices.map((d, i) => (
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
                ) : (
                  <EmptyState title="No devices found" description="No devices responded to ARP requests" />
                )}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

