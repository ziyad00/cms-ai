// Production-ready JWT authentication using httpOnly cookies
// Client-side: minimal - just check if authenticated
// Server-side: cookies are automatically sent with requests

export function getAuth() {
  // Client-side: we can't read httpOnly cookies, so we check via API
  // The cookie is automatically sent with requests
  return null // Always return null on client - server will validate
}

export function clearAuth() {
  // Clear cookie by calling logout endpoint
  fetch('/api/auth/logout', { method: 'POST' })
    .catch(() => {}) // Ignore errors
}

export function getAuthHeaders() {
  // Client-side: don't manually add headers - cookies are sent automatically
  // Return empty object - the cookie will be included by the browser
  return {}
}
