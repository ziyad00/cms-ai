// Middleware to add Authorization header from cookie for API routes
// This allows the API routes to read the header instead of the cookie
import { NextResponse } from 'next/server'

export function middleware(request) {
  // For API routes and v1 routes (except auth routes which handle their own cookies)
  if ((request.nextUrl.pathname.startsWith('/api/') &&
       !request.nextUrl.pathname.startsWith('/api/auth/')) ||
      request.nextUrl.pathname.startsWith('/v1/')) {

    // Read auth-token cookie and add as Authorization header
    const token = request.cookies.get('auth-token')?.value

    if (token) {
      const requestHeaders = new Headers(request.headers)
      requestHeaders.set('Authorization', `Bearer ${token}`)

      return NextResponse.next({
        request: {
          headers: requestHeaders,
        },
      })
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/api/:path*', '/v1/:path*'],
}
