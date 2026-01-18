import NextAuth from 'next-auth'

// Lazy initialization to avoid build-time errors
let handler = null

function getHandler() {
  if (!handler) {
    const providers = []

    // Add GitHub provider if credentials are available
    if (process.env.GITHUB_CLIENT_ID && process.env.GITHUB_CLIENT_SECRET) {
      providers.push({
        id: 'github',
        name: 'GitHub',
        type: 'oauth',
        authorization: {
          params: { scope: 'read:user user:email' },
        },
        clientId: process.env.GITHUB_CLIENT_ID,
        clientSecret: process.env.GITHUB_CLIENT_SECRET,
        checks: ['pkce', 'state'],
        token: 'https://github.com/login/oauth/access_token',
        userinfo: 'https://api.github.com/user',
        profile(profile) {
          return {
            id: profile.id.toString(),
            name: profile.name || profile.login,
            email: profile.email,
            image: profile.avatar_url,
          }
        },
      })
    }

    handler = NextAuth({
      providers: providers.length > 0 ? providers : [{
        id: 'credentials',
        name: 'Credentials',
        type: 'credentials',
        credentials: {},
        async authorize() { return null }
      }],
      callbacks: {
        async jwt({ token, account, user }) {
          if (user) {
            token.id = user.id
          }
          return token
        },
        async session({ session, token }) {
          if (token) {
            session.user.id = token.id
          }
          return session
        },
      },
      session: {
        strategy: 'jwt',
      },
      pages: {
        signIn: '/auth/signin',
      },
    })
  }
  return handler
}

export async function GET(req, res) {
  return (await getHandler()).GET(req, res)
}

export async function POST(req, res) {
  return (await getHandler()).POST(req, res)
}

export const dynamic = 'force-dynamic'