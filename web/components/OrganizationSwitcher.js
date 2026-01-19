'use client'

import { useState, useEffect } from 'react'
import { useSession } from 'next-auth/react'

export default function OrganizationSwitcher({ currentOrgId }) {
  const { data: session } = useSession()
  const [organizations, setOrganizations] = useState([])
  const [isOpen, setIsOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (session?.user) {
      fetchUserOrganizations()
    }
  }, [session])

  const fetchUserOrganizations = async () => {
    try {
      const response = await fetch('/api/user/organizations')
      if (response.ok) {
        const data = await response.json()
        setOrganizations(data.organizations || [])
      }
    } catch (error) {
      console.error('Failed to fetch organizations:', error)
    }
  }

  const handleOrgSwitch = (orgId) => {
    if (orgId !== currentOrgId) {
      window.location.href = `/organizations/${orgId}`
    }
    setIsOpen(false)
  }

  const currentOrg = organizations.find(org => org.id === currentOrgId)

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center space-x-2 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <div className="h-6 w-6 rounded-full bg-blue-500 flex items-center justify-center">
          <span className="text-xs font-medium text-white">
            {currentOrg?.name?.charAt(0)?.toUpperCase() || 'O'}
          </span>
        </div>
        <span>{currentOrg?.name || 'Organization'}</span>
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-64 bg-white border border-gray-200 rounded-md shadow-lg z-50">
          <div className="py-1">
            {organizations.map((org) => (
              <button
                key={org.id}
                onClick={() => handleOrgSwitch(org.id)}
                className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-50 flex items-center space-x-3 ${
                  org.id === currentOrgId ? 'bg-blue-50 text-blue-700' : 'text-gray-700'
                }`}
              >
                <div className="h-6 w-6 rounded-full bg-gray-400 flex items-center justify-center">
                  <span className="text-xs font-medium text-white">
                    {org.name?.charAt(0)?.toUpperCase() || 'O'}
                  </span>
                </div>
                <div>
                  <div className="font-medium">{org.name}</div>
                  <div className="text-xs text-gray-500">{org.role}</div>
                </div>
                {org.id === currentOrgId && (
                  <svg className="w-4 h-4 text-blue-600 ml-auto" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                )}
              </button>
            ))}
          </div>
          
          <div className="border-t border-gray-200 px-4 py-2">
            <button className="w-full text-left text-sm text-gray-600 hover:text-gray-900">
              + Create new organization
            </button>
          </div>
        </div>
      )}
    </div>
  )
}