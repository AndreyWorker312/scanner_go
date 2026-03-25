import { create } from 'zustand'

export const useStore = create((set, get) => ({
  wsStatus:      'disconnected',
  activeScan:    null,
  latestResult:  null,
  recentResults: [],

  // ── Change Detection (populated by WS change_event messages) ─────────────
  changeEvents:     [],
  newChangesCount:  0,

  setWsStatus: (wsStatus) => set({ wsStatus }),

  startScan: (scanner_service, options) =>
    set({ activeScan: { scanner_service, options, startedAt: Date.now() }, latestResult: null }),

  finishScan: (response) => {
    const active = get().activeScan
    const entry  = {
      ...response,
      scanner_service: active?.scanner_service ?? 'unknown',
      options:         active?.options ?? {},
      receivedAt:      Date.now(),
    }
    set((s) => ({
      activeScan:    null,
      latestResult:  entry,
      recentResults: [entry, ...s.recentResults].slice(0, 100),
    }))
  },

  clearRecent: () => set({ recentResults: [] }),

  // ── Change event actions ──────────────────────────────────────────────────
  addChangeEvent: (event) =>
    set((s) => ({
      changeEvents:    [event, ...s.changeEvents].slice(0, 500),
      newChangesCount: s.newChangesCount + 1,
    })),

  clearNewChangesCount: () => set({ newChangesCount: 0 }),

  clearAllChangeEvents: () => set({ changeEvents: [], newChangesCount: 0 }),
}))
