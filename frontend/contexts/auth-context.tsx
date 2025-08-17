'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: string;
  email: string;
  username: string;
  fullName: string;
  role: 'super_admin' | 'admin' | 'viewer';
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (accessToken: string, refreshToken: string, user: User) => void;
  logout: () => void;
  checkAuth: () => Promise<boolean>;
  hasPermission: (resource: string, action: string) => boolean;
  canManageServers: () => boolean;
  canManageUsers: () => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async (): Promise<boolean> => {
    try {
      const accessToken = localStorage.getItem('access_token');
      const userStr = localStorage.getItem('user');
      
      if (!accessToken || !userStr) {
        setIsLoading(false);
        return false;
      }

      const userData = JSON.parse(userStr);
      setUser(userData);
      setIsLoading(false);
      return true;
    } catch (error) {
      console.error('Auth check failed:', error);
      setIsLoading(false);
      return false;
    }
  };

  const login = (accessToken: string, refreshToken: string, userData: User) => {
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('user', JSON.stringify(userData));
    setUser(userData);
  };

  const logout = async () => {
    try {
      const accessToken = localStorage.getItem('access_token');
      if (accessToken) {
        await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9090'}/api/auth/logout`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${accessToken}`,
          },
        });
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      setUser(null);
      router.push('/login');
    }
  };

  const hasPermission = (resource: string, action: string): boolean => {
    if (!user) return false;
    
    // Super admin has all permissions
    if (user.role === 'super_admin') return true;
    
    // Define role-based permissions
    const permissions: Record<string, Record<string, string[]>> = {
      admin: {
        servers: ['create', 'read', 'update', 'delete'],
        logs: ['read'],
        users: ['read'],
      },
      viewer: {
        servers: ['read'],
        logs: ['read'],
      },
    };

    const rolePerms = permissions[user.role];
    if (!rolePerms) return false;
    
    const resourcePerms = rolePerms[resource];
    if (!resourcePerms) return false;
    
    return resourcePerms.includes(action);
  };

  const canManageServers = (): boolean => {
    return user ? user.role === 'super_admin' || user.role === 'admin' : false;
  };

  const canManageUsers = (): boolean => {
    return user ? user.role === 'super_admin' : false;
  };

  return (
    <AuthContext.Provider 
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        checkAuth,
        hasPermission,
        canManageServers,
        canManageUsers,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}