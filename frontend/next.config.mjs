/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    // Support individual service URLs or a shared backend URL
    const IDENTITY_URL = process.env.IDENTITY_SERVICE_URL || process.env.BACKEND_URL || 'http://localhost:8081';
    const LEDGER_URL = process.env.LEDGER_SERVICE_URL || process.env.BACKEND_URL || 'http://localhost:8082';
    const PAYMENT_URL = process.env.PAYMENT_SERVICE_URL || process.env.BACKEND_URL || 'http://localhost:8083';
    const PRODUCT_URL = process.env.PRODUCT_SERVICE_URL || process.env.BACKEND_URL || 'http://localhost:8084';
    const CARD_URL = process.env.CARD_SERVICE_URL || process.env.BACKEND_URL || 'http://localhost:8085';
    
    return [
      {
        source: '/api/identity/:path*',
        destination: `${IDENTITY_URL}/:path*`,
      },
      {
        source: '/api/ledger/:path*',
        destination: `${LEDGER_URL}/api/v1/:path*`,
      },
      {
        source: '/api/payment/:path*',
        destination: `${PAYMENT_URL}/api/v1/:path*`,
      },
      {
        source: '/api/product/:path*',
        destination: `${PRODUCT_URL}/api/v1/:path*`,
      },
      {
        source: '/api/card/:path*',
        destination: `${CARD_URL}/api/v1/:path*`,
      },
    ]
  },
}

export default nextConfig
