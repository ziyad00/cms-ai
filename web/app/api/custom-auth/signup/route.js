import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function GET() {
  return NextResponse.json({
    message: 'custom-auth test route working',
    timestamp: new Date().toISOString()
  })
}

export async function POST() {
  return NextResponse.json({
    message: 'custom-auth POST test working',
    timestamp: new Date().toISOString()
  })
}