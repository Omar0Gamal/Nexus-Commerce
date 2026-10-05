import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export const config = {
  matcher: [
    "/((?!api/|_next/|_static/|_vercel|[\\w-]+\\.\\w+).*)",
  ],
}

export default function middleware(req: NextRequest) {
  const url = req.nextUrl
  const hostname = req.headers.get("host") || "egyptcommerce.com"

  const isLocal = hostname.includes("localhost")
  const baseDomain = process.env.NEXT_PUBLIC_ROOT_DOMAIN || (isLocal ? "localhost:3000" : "egyptcommerce.com")
  
  const isAppRoute = hostname === baseDomain || 
                     hostname === `app.${baseDomain}` || 
                     hostname === `admin.${baseDomain}` ||
                     hostname.startsWith("127.0.0.1")

  if (isAppRoute) {
    if (url.pathname.startsWith('/admin') && url.pathname !== '/admin/login') {
      const token = req.cookies.get('admin_access_token')?.value
      if (!token) {
        return NextResponse.redirect(new URL('/admin/login', req.url))
      }
    }
    return NextResponse.next()
  }

  // Multi-tenant storefront requests on custom domains or subdomains rewrite to /app/[domain]/path without port.
  const hostOnly = hostname.split(':')[0]
  return NextResponse.rewrite(new URL(`/${hostOnly}${url.pathname}`, req.url))
}
