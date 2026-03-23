import { useState, useCallback } from 'react'
import toast from 'react-hot-toast'
import { api } from '@/api/http'

export function useHistory(type) {
  const [records, setRecords] = useState([])
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (params = {}) => {
    setLoading(true)
    try {
      const res = await api.getHistory(type, params)
      if (res?.success) {
        const data = res.data
        if (type === 'nmap') {
          setRecords(data ?? {})
        } else {
          setRecords(Array.isArray(data) ? data : [])
        }
      } else {
        toast.error(res?.error ?? `Failed to load ${type} history`)
      }
    } catch {
      toast.error(`Failed to load ${type} history`)
    } finally {
      setLoading(false)
    }
  }, [type])

  const clear = useCallback(async (params = {}) => {
    try {
      const res = await api.deleteHistory(type, params)
      if (res?.success) {
        if (type === 'nmap') setRecords({})
        else setRecords([])
        toast.success('History cleared')
      } else {
        toast.error(res?.error ?? 'Failed to clear history')
      }
    } catch {
      toast.error('Failed to clear history')
    }
  }, [type])

  return { records, loading, load, clear }
}

