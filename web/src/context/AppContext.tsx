import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'

interface AppContextType {
  wsConnected: boolean
  toast: (message: string, type?: 'success' | 'error' | 'info' | 'warning') => void
}

const AppContext = createContext<AppContextType | undefined>(undefined)

export function AppProvider({ children }: { children: ReactNode }) {
  const [wsConnected, setWsConnected] = useState(false)

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/dashboard/api/ws`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => setWsConnected(true)
    ws.onclose = () => setWsConnected(false)
    ws.onerror = () => setWsConnected(false)

    return () => ws.close()
  }, [])

  const toast = useCallback((message: string, type: 'success' | 'error' | 'info' | 'warning' = 'info') => {
    const container = document.getElementById('toast-container')
    if (!container) return

    const icons: Record<string, string> = {
      success: '✓',
      error: '✕',
      info: 'ℹ',
      warning: '⚠',
    }

    const toastEl = document.createElement('div')
    toastEl.className = `toast toast-${type}`
    toastEl.innerHTML = `
      <span class="toast-icon">${icons[type]}</span>
      <span class="toast-message">${message}</span>
    `
    container.appendChild(toastEl)
    setTimeout(() => toastEl.remove(), 3000)
  }, [])

  return (
    <AppContext.Provider value={{ wsConnected, toast }}>
      {children}
    </AppContext.Provider>
  )
}

export function useApp() {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useApp must be used within AppProvider')
  return ctx
}
