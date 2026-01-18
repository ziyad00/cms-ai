// JWT-based auth - read Authorization header
export async function getAuthHeaders(req) {
  // For server-side API routes, read JWT from Authorization header
  if (req) {
    const authHeader = req.headers.get('authorization') || req.headers.get('Authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return null
    }
    
    return {
      'Authorization': authHeader,
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