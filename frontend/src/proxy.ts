import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function proxy(request: NextRequest) {
  const isDashboard = request.nextUrl.pathname.startsWith('/dashboard');
  const isAuthPage = request.nextUrl.pathname.startsWith('/login') || request.nextUrl.pathname.startsWith('/register');

  const hasAccessToken = request.cookies.has('goshort_at');
  const hasRefreshToken = request.cookies.has('goshort_rt');
  const isAuthenticated = hasAccessToken || hasRefreshToken;

  if (isDashboard && !isAuthenticated) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  if (isAuthPage && isAuthenticated) {
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/dashboard/:path*', '/login', '/register'],
};