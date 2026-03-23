import { useState, useCallback } from 'react'
import { Radio, Copy, Check, Plus, X } from 'lucide-react'
import { useStore }         from '@/store'
import { useSend }          from '@/hooks/useWebSocket'
import { Button, Badge, Card, EmptyState, ScanAnimation } from '@/components/ui'

const SERVICE = 'icmp_service'

export default function ICMPScanner() {
  const [targets,   setTargets]   = useState('192.168.1.1')
  const [pingCount, setPingCount] = useState(4)
  const [copied,    setCopied]    = useState(false)

  const send         = useSend()
  const activeScan   = useStore((s) => s.activeScan)
  const latestResult = useStore((s) => s.latestResult)

  const isScanning = activeScan?.scanner_service === SERVICE
  const result     = latestResult?.scanner_service === SERVICE ? latestResult : null
  const scanResult = result?.result

  const parsedTargets = targets
    .split(/[\n,]+/)
    .map((t) => t.trim())
    .filter(Boolean)

  const handleScan = useCallback(() => {
    if (!parsedTargets.length || isScanning) return
    send(SERVICE, { targets: parsedTargets, ping_count: Number(pingCount) })
  }, [parsedTargets, pingCount, isScanning, send])

  const handleKey = (e) => {
    if (e.ctrlKey && e.key === 'Enter') handleScan()
  }

  const copyJSON = () => {
    navigator.clipboard.writeText(JSON.stringify(scanResult, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  const lossColor = (pct) => {
    if (pct === 0)   return 'var(--green)'
    if (pct < 50)    return 'var(--yellow)'
    return 'var(--red)'
  }

  return (
    <div onKeyDown={handleKey}>
      <div className="page-header">
        <div>
          <h1 className="page-title"><Radio size={22} /> ICMP Ping</h1>
          <p className="page-subtitle">Ping multiple hosts simultaneously and measure packet loss</p>
        </div>
      </div>

      <div style={{ display: 'grid', gap: 20 }}>
        <Card title="Scan Configuration">
          <div className="grid-2">
            <div className="form-group">
              <label>Targets (comma or newline separated)</label>
              <textarea
                rows={4}
                value={targets}
                onChange={(e) => setTargets(e.target.value)}
                placeholder={'192.168.1.1\ngoogle.com\n8.8.8.8'}
              />
              <span className="text-muted text-xs" style={{ marginTop: 4 }}>
                {parsedTargets.length} target{parsedTargets.length !== 1 ? 's' : ''}
              </span>
            </div>
            <div className="form-group">
              <label>Ping Count</label>
              <input
                type="number"
                min={1}
                max={100}
                value={pingCount}
                onChange={(e) => setPingCount(e.target.value)}
              />
              <div className="text-muted text-xs" style={{ marginTop: 8 }}>
                Packets per target: <strong>{pingCount}</strong>
              </div>
            </div>
          </div>
          <div style={{ marginTop: 16, display: 'flex', gap: 10, alignItems: 'center' }}>
            <Button
              variant="primary" size="lg"
              loading={isScanning}
              onClick={handleScan}
              disabled={!parsedTargets.length}
            >
              {isScanning ? 'Pinging…' : `Ping ${parsedTargets.length} Host${parsedTargets.length !== 1 ? 's' : ''}`}
            </Button>
            <span className="text-muted text-sm">Ctrl + Enter</span>
          </div>
        </Card>

        {isScanning && (
          <Card>
            <ScanAnimation label={`Pinging ${parsedTargets.length} host${parsedTargets.length !== 1 ? 's' : ''}…`} />
          </Card>
        )}

        {scanResult && !isScanning && (
          <div className="result-panel animate-in">
            <div className="result-header">
              <span className="result-title"><Radio size={14} /> Ping Results</span>
              <div style={{ display: 'flex', gap: 8 }}>
                <Badge>{scanResult.status ?? 'completed'}</Badge>
                <button className="copy-btn" onClick={copyJSON}>
                  {copied ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            {scanResult.error ? (
              <div className="result-error">Error: {scanResult.error}</div>
            ) : scanResult.results?.length > 0 ? (
              <div className="table-wrap">
                <table>
                  <thead><tr>
                    <th>Target</th>
                    <th>Address</th>
                    <th style={{ textAlign: 'right' }}>Sent</th>
                    <th style={{ textAlign: 'right' }}>Received</th>
                    <th style={{ textAlign: 'right' }}>Loss</th>
                    <th>Status</th>
                  </tr></thead>
                  <tbody>
                    {scanResult.results.map((r, i) => (
                      <tr key={i}>
                        <td className="td-mono">{r.target}</td>
                        <td className="td-mono td-muted">{r.address || '—'}</td>
                        <td style={{ textAlign: 'right', fontFamily: 'var(--font-mono)' }}>{r.packets_sent}</td>
                        <td style={{ textAlign: 'right', fontFamily: 'var(--font-mono)' }}>{r.packets_received}</td>
                        <td style={{ textAlign: 'right', fontFamily: 'var(--font-mono)', color: lossColor(r.packet_loss_percent) }}>
                          {r.packet_loss_percent?.toFixed(1)}%
                        </td>
                        <td>
                          <Badge>{r.error ? 'error' : r.packet_loss_percent === 0 ? 'online' : r.packet_loss_percent >= 100 ? 'offline' : 'partial'}</Badge>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <EmptyState title="No results" />
            )}
          </div>
        )}
      </div>
    </div>
  )
}

