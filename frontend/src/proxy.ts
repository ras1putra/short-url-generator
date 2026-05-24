import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const ROLE_ROUTES: Record<string, string[]> = {
  '/dashboard/links': ['user', 'admin'],
  '/dashboard/campaigns': ['advertiser', 'admin'],
  '/dashboard/wallet': ['user', 'advertiser', 'admin'],
  '/dashboard/admin': ['admin'],
};

function getRoleFromToken(request: NextRequest): string | null {
  const token = request.cookies.get('goshort_at')?.value;
  if (!token) return null;
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return decoded.role || null;
  } catch {
    return null;
  }
}

export function proxy(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  const isDashboard = pathname.startsWith('/dashboard');
  const isAuthPage = pathname.startsWith('/login') || pathname.startsWith('/register');

  const hasAccessToken = request.cookies.has('goshort_at');
  const hasRefreshToken = request.cookies.has('goshort_rt');
  const isAuthenticated = hasAccessToken || hasRefreshToken;

  if (isDashboard && !isAuthenticated) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  if (isAuthPage && isAuthenticated) {
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }

  if (isAuthenticated && isDashboard) {
    const role = getRoleFromToken(request);
    if (role) {
      const allowedRoles = ROLE_ROUTES[pathname];
      if (allowedRoles && !allowedRoles.includes(role)) {
        return NextResponse.redirect(new URL('/dashboard', request.url));
      }
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/dashboard/:path*', '/login', '/register'],
};
