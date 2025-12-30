import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Security headers configuration for bank-grade security
const securityHeaders = {
  // Prevent clickjacking attacks
  "X-Frame-Options": "DENY",

  // Prevent MIME type sniffing
  "X-Content-Type-Options": "nosniff",

  // Enable XSS filter in older browsers
  "X-XSS-Protection": "1; mode=block",

  // Control referrer information
  "Referrer-Policy": "strict-origin-when-cross-origin",

  // HTTP Strict Transport Security - force HTTPS
  "Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",

  // Permissions Policy - restrict browser features
  "Permissions-Policy":
    "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()",

  // Content Security Policy - strict CSP for XSS prevention
  "Content-Security-Policy": [
    "default-src 'self'",
    "script-src 'self' 'unsafe-eval' 'unsafe-inline'", // Next.js requires unsafe-eval for development
    "style-src 'self' 'unsafe-inline'", // Required for styled-jsx and inline styles
    "img-src 'self' data: blob: https:",
    "font-src 'self' data:",
    "connect-src 'self' http://localhost:* ws://localhost:*",
    "frame-ancestors 'none'",
    "form-action 'self'",
    "base-uri 'self'",
    "object-src 'none'",
  ].join("; "),

  // Cross-Origin policies
  "Cross-Origin-Opener-Policy": "same-origin",
  "Cross-Origin-Resource-Policy": "same-origin",
  "Cross-Origin-Embedder-Policy": "credentialless",
};

// Paths that should have additional cache control (sensitive pages)
const sensitivePathPatterns = [
  "/login",
  "/register",
  "/dashboard",
  "/transfers",
  "/cards",
  "/profile",
  "/settings",
];

// Generate CSRF token
function generateCSRFToken(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, "0")).join(
    ""
  );
}

export function middleware(request: NextRequest) {
  const response = NextResponse.next();

  // Add all security headers
  Object.entries(securityHeaders).forEach(([key, value]) => {
    response.headers.set(key, value);
  });

  // Add request ID for tracing
  const requestId = request.headers.get("x-request-id") || generateRequestId();
  response.headers.set("X-Request-ID", requestId);

  // Check if the request is for a sensitive path
  const isSensitivePath = sensitivePathPatterns.some((pattern) =>
    request.nextUrl.pathname.startsWith(pattern)
  );

  // Add strict cache control for sensitive pages
  if (isSensitivePath) {
    response.headers.set(
      "Cache-Control",
      "no-store, no-cache, must-revalidate, private, max-age=0"
    );
    response.headers.set("Pragma", "no-cache");
    response.headers.set("Expires", "0");
  }

  // Set CSRF token in cookie if not present
  const csrfToken = request.cookies.get("csrf-token")?.value;
  if (!csrfToken && request.method === "GET") {
    const newCSRFToken = generateCSRFToken();
    response.cookies.set("csrf-token", newCSRFToken, {
      httpOnly: false, // JavaScript needs to read this for CSRF protection
      secure: process.env.NODE_ENV === "production",
      sameSite: "strict",
      path: "/",
      maxAge: 60 * 60 * 24, // 24 hours
    });
  }

  return response;
}

function generateRequestId(): string {
  const timestamp = Date.now().toString(36);
  const randomPart = Math.random().toString(36).substring(2, 15);
  return `${timestamp}-${randomPart}`;
}

// Configure which paths the middleware runs on
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes - handled by backend services)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!api|_next/static|_next/image|favicon.ico).*)",
  ],
};
