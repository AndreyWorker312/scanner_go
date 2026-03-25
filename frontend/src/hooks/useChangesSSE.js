/**
 * useChanges
 *
 * Provides real-time change events from the Zustand store
 * (populated via the WebSocket change_event messages from the backend).
 *
 * For historical data (events from before the current session) the caller
 * should also call api.getChanges() on mount — see ChangesPage.jsx.
 *
 * The hook preserves the same interface as the old useChangesSSE hook so
 * no other callers need to change.
 */
import { useStore } from '@/store'

export function useChangesSSE() {
  const liveEvents         = useStore((s) => s.changeEvents)
  const newCount           = useStore((s) => s.newChangesCount)
  const clearNewCount      = useStore((s) => s.clearNewChangesCount)
  const clearEvents        = useStore((s) => s.clearAllChangeEvents)
  // "connected" follows the main WS status
  const wsStatus           = useStore((s) => s.wsStatus)
  const connected          = wsStatus === 'connected'

  return { events: liveEvents, connected, newCount, clearNewCount, clearEvents }
}
