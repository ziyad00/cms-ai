/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    GO_API_BASE_URL: process.env.GO_API_BASE_URL || 'http://127.0.0.1:8081',
  },
  webpack: (config, { isServer }) => {
    if (isServer) {
      config.resolve.alias = {
        ...config.resolve.alias,
        '@lib': './lib',
      }
    }
    return config
  },
}

export default nextConfig
