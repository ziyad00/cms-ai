// Simple JWT-based auth - no NextAuth needed
// Backend just reads headers, so we don't need complex OAuth

const AUTH_KEY = 'cms-ai-auth'

export function setAuth(user) {
  // Store user info in localStorage
  // In production, you'd use httpOnly cookies, but this is simpler
  localStorage.setItem(AUTH_KEY, JSON.stringify({
    userId: user.userId,
    email: user.email,
    name: user.name,
    orgId: user.orgId,
    role: user.role,
  }))
}

export function getAuth() {
  const stored = localStorage.getItem(AUTH_KEY)
  if (!stored) return null
  try {
    return JSON.parse(stored)
  } catch {
    return null
  }
}

export function clearAuth() {
  localStorage.removeItem(AUTH_KEY)
}

export function getAuthHeaders() {
  const auth = getAuth()
  if (!auth) return null
  
  return {
    'X-User-Id': auth.userId,
    'X-Org-Id': auth.orgId,
    'X-Role': auth.role,
  }
}
