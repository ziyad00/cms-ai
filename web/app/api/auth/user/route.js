import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth/next'
import { postJSON } from '../../../../../lib/goApi.js'

export async function POST() {
  const session = await getServerSession()
  
  if (!session?.user?.id) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  try {
    // First, try to get existing user/org
    const getUserResult = await postJSON('/v1/auth/user', {
      userId: session.user.id,
      email: session.user.email,
      name: session.user.name,
    })

    if (getUserResult.status === 200) {
      return NextResponse.json(getUserResult.body)
    }

    // If user doesn't exist, create them with a default org
    const createUserResult = await postJSON('/v1/auth/signup', {
      userId: session.user.id,
      email: session.user.email,
      name: session.user.name,
    })

    return NextResponse.json(createUserResult.body, { status: createUserResult.status })
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to authenticate user' },
      { status: 500 }
    )
  }
}