import { NextResponse } from 'next/server'
import { withAuth } from '../../../../lib/auth'

export const GET = withAuth(async (request, { params }) => {
  const { id } = params
  
  try {
    const response = await fetch(`${process.env.API_BASE_URL || 'http://localhost:8080'}/v1/organizations/${id}`, {
      headers: request.auth
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }
    
    return NextResponse.json(data)
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to fetch organization' },
      { status: 500 }
    )
  }
})

export const PUT = withAuth(async (request, { params }) => {
  const { id } = params
  const body = await request.json()
  
  try {
    const response = await fetch(`${process.env.API_BASE_URL || 'http://localhost:8080'}/v1/organizations/${id}`, {
      method: 'PUT',
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
      { error: 'Failed to update organization' },
      { status: 500 }
    )
  }
})