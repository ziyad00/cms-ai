'use client'

import { useState, useEffect } from 'react'
import { useSession } from 'next-auth/react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import OrganizationSwitcher from '../../../components/OrganizationSwitcher'

export default function OrganizationLayout({ children }) {
  const { data: session } = useSession()
  const params = useParams()
  const [org, setOrg] = useState(null)
  const [sidebarOpen, setSidebarOpen] = useState(false)

  useEffect(() => {
    if (params.id) {
      fetchOrganization()
    }
  }, [params.id])

  const fetchOrganization = async () => {
    try {
      const response = await fetch(`/api/organizations/${params.id}`)
      if (response.ok) {
        const data = await response.json()
        setOrg(data)
      }
    } catch (error) {
      console.error('Failed to fetch organization:', error)
    }
  }

  const navigation = [
    { name: 'Dashboard', href: `/organizations/${params.id}`, icon: '📊' },
    { name: 'Templates', href: `/templates`, icon: '📄' },
    { name: 'Team Members', href: `/organizations/${params.id}/members`, icon: '👥' },
    { name: 'Settings', href: `/organizations/${params.id}/settings`, icon: '⚙️' },
  ]

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center space-x-4">
              <Link href="/templates" className="text-xl font-bold text-gray-900">
                CMS AI
              </Link>
              <OrganizationSwitcher currentOrgId={params.id} />
            </div>
            
            <div className="flex items-center space-x-4">
              <Link
                href="/templates"
                className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
              >
                Templates
              </Link>
              
              <div className="flex items-center space-x-2">
                <div className="h-8 w-8 rounded-full bg-blue-500 flex items-center justify-center">
                  <span className="text-sm font-medium text-white">
                    {session?.user?.name?.charAt(0)?.toUpperCase() || 'U'}
                  </span>
                </div>
                <span className="text-sm text-gray-700">{session?.user?.name}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="flex">
        <div className="w-64 bg-white shadow-sm min-h-screen">
          <nav className="mt-5 px-2">
            <div className="space-y-1">
              {navigation.map((item) => {
                const isActive = typeof window !== 'undefined' && 
                  window.location.pathname === item.href
                
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={`group flex items-center px-2 py-2 text-sm font-medium rounded-md ${
                      isActive
                        ? 'bg-blue-100 text-blue-700'
                        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                    }`}
                  >
                    <span className="mr-3 text-lg">{item.icon}</span>
                    {item.name}
                  </Link>
                )
              })}
            </div>
          </nav>

          {org && (
            <div className="mt-8 px-4">
              <div className="bg-gray-50 rounded-lg p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-2">Usage</h3>
                <div className="space-y-2 text-xs text-gray-600">
                  <div>
                    Templates: {org.usage?.templatesGenerated || 0}/{org.quotas?.templatesPerMonth || '∞'}
                  </div>
                  <div>
                    Members: {org.memberCount || 1}
                  </div>
                  <div>
                    Plan: {org.billing?.plan || 'Free'}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        <main className="flex-1">
          {children}
        </main>
      </div>
    </div>
  )
}