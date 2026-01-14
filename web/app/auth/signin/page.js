import { getServerSession } from 'next-auth/next'
import { redirect } from 'next/navigation'

export default async function SignInPage() {
  const session = await getServerSession()
  
  if (session) {
    redirect('/')
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
        <h1 className="text-2xl font-bold mb-4 text-center">PPTX Template CMS</h1>
        <p className="mb-6 text-center text-gray-600">Sign in to access your templates</p>
        <div className="space-y-4">
          <a
            href="/api/auth/signin/github"
            className="block w-full bg-gray-900 text-white py-2 px-4 rounded hover:bg-gray-800 text-center"
          >
            Sign in with GitHub
          </a>
        </div>
        <p className="mt-4 text-sm text-gray-500 text-center">
          By signing in, you agree to our terms of service.
        </p>
      </div>
    </div>
  )
}