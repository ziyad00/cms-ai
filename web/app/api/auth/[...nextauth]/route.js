import NextAuth from 'next-auth'

function getAuthConfig() {
  const providers = []

  // Add GitHub provider if credentials are available, otherwise use dummy for build
  const clientId = process.env.GITHUB_CLIENT_ID || 'dummy-build-id'
  const clientSecret = process.env.GITHUB_CLIENT_SECRET || 'dummy-build-secret'
  
  providers.push({
    id: 'github',
    name: 'GitHub',
    type: 'oauth',
    authorization: {
      params: { scope: 'read:user user:email' },
    },
    clientId,
    clientSecret,
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

  return {
    providers,
    callbacks: {
      async jwt({ token, account, user }) {
        // Add user ID to token
        if (user) {
          token.id = user.id
        }
        return token
      },
      async session({ session, token }) {
        // Add user ID to session
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
  }
}

// Use dynamic config to avoid build-time errors
const handler = NextAuth(getAuthConfig())

export { handler as GET, handler as POST }
export const dynamic = 'force-dynamic'