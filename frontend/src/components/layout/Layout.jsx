import { useWebSocket } from '@/hooks/useWebSocket'
import Sidebar from './Sidebar'

export default function Layout({ children }) {
  useWebSocket()

  return (
    <div className="layout">
      <Sidebar />
      <main className="main-content">
        {children}
      </main>
    </div>
  )
}
