'use client'

import { useState, useEffect } from 'react'
import { getAuth, setAuth, clearAuth, getAuthHeaders } from '../lib/simpleAuth'

export default function Page() {
  const [user, setUser] = useState(null)
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [showLogin, setShowLogin] = useState(false)
  const [loginForm, setLoginForm] = useState({ userId: '', email: '', name: '' })

  useEffect(() => {
    // Check if user is already logged in
    const auth = getAuth()
    if (auth) {
      setUser(auth)
    }
  }, [])

  async function handleLogin() {
    if (!loginForm.userId || !loginForm.email) {
      setMessage('Please enter User ID and Email')
      return
    }

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(loginForm),
      })
      
      const data = await res.json()
      if (data.user) {
        setAuth(data.user)
        setUser(data.user)
        setShowLogin(false)
        setMessage('')
      } else {
        setMessage('Login failed')
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    }
  }

  function handleLogout() {
    clearAuth()
    setUser(null)
    setTemplates([])
  }

  async function generateTemplate() {
    if (!user) return
    setLoading(true)
    setMessage('')
    const headers = getAuthHeaders()
    try {
      const res = await fetch('/api/templates/generate', {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          ...headers,
        },
        body: JSON.stringify({ prompt: 'Corporate presentation template' }),
      })
      if (!res.ok) {
        setMessage(`Error: ${res.status}`)
        return
      }
      const data = await res.json()
      setMessage(`Generated template: ${data.template.name}`)
      loadTemplates()
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  async function loadTemplates() {
    if (!user) return
    const headers = getAuthHeaders()
    try {
      const res = await fetch('/api/templates', { 
        method: 'GET',
        headers: {
          'X-User-Id': headers['X-User-Id'],
          'X-Org-Id': headers['X-Org-Id'],
          'X-Role': headers['X-Role'],
        },
      })
      if (res.ok) {
        const data = await res.json()
        setTemplates(data.templates || [])
      }
    } catch (err) {
      console.error('Failed to load templates:', err)
    }
  }

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
          <h1 className="text-2xl font-bold mb-4 text-center">PPTX Template CMS</h1>
          {!showLogin ? (
            <>
              <p className="mb-4 text-center">Please sign in to continue.</p>
              <button 
                onClick={() => setShowLogin(true)}
                className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700"
              >
                Sign In
              </button>
            </>
          ) : (
            <div className="space-y-4">
              <input
                type="text"
                placeholder="User ID"
                value={loginForm.userId}
                onChange={(e) => setLoginForm({...loginForm, userId: e.target.value})}
                className="w-full px-3 py-2 border rounded"
              />
              <input
                type="email"
                placeholder="Email"
                value={loginForm.email}
                onChange={(e) => setLoginForm({...loginForm, email: e.target.value})}
                className="w-full px-3 py-2 border rounded"
              />
              <input
                type="text"
                placeholder="Name (optional)"
                value={loginForm.name}
                onChange={(e) => setLoginForm({...loginForm, name: e.target.value})}
                className="w-full px-3 py-2 border rounded"
              />
              <button 
                onClick={handleLogin}
                className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700"
              >
                Sign In
              </button>
              <button 
                onClick={() => setShowLogin(false)}
                className="w-full bg-gray-300 text-gray-700 py-2 px-4 rounded hover:bg-gray-400"
              >
                Cancel
              </button>
              {message && <p className="text-red-500 text-sm">{message}</p>}
            </div>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">PPTX Template CMS</h1>
          <div className="flex items-center space-x-4">
            <a href={`/organizations/${user.orgId}`} className="text-blue-600 hover:text-blue-800">
              {user.orgId}
            </a>
            <span>Welcome, {user.name} ({user.role})</span>
            <a href="/templates" className="text-blue-600 hover:text-blue-800">
              Templates
            </a>
            <button 
              onClick={handleLogout}
              className="bg-red-500 text-white py-1 px-3 rounded hover:bg-red-600"
            >
              Logout
            </button>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 py-8">
        <section className="mb-8">
          <button 
            onClick={generateTemplate} 
            disabled={loading} 
            className="bg-green-500 text-white py-2 px-4 rounded hover:bg-green-600 disabled:opacity-50"
          >
            {loading ? 'Generating...' : 'Generate Template'}
          </button>
          {message && <p className="mt-2 text-green-600">{message}</p>}
        </section>
        <section>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Templates</h2>
            <button 
              onClick={loadTemplates} 
              className="bg-blue-500 text-white py-1 px-3 rounded hover:bg-blue-600"
            >
              Refresh
            </button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {templates.map(t => (
              <div key={t.id} className="bg-white p-4 rounded-lg shadow">
                <h3 className="font-semibold">{t.name}</h3>
                <p className="text-gray-600">Version: {t.latestVersionNo}</p>
                <a href={`/templates/${t.id}`} className="text-blue-500 hover:underline">Edit</a>
              </div>
            ))}
          </div>
        </section>
      </main>
    </div>
  )
}