import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function POST(req) {
  try {
    return NextResponse.json({ error: 'Signup temporarily disabled' }, { status: 503 })
  } catch (error) {
    return NextResponse.json({ error: 'Server error' }, { status: 500 })
  }
}