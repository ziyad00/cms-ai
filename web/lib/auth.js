// Production-ready JWT auth - read from httpOnly cookie
export async function getAuthHeaders(req) {
  // For server-side API routes, read JWT from httpOnly cookie
  if (req) {
    const token = req.cookies.get('auth-token')?.value
    if (!token) {
      return null
    }
    
    return {
      'Authorization': `Bearer ${token}`,
    }
  }
  
  return null
}

export async function withAuth(handler) {
  return async (request, ...args) => {
    const authHeaders = await getAuthHeaders(request)
    
    if (!authHeaders || !authHeaders['X-User-Id']) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Add auth info to request for the handler to use
    request.auth = authHeaders
    
    return handler(request, ...args)
  }
}

export function canManageOrganization(role) {
  return ['admin', 'owner'].includes(role)
}