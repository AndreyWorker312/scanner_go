import { useState, useCallback } from 'react'
import { Terminal, Copy, Check } from 'lucide-react'
import { useStore }         from '@/store'
import { useSend }          from '@/hooks/useWebSocket'
import { Button, Badge, Card, EmptyState, ScanAnimation } from '@/components/ui'

const SERVICE = 'tcp_service'

export default function TCPScanner() {
  const [host,   setHost]   = useState('')
  const [port,   setPort]   = useState('80')
  const [copied, setCopied] = useState(false)

  const send         = useSend()
  const activeScan   = useStore((s) => s.activeScan)
  const latestResult = useStore((s) => s.latestResult)

  const isScanning = activeScan?.scanner_service === SERVICE
  const result     = latestResult?.scanner_service === SERVICE ? latestResult : null
  const scanResult = result?.result

  const handleScan = useCallback(() => {
    if (!host || !port || isScanning) return
    send(SERVICE, { host, port })
  }, [host, port, isScanning, send])

  const handleKey = (e) => {
    if (e.ctrlKey && e.key === 'Enter') handleScan()
  }

  const copyText = () => {
    const text = scanResult?.decoded_text ?? JSON.stringify(scanResult, null, 2)
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div onKeyDown={handleKey}>
      <div className="page-header">
        <div>
          <h1 className="page-title"><Terminal size={22} /> TCP Banner Grabber</h1>
          <p className="page-subtitle">Connect to a TCP port and capture the service banner</p>
        </div>
      </div>

      <div style={{ display: 'grid', gap: 20 }}>
        <Card title="Scan Configuration">
          <div className="grid-2">
            <div className="form-group">
              <label>Host / IP</label>
              <input value={host} onChange={(e) => setHost(e.target.value)} placeholder="192.168.1.1 or example.com" />
            </div>
            <div className="form-group">
              <label>Port</label>
              <input
                type="number"
                min={1}
                max={65535}
                value={port}
                onChange={(e) => setPort(e.target.value)}
                placeholder="80"
              />
            </div>
          </div>
          <div style={{ marginTop: 16, display: 'flex', gap: 10, alignItems: 'center' }}>
            <Button variant="primary" size="lg" loading={isScanning} onClick={handleScan} disabled={!host || !port}>
              {isScanning ? 'Connecting…' : 'Grab Banner'}
            </Button>
            <span className="text-muted text-sm">Ctrl + Enter</span>
          </div>
        </Card>

        {isScanning && (
          <Card>
            <ScanAnimation label={`Connecting to ${host}:${port}…`} />
          </Card>
        )}

        {scanResult && !isScanning && (
          <div className="result-panel animate-in">
            <div className="result-header">
              <span className="result-title"><Terminal size={14} /> {host}:{port}</span>
              <div style={{ display: 'flex', gap: 8 }}>
                <Badge>{scanResult.status ?? 'completed'}</Badge>
                <button className="copy-btn" onClick={copyText} title="Copy response">
                  {copied ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            {scanResult.error ? (
              <div className="result-error">Error: {scanResult.error}</div>
            ) : (
              <>
                {scanResult.decoded_text ? (
                  <div style={{ padding: '0 0 0 0' }}>
                    <div style={{ padding: '10px 20px 6px', fontSize: 11, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                      Decoded Response
                    </div>
                    <pre className="decoded-text">{scanResult.decoded_text}</pre>
                  </div>
                ) : (
                  <EmptyState title="No banner received" description="The service did not return any data" />
                )}

                {scanResult.hex_object_key && (
                  <div style={{ padding: '10px 20px', borderTop: '1px solid var(--border)' }}>
                    <div style={{ fontSize: 11, color: 'var(--text-muted)', marginBottom: 6, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                      Raw Hex (MinIO Key)
                    </div>
                    <code style={{ fontSize: 12, color: 'var(--text-secondary)', wordBreak: 'break-all' }}>
                      {scanResult.hex_object_key}
                    </code>
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

