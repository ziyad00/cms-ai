import NextAuth from 'next-auth'

const handler = NextAuth({
  providers: [
    {
      id: 'github',
      name: 'GitHub',
      type: 'oauth',
      authorization: {
        params: { scope: 'read:user user:email' },
      },
      clientId: process.env.GITHUB_CLIENT_ID || 'dummy',
      clientSecret: process.env.GITHUB_CLIENT_SECRET || 'dummy',
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
    },
  ],
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
})

export { handler as GET, handler as POST }