import { NextResponse } from 'next/server'
import { withAuth, canManageOrganization } from '../../../../../lib/auth'

export const POST = withAuth(async (request, { params }) => {
  const { id } = params
  const userRole = request.auth['X-Role']
  
  if (!canManageOrganization(userRole)) {
    return NextResponse.json({ error: 'Insufficient permissions' }, { status: 403 })
  }
  
  const body = await request.json()
  
  try {
    const response = await fetch(`${process.env.API_BASE_URL || 'http://localhost:8080'}/v1/organizations/${id}/invite`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...request.auth
      },
      body: JSON.stringify(body)
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }
    
    return NextResponse.json(data)
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to send invitation' },
      { status: 500 }
    )
  }
})

export const dynamic = "force-dynamic"
