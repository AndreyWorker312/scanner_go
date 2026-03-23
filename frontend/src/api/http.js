const BASE = '/api'

async function request(path, options = {}) {
  const res  = await fetch(`${BASE}${path}`, options)
  const text = await res.text()
  try { return JSON.parse(text) } catch { return { success: false, error: text } }
}

export const api = {
  getHistory:    (type, params = {}) => {
    const q = new URLSearchParams(params).toString()
    return request(`/history/${type}${q ? '?' + q : ''}`)
  },
  getHistoryById: (type, id)   => request(`/history/${type}/by-id?id=${encodeURIComponent(id)}`),
  deleteHistory:  (type, params = {}) => {
    const q = new URLSearchParams(params).toString()
    return request(`/history/${type}/delete${q ? '?' + q : ''}`, { method: 'DELETE' })
  },
  search: (type, body) => request(`/search/${type}`, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(body),
  }),
}
