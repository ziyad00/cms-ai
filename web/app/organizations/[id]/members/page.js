'use client'

import { useState, useEffect } from 'react'
import { useSession } from 'next-auth/react'
import { useParams, useRouter } from 'next/navigation'

const ROLES = {
  'admin': 'Admin',
  'editor': 'Editor', 
  'viewer': 'Viewer'
}

export default function OrganizationMembers() {
  const { data: session } = useSession()
  const params = useParams()
  const router = useRouter()
  const [members, setMembers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('viewer')
  const [inviting, setInviting] = useState(false)

  useEffect(() => {
    if (params.id) {
      fetchMembers()
    }
  }, [params.id])

  const fetchMembers = async () => {
    try {
      const response = await fetch(`/api/organizations/${params.id}/members`)
      if (!response.ok) {
        if (response.status === 403) {
          router.push('/templates')
          return
        }
        throw new Error('Failed to fetch members')
      }
      const data = await response.json()
      setMembers(data.members || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleInvite = async (e) => {
    e.preventDefault()
    if (!inviteEmail) return

    setInviting(true)
    setError('')

    try {
      const response = await fetch(`/api/organizations/${params.id}/invite`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: inviteEmail,
          role: inviteRole
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to send invitation')
      }

      setInviteEmail('')
      setInviteRole('viewer')
      await fetchMembers()
    } catch (err) {
      setError(err.message)
    } finally {
      setInviting(false)
    }
  }

  const handleRemoveMember = async (userId) => {
    if (!confirm('Are you sure you want to remove this member?')) return

    try {
      const response = await fetch(`/api/organizations/${params.id}/members/${userId}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        throw new Error('Failed to remove member')
      }

      setMembers(members.filter(m => m.userId !== userId))
    } catch (err) {
      setError(err.message)
    }
  }

  const handleRoleChange = async (userId, newRole) => {
    try {
      const response = await fetch(`/api/organizations/${params.id}/members/${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ role: newRole }),
      })

      if (!response.ok) {
        throw new Error('Failed to update member role')
      }

      setMembers(members.map(m => 
        m.userId === userId ? { ...m, role: newRole } : m
      ))
    } catch (err) {
      setError(err.message)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Team Members</h1>
          <p className="mt-2 text-gray-600">Manage organization members and their roles</p>
        </div>

        {error && (
          <div className="mb-6 rounded-md bg-red-50 p-4">
            <div className="text-sm text-red-700">{error}</div>
          </div>
        )}

        <div className="bg-white shadow rounded-lg mb-6">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Invite New Member</h2>
          </div>
          
          <form onSubmit={handleInvite} className="p-6">
            <div className="flex gap-4">
              <div className="flex-1">
                <input
                  type="email"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="Enter email address"
                  className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm p-2 border"
                  required
                />
              </div>
              <div>
                <select
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value)}
                  className="block border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm p-2 border"
                >
                  {Object.entries(ROLES).map(([value, label]) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </select>
              </div>
              <button
                type="submit"
                disabled={inviting || !inviteEmail}
                className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
              >
                {inviting ? 'Sending...' : 'Send Invite'}
              </button>
            </div>
          </form>
        </div>

        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Members ({members.length})</h2>
          </div>
          
          <div className="overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Role
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {members.map((member) => (
                  <tr key={member.userId}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="h-10 w-10 flex-shrink-0">
                          <div className="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                            <span className="text-sm font-medium text-gray-700">
                              {member.name?.charAt(0)?.toUpperCase() || 'U'}
                            </span>
                          </div>
                        </div>
                        <div className="ml-4">
                          <div className="text-sm font-medium text-gray-900">{member.name}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{member.email}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <select
                        value={member.role}
                        onChange={(e) => handleRoleChange(member.userId, e.target.value)}
                        className="text-sm border-gray-300 rounded shadow-sm focus:ring-blue-500 focus:border-blue-500 p-1 border"
                        disabled={member.userId === session?.user?.id}
                      >
                        {Object.entries(ROLES).map(([value, label]) => (
                          <option key={value} value={value}>{label}</option>
                        ))}
                      </select>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                        {member.status || 'Active'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      {member.userId !== session?.user?.id && (
                        <button
                          onClick={() => handleRemoveMember(member.userId)}
                          className="text-red-600 hover:text-red-900"
                        >
                          Remove
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            
            {members.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                No members found. Invite your first team member above.
              </div>
            )}
          </div>
        </div>

        <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-blue-900 mb-2">Role Permissions</h3>
          <div className="text-sm text-blue-700 space-y-1">
            <p><strong>Admin:</strong> Manage organization settings, members, and billing</p>
            <p><strong>Editor:</strong> Create and edit templates, view reports</p>
            <p><strong>Viewer:</strong> View templates and reports only</p>
          </div>
        </div>
      </div>
    </div>
  )
}