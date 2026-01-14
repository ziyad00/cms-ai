import { getServerSession } from 'next-auth/next'

export async function getAuthHeaders() {
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
}

export async function withAuth(handler) {
  return async (request, ...args) => {
    const authHeaders = await getAuthHeaders()
    
    if (!authHeaders) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Add auth info to request for the handler to use
    request.auth = authHeaders
    
    return handler(request, ...args)
  }
}