'use client'

import { useState, useEffect } from 'react'
import { useSession, signIn, signOut } from 'next-auth/react'

export default function Page() {
  const { data: session, status } = useSession()
  const [userOrg, setUserOrg] = useState(null)
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    if (session?.user) {
      // Get or create user organization
      fetch('/api/auth/user', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
          if (data.user) {
            setUserOrg(data.user)
          }
        })
        .catch(err => console.error('Failed to get user org:', err))
    }
  }, [session])

  async function generateTemplate() {
    if (!userOrg) return
    setLoading(true)
    setMessage('')
    try {
      const res = await fetch('/api/templates/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
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
    if (!userOrg) return
    try {
      const res = await fetch('/api/templates', { method: 'GET' })
      if (res.ok) {
        const data = await res.json()
        setTemplates(data.templates || [])
      }
    } catch (err) {
      console.error('Failed to load templates:', err)
    }
  }

  if (status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-2">Loading...</p>
        </div>
      </div>
    )
  }

  if (!session) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
          <h1 className="text-2xl font-bold mb-4 text-center">PPTX Template CMS</h1>
          <p className="mb-4 text-center">Please sign in to continue.</p>
          <button 
            onClick={() => signIn(process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID ? 'github' : 'dev')}
            className="w-full bg-gray-900 text-white py-2 px-4 rounded hover:bg-gray-800"
          >
            {process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID ? 'Sign in with GitHub' : 'Sign in (Dev Mode)'}
          </button>
        </div>
      </div>
    )
  }

  if (!userOrg) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-2">Setting up your organization...</p>
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
            <a href={`/organizations/${userOrg.orgId}`} className="text-blue-600 hover:text-blue-800">
              {userOrg.orgId}
            </a>
            <span>Welcome, {session.user.name} ({userOrg.role})</span>
            <a href="/templates" className="text-blue-600 hover:text-blue-800">
              Templates
            </a>
            <button 
              onClick={() => signOut()}
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