import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import App from './App'
import './styles/globals.css'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
      <Toaster
        position="bottom-right"
        toastOptions={{
          duration: 3000,
          style: {
            background:  'var(--bg-card)',
            color:       'var(--text-primary)',
            border:      '1px solid var(--border)',
            fontFamily:  'var(--font-mono)',
            fontSize:    '13px',
            borderRadius: '8px',
          },
          success: { iconTheme: { primary: 'var(--green)',  secondary: 'var(--bg-card)' } },
          error:   { iconTheme: { primary: 'var(--red)',    secondary: 'var(--bg-card)' } },
        }}
      />
    </BrowserRouter>
  </React.StrictMode>
)
