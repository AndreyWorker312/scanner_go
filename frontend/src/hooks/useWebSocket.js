import { useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { wsClient } from '@/api/websocket'
import { useStore }  from '@/store'

export function useWebSocket() {
  const setWsStatus    = useStore((s) => s.setWsStatus)
  const finishScan     = useStore((s) => s.finishScan)
  const addChangeEvent = useStore((s) => s.addChangeEvent)

  useEffect(() => {
    wsClient.connect()

    const offStatus = wsClient.on('status', (status) => {
      setWsStatus(status)
      if (status === 'connected')    toast.success('Backend connected',     { id: 'ws' })
      if (status === 'disconnected') toast.error('Disconnected — retrying…', { id: 'ws' })
    })

    const offMsg = wsClient.on('message', (msg) => {
      if (msg.type === 'response' && msg.response) {
        finishScan(msg.response)
      }
      // Real-time change events pushed by the change_detector via RabbitMQ → backend
      if (msg.type === 'change_event' && msg.change) {
        addChangeEvent(msg.change)
        const sev   = msg.change.severity ?? 'INFO'
        const emoji = sev === 'CRITICAL' ? '🚨' : sev === 'HIGH' ? '⚠️' : 'ℹ️'
        toast(`${emoji} ${msg.change.title}`, {
          id:       msg.change.event_id,
          duration: sev === 'CRITICAL' ? 8000 : 4000,
          style:    sev === 'CRITICAL'
            ? { background: '#1a0a0a', color: '#ff4757', border: '1px solid #ff475740' }
            : undefined,
        })
      }
    })

    return () => { offStatus(); offMsg() }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps
}

export function useSend() {
  const startScan = useStore((s) => s.startScan)

  return useCallback((scanner_service, options) => {
    if (!wsClient.connected) {
      toast.error('Not connected to backend — please wait…', { id: 'ws-send' })
      return
    }
    try {
      wsClient.send({ type: 'scan', request: { scanner_service, options } })
      startScan(scanner_service, options)
    } catch (err) {
      toast.error('Failed to send request: ' + (err?.message ?? err), { id: 'ws-send' })
    }
  }, [startScan])
}
