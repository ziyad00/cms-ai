import { NextResponse } from 'next/server'
import { withAuth } from '../../../../../../lib/auth.js'

export const DELETE = withAuth(async (request, { params }) => {
  const { id, userId } = params
  
  try {
    const response = await fetch(`${process.env.API_BASE_URL || 'http://localhost:8080'}/v1/organizations/${id}/members/${userId}`, {
      method: 'DELETE',
      headers: request.auth
    })
    
    if (!response.ok) {
      const data = await response.json()
      return NextResponse.json(data, { status: response.status })
    }
    
    return NextResponse.json({ success: true })
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to remove member' },
      { status: 500 }
    )
  }
})