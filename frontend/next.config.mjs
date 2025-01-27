/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost';
    return [
      {
        source: '/api/identity/:path*',
        destination: `${BACKEND_URL}:8081/:path*`,
      },
      {
        source: '/api/ledger/:path*',
        destination: `${BACKEND_URL}:8082/api/v1/:path*`,
      },
      {
        source: '/api/payment/:path*',
        destination: `${BACKEND_URL}:8083/api/v1/:path*`,
      },
      {
        source: '/api/product/:path*',
        destination: `${BACKEND_URL}:8084/api/v1/:path*`,
      },
      {
        source: '/api/card/:path*',
        destination: `${BACKEND_URL}:8085/api/v1/:path*`,
      },
    ]
  },
}

export default nextConfig
