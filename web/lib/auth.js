// Simple auth - get headers from request cookies/headers
// Client-side: use simpleAuth.js
// Server-side: read from request headers (sent by client)

export async function getAuthHeaders(req) {
  // For server-side API routes, read headers from the request
  if (req) {
    return {
      'X-User-Id': req.headers.get('x-user-id'),
      'X-Org-Id': req.headers.get('x-org-id'),
      'X-Role': req.headers.get('x-role'),
    }
  }
  
  // Fallback: try to get from cookies (if we add cookie support later)
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