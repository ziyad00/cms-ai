import './globals.css'
import { Providers } from '../components/Providers'

export const metadata = {
  title: 'PPTX Template CMS',
  description: 'Prompt-to-template CMS',
}

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body className="bg-gray-100 min-h-screen">
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  )
}
