import { NextResponse } from 'next/server'
import { postJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

export async function POST(req) {
  try {
    const body = await req.json()
    const { userId, email, name } = body

    if (!userId || !email) {
      return NextResponse.json({ error: 'userId and email required' }, { status: 400 })
    }

    // Try to get or create user in backend
    const getUserResult = await postJSON('/v1/auth/user', {
      userId,
      email,
      name: name || email.split('@')[0],
    })

    if (getUserResult.status === 200) {
      return NextResponse.json(getUserResult.body)
    }

    // If user doesn't exist, create them
    const createUserResult = await postJSON('/v1/auth/signup', {
      userId,
      email,
      name: name || email.split('@')[0],
    })

    if (createUserResult.status === 200) {
      return NextResponse.json(createUserResult.body)
    }

    return NextResponse.json({ error: 'Failed to authenticate' }, { status: 500 })
  } catch (error) {
    return NextResponse.json({ error: 'Failed to authenticate' }, { status: 500 })
  }
}
