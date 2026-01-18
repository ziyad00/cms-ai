import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function GET() {
  // Health check endpoint - proxy to Go backend or return ok
  // This is for Railway health checks
  return NextResponse.json({ status: 'ok' })
}
