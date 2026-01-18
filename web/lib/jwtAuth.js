// JWT-based authentication
const TOKEN_KEY = 'cms-ai-token'
const USER_KEY = 'cms-ai-user'

export function setAuth(token, user) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function getAuth() {
  const token = localStorage.getItem(TOKEN_KEY)
  const userStr = localStorage.getItem(USER_KEY)
  if (!token || !userStr) return null
  
  try {
    const user = JSON.parse(userStr)
    return { token, user }
  } catch {
    return null
  }
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function getAuthHeaders() {
  const auth = getAuth()
  if (!auth) return null
  
  return {
    'Authorization': `Bearer ${auth.token}`,
  }
}
