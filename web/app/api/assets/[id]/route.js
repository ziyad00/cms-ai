import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"

// Inline the functions to avoid import issues
function goApiBaseUrl() {
  return process.env.GO_API_BASE_URL || 'http://localhost:8080'
}

async function getAuthHeaders(req) {
  try {
    const { getServerSession } = await import('next-auth/next')
    const session = await getServerSession()
    
    if (!session?.user?.id) {
      return null
    }

    // Get user org info
    const userResponse = await fetch(`${process.env.NEXTAUTH_URL || 'http://localhost:3000'}/api/auth/user`, { 
      method: 'POST' 
    })
    const userData = await userResponse.json()
    
    if (!userData.user) {
      return null
    }

    return {
      'X-User-Id': userData.user.userId,
      'X-Org-Id': userData.user.orgId,
      'X-Role': userData.user.role,
    }
  } catch (e) {
    console.error('Auth error:', e)
    return null
  }
}

export async function GET(req, { params }) {
  const headers = await getAuthHeaders(req)
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  const baseUrl = goApiBaseUrl()
  
  // Proxy to the Go API asset download endpoint
  const assetUrl = `${baseUrl}/v1/assets/${id}`
  
  const response = await fetch(assetUrl, { headers })
  
  if (!response.ok) {
    return NextResponse.json({ error: 'Asset not found' }, { status: response.status })
  }
  
  // Get the content type and disposition headers
  const contentType = response.headers.get('content-type') || 'application/octet-stream'
  const contentDisposition = response.headers.get('content-disposition') || `attachment; filename="export.pptx"`
  
  // Get the asset data
  const arrayBuffer = await response.arrayBuffer()
  const buffer = Buffer.from(arrayBuffer)
  
  // Return the file with proper headers
  return new NextResponse(buffer, {
    status: 200,
    headers: {
      'Content-Type': contentType,
      'Content-Disposition': contentDisposition,
    },
  })
}