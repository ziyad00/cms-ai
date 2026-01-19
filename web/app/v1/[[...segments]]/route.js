// Catch-all route to proxy /v1/* requests to the Go backend
import { NextResponse } from 'next/server'

const GO_API_BASE_URL = process.env.GO_API_BASE_URL || 'http://127.0.0.1:8081'

async function proxyToGoBackend(request, segments) {
  try {
    // Reconstruct the full path
    const path = `/v1/${segments ? segments.join('/') : ''}`
    const url = `${GO_API_BASE_URL}${path}`

    // Get search params from original request
    const { searchParams } = new URL(request.url)
    const queryString = searchParams.toString()
    const fullUrl = queryString ? `${url}?${queryString}` : url

    console.log(`Proxying ${request.method} ${path} to ${fullUrl}`)

    // Forward headers (except host)
    const headers = new Headers()
    for (const [key, value] of request.headers.entries()) {
      if (key.toLowerCase() !== 'host') {
        headers.set(key, value)
      }
    }

    // Get request body if present
    let body = undefined
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      body = await request.text()
    }

    // Make request to Go backend
    const response = await fetch(fullUrl, {
      method: request.method,
      headers: headers,
      body: body,
    })

    // Forward response
    const responseBody = await response.text()

    return new NextResponse(responseBody, {
      status: response.status,
      statusText: response.statusText,
      headers: {
        'Content-Type': response.headers.get('Content-Type') || 'application/json',
      },
    })
  } catch (error) {
    console.error('Proxy error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

export async function GET(request, { params }) {
  try {
    return proxyToGoBackend(request, params?.segments)
  } catch (error) {
    console.error('GET error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function POST(request, { params }) {
  try {
    return proxyToGoBackend(request, params?.segments)
  } catch (error) {
    console.error('POST error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PUT(request, { params }) {
  try {
    return proxyToGoBackend(request, params?.segments)
  } catch (error) {
    console.error('PUT error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function DELETE(request, { params }) {
  try {
    return proxyToGoBackend(request, params?.segments)
  } catch (error) {
    console.error('DELETE error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PATCH(request, { params }) {
  try {
    return proxyToGoBackend(request, params?.segments)
  } catch (error) {
    console.error('PATCH error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}