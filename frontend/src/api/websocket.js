const WS_URL = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws`

class WSClient {
  _ws           = null
  _listeners    = new Map()
  _reconnectTid = null
  _retryDelay   = 2000

  get connected() {
    return this._ws?.readyState === WebSocket.OPEN
  }

  connect() {
    if (this.connected) return

    this._ws = new WebSocket(WS_URL)

    this._ws.onopen = () => {
      this._retryDelay = 2000
      this._emit('status', 'connected')
    }

    this._ws.onmessage = (evt) => {
      try { this._emit('message', JSON.parse(evt.data)) } catch {}
    }

    this._ws.onclose = () => {
      this._emit('status', 'disconnected')
      this._scheduleReconnect()
    }

    this._ws.onerror = () => {
      this._emit('status', 'error')
    }
  }

  send(payload) {
    if (!this.connected) throw new Error('WebSocket not connected')
    this._ws.send(JSON.stringify(payload))
  }

  on(event, fn) {
    if (!this._listeners.has(event)) this._listeners.set(event, new Set())
    this._listeners.get(event).add(fn)
    return () => this._listeners.get(event)?.delete(fn)
  }

  _emit(event, data) {
    this._listeners.get(event)?.forEach((fn) => fn(data))
  }

  _scheduleReconnect() {
    clearTimeout(this._reconnectTid)
    this._reconnectTid = setTimeout(() => {
      this._retryDelay = Math.min(this._retryDelay * 1.5, 30_000)
      this.connect()
    }, this._retryDelay)
  }
}

export const wsClient = new WSClient()

