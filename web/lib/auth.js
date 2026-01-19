// Production-ready JWT auth - read from httpOnly cookie or Authorization header (set by middleware)
export async function getAuthHeaders(req) {
  // For server-side API routes, check Authorization header first (set by middleware)
  // Then fall back to reading JWT from httpOnly cookie
  if (req) {
    // Check if middleware already added Authorization header
    const authHeader = req.headers.get('authorization') || req.headers.get('Authorization')
    if (authHeader) {
      return {
        'Authorization': authHeader,
      }
    }
    
    // Fall back to reading cookie directly
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