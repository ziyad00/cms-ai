import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function POST(req) {
  try {
    const body = await req.json()
    const { name, email, password } = body

    if (!email || !password) {
      return NextResponse.json({ error: 'email and password required' }, { status: 400 })
    }

    // Use node:http for external request to bypass Next.js routing
    const http = await import('node:http')
    const { URL } = await import('node:url')

    const backendUrl = process.env.GO_API_BASE_URL || 'http://127.0.0.1:8081'
    const url = `${backendUrl}/v1/auth/signup`

    const result = await new Promise((resolve, reject) => {
      const parsedUrl = new URL(url)
      const postData = JSON.stringify({
        name: name || email.split('@')[0],
        email,
        password
      })

      const options = {
        hostname: parsedUrl.hostname,
        port: parsedUrl.port || 80,
        path: parsedUrl.pathname + parsedUrl.search,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
          'Host': parsedUrl.hostname + (parsedUrl.port ? ':' + parsedUrl.port : '')
        }
      }

      const req = http.request(options, (res) => {
        let data = ''
        res.on('data', (chunk) => { data += chunk })
        res.on('end', () => {
          let parsed
          try {
            parsed = data ? JSON.parse(data) : null
          } catch {
            parsed = { raw: data }
          }
          resolve({ status: res.statusCode, body: parsed })
        })
      })

      req.on('error', (error) => {
        reject(error)
      })

      req.setTimeout(10000, () => {
        req.destroy()
        reject(new Error('Request timeout'))
      })

      req.write(postData)
      req.end()
    })

    if (result.status === 200 && result.body.token) {
      // Set httpOnly cookie with JWT token
      const response = NextResponse.json({ user: result.body.user })

      response.cookies.set('auth-token', result.body.token, {
        httpOnly: true,
        secure: true, // Always use secure in production (Railway uses HTTPS)
        sameSite: 'lax',
        maxAge: 60 * 60 * 24 * 7, // 7 days
        path: '/',
      })

      return response
    }

    return NextResponse.json(result.body || { error: 'Sign up failed' }, { status: result.status })
  } catch (error) {
    return NextResponse.json({ error: 'Failed to sign up' }, { status: 500 })
  }
}