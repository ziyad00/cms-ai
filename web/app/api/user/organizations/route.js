import { NextResponse } from 'next/server'
import { withAuth } from '../../../../lib/auth'

export const GET = withAuth(async (request) => {
  try {
    const response = await fetch(`${process.env.API_BASE_URL || 'http://localhost:8080'}/v1/user/organizations`, {
      headers: request.auth
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }
    
    return NextResponse.json(data)
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to fetch organizations' },
      { status: 500 }
    )
  }
})