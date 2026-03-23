import { useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { wsClient } from '@/api/websocket'
import { useStore }  from '@/store'

export function useWebSocket() {
  const setWsStatus = useStore((s) => s.setWsStatus)
  const finishScan  = useStore((s) => s.finishScan)

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
    })

    return () => { offStatus(); offMsg() }
  }, [])
}

export function useSend() {
  const startScan = useStore((s) => s.startScan)

  return useCallback((scanner_service, options) => {
    wsClient.send({ type: 'scan', request: { scanner_service, options } })
    startScan(scanner_service, options)
  }, [startScan])
}

