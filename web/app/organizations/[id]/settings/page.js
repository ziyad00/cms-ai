'use client'

import { useState, useEffect } from 'react'
import { useSession } from 'next-auth/react'
import { useParams, useRouter } from 'next/navigation'

export default function OrganizationSettings() {
  const { data: session } = useSession()
  const params = useParams()
  const router = useRouter()
  const [org, setOrg] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (params.id) {
      fetchOrganization()
    }
  }, [params.id])

  const fetchOrganization = async () => {
    try {
      const response = await fetch(`/api/organizations/${params.id}`)
      if (!response.ok) {
        if (response.status === 403) {
          router.push('/templates')
          return
        }
        throw new Error('Failed to fetch organization')
      }
      const data = await response.json()
      setOrg(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    setError('')

    try {
      const response = await fetch(`/api/organizations/${params.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(org),
      })

      if (!response.ok) {
        throw new Error('Failed to update organization')
      }

      const data = await response.json()
      setOrg(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  if (error && !org) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-red-600">{error}</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Organization Settings</h1>
          <p className="mt-2 text-gray-600">Manage your organization details and branding</p>
        </div>

        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">General Settings</h2>
          </div>
          
          <form onSubmit={handleSubmit} className="p-6 space-y-6">
            {error && (
              <div className="rounded-md bg-red-50 p-4">
                <div className="text-sm text-red-700">{error}</div>
              </div>
            )}

            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                Organization Name
              </label>
              <input
                type="text"
                id="name"
                value={org?.name || ''}
                onChange={(e) => setOrg({ ...org, name: e.target.value })}
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm p-2 border"
                required
              />
            </div>

            <div>
              <label htmlFor="domain" className="block text-sm font-medium text-gray-700">
                Domain
              </label>
              <input
                type="text"
                id="domain"
                value={org?.domain || ''}
                onChange={(e) => setOrg({ ...org, domain: e.target.value })}
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm p-2 border"
                placeholder="your-company.com"
              />
            </div>

            <div>
              <label htmlFor="logo" className="block text-sm font-medium text-gray-700">
                Logo URL
              </label>
              <input
                type="url"
                id="logo"
                value={org?.logo || ''}
                onChange={(e) => setOrg({ ...org, logo: e.target.value })}
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm p-2 border"
                placeholder="https://example.com/logo.png"
              />
            </div>

            <div>
              <label htmlFor="primaryColor" className="block text-sm font-medium text-gray-700">
                Primary Color
              </label>
              <input
                type="color"
                id="primaryColor"
                value={org?.branding?.primaryColor || '#3B82F6'}
                onChange={(e) => setOrg({ 
                  ...org, 
                  branding: { 
                    ...org.branding, 
                    primaryColor: e.target.value 
                  } 
                })}
                className="mt-1 block h-10 w-20 border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              />
            </div>

            <div className="pt-4">
              <button
                type="submit"
                disabled={saving}
                className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </form>
        </div>

        <div className="mt-6 bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Billing & Usage</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="bg-gray-50 p-4 rounded-lg">
              <p className="text-sm text-gray-600">Current Plan</p>
              <p className="text-lg font-semibold text-gray-900">{org?.billing?.plan || 'Free'}</p>
            </div>
            <div className="bg-gray-50 p-4 rounded-lg">
              <p className="text-sm text-gray-600">Templates Generated</p>
              <p className="text-lg font-semibold text-gray-900">
                {org?.usage?.templatesGenerated || 0} / {org?.quotas?.templatesPerMonth || '∞'}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}