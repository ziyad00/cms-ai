import { getAuthHeaders } from './auth.js'

export async function requireRole(requiredRole) {
  return async function (request, ...args) {
    const authHeaders = await getAuthHeaders()
    
    if (!authHeaders) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const userRole = authHeaders['X-Role']
    
    const roleHierarchy = {
      'admin': 3,
      'editor': 2, 
      'viewer': 1
    }
    
    const userLevel = roleHierarchy[userRole] || 0
    const requiredLevel = roleHierarchy[requiredRole] || 0
    
    if (userLevel < requiredLevel) {
      return NextResponse.json({ error: 'Insufficient permissions' }, { status: 403 })
    }

    request.auth = authHeaders
    return handler(request, ...args)
  }
}

export function hasRole(userRole, requiredRole) {
  const roleHierarchy = {
    'admin': 3,
    'editor': 2,
    'viewer': 1
  }
  
  const userLevel = roleHierarchy[userRole] || 0
  const requiredLevel = roleHierarchy[requiredRole] || 0
  
  return userLevel >= requiredLevel
}

export function canManageOrganization(role) {
  return role === 'admin'
}

export function canEditTemplates(role) {
  return role === 'admin' || role === 'editor'
}

export function canViewTemplates(role) {
  return role === 'admin' || role === 'editor' || role === 'viewer'
}