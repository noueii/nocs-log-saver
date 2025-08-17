'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: 'super_admin' | 'admin' | 'viewer';
  requiredPermission?: {
    resource: string;
    action: string;
  };
}

export function ProtectedRoute({ 
  children, 
  requiredRole,
  requiredPermission 
}: ProtectedRouteProps) {
  const router = useRouter();
  const { isAuthenticated, isLoading, user, hasPermission } = useAuth();

  useEffect(() => {
    if (!isLoading) {
      // Check if user is authenticated
      if (!isAuthenticated) {
        router.push('/login');
        return;
      }

      // Check role requirement
      if (requiredRole && user) {
        const roleHierarchy = {
          'super_admin': 3,
          'admin': 2,
          'viewer': 1,
        };

        const userRoleLevel = roleHierarchy[user.role];
        const requiredRoleLevel = roleHierarchy[requiredRole];

        if (userRoleLevel < requiredRoleLevel) {
          router.push('/unauthorized');
          return;
        }
      }

      // Check permission requirement
      if (requiredPermission && !hasPermission(requiredPermission.resource, requiredPermission.action)) {
        router.push('/unauthorized');
        return;
      }
    }
  }, [isAuthenticated, isLoading, user, requiredRole, requiredPermission, router, hasPermission]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}