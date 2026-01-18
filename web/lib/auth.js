// Simple auth - get headers from request cookies/headers
// Client-side: use simpleAuth.js
// Server-side: read from request headers (sent by client)

export async function getAuthHeaders(req) {
  // For server-side API routes, read headers from the request
  if (req) {
    // Next.js headers are case-insensitive, try both cases
    const userId = req.headers.get('x-user-id') || req.headers.get('X-User-Id')
    const orgId = req.headers.get('x-org-id') || req.headers.get('X-Org-Id')
    const role = req.headers.get('x-role') || req.headers.get('X-Role')
    
    if (!userId || !orgId) {
      return null
    }
    
    return {
      'X-User-Id': userId,
      'X-Org-Id': orgId,
      'X-Role': role || 'Editor',
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