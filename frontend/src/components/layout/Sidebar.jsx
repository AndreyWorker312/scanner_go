import { NavLink } from 'react-router-dom'
import {
  Wifi, LayoutDashboard, Network, Radio,
  Shield, Terminal, Clock, Search, Bell,
} from 'lucide-react'
import { useStore }  from '@/store'
import { StatusDot } from '@/components/ui'

const SCANNERS = [
  { to: '/arp',  label: 'ARP Scanner',  Icon: Network,   color: 'var(--green)'  },
  { to: '/icmp', label: 'ICMP Ping',    Icon: Radio,     color: 'var(--blue)'   },
  { to: '/nmap', label: 'Nmap Scanner', Icon: Shield,    color: 'var(--purple)' },
  { to: '/tcp',  label: 'TCP Banner',   Icon: Terminal,  color: 'var(--yellow)' },
]

const TOOLS = [
  { to: '/history', label: 'History', Icon: Clock  },
  { to: '/search',  label: 'Search',  Icon: Search },
]

const WS_LABEL = { connected: 'Connected', disconnected: 'Disconnected', error: 'Error' }

export default function Sidebar() {
  const wsStatus         = useStore((s) => s.wsStatus)
  const activeScan       = useStore((s) => s.activeScan)
  const newChangesCount  = useStore((s) => s.newChangesCount)
  const clearNewChanges  = useStore((s) => s.clearNewChangesCount)

  return (
    <nav className="sidebar">
      <div className="sidebar-logo">
        <div className="sidebar-logo-icon">
          <Wifi size={16} />
        </div>
        <div>
          <div className="sidebar-logo-text">NetScan</div>
          <div className="sidebar-logo-sub">WebScanAPI</div>
        </div>
      </div>

      <div className="sidebar-nav">
        <NavLink to="/" end className={({ isActive }) => `nav-item${isActive ? ' active' : ''}`}>
          <LayoutDashboard size={16} className="nav-item-icon" />
          Dashboard
        </NavLink>

        <div className="nav-section-label">Scanners</div>
        {SCANNERS.map(({ to, label, Icon }) => (
          <NavLink key={to} to={to} className={({ isActive }) => `nav-item${isActive ? ' active' : ''}`}>
            <Icon size={16} className="nav-item-icon" />
            {label}
            {activeScan?.scanner_service && to.slice(1) + '_service' === activeScan.scanner_service && (
              <span className="nav-scanning-dot" />
            )}
          </NavLink>
        ))}

        <div className="nav-section-label">Tools</div>
        {TOOLS.map(({ to, label, Icon }) => (
          <NavLink key={to} to={to} className={({ isActive }) => `nav-item${isActive ? ' active' : ''}`}>
            <Icon size={16} className="nav-item-icon" />
            {label}
          </NavLink>
        ))}

        <div className="nav-section-label">Security</div>
        <NavLink
          to="/changes"
          className={({ isActive }) => `nav-item${isActive ? ' active' : ''}`}
          onClick={clearNewChanges}
        >
          <Bell size={16} className="nav-item-icon" />
          Changes
          {newChangesCount > 0 && (
            <span className="nav-alert-badge">{newChangesCount > 99 ? '99+' : newChangesCount}</span>
          )}
        </NavLink>
      </div>

      <div className="sidebar-footer">
        <div className="ws-status">
          <StatusDot status={wsStatus} />
          <span>{WS_LABEL[wsStatus] ?? 'Connecting…'}</span>
        </div>
      </div>
    </nav>
  )
}
