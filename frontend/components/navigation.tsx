'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, Server, Database, Activity, Shield, LogIn, LogOut } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/contexts/auth-context';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/theme-toggle';

const navItems = [
  {
    href: '/',
    label: 'Dashboard',
    icon: Home,
  },
  {
    href: '/servers',
    label: 'Servers',
    icon: Server,
  },
  {
    href: '/logs',
    label: 'Logs',
    icon: Database,
  },
  {
    href: '/sessions',
    label: 'Sessions',
    icon: Activity,
  },
  {
    href: '/admin',
    label: 'Admin',
    icon: Shield,
  },
];

export function Navigation() {
  const pathname = usePathname();
  const { isAuthenticated, user, logout, canManageServers } = useAuth();

  // Filter nav items based on permissions
  const visibleNavItems = navItems.filter(item => {
    if (item.href === '/admin' && !canManageServers()) {
      return false;
    }
    return true;
  });

  return (
    <nav className="border-b">
      <div className="container mx-auto px-6">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center space-x-8">
            <Link href="/" className="flex items-center space-x-2">
              <Database className="h-6 w-6" />
              <span className="font-bold text-xl">CS2 Log Saver</span>
            </Link>
            {isAuthenticated && (
              <div className="flex space-x-6">
                {visibleNavItems.map((item) => {
                  const Icon = item.icon;
                  const isActive = pathname === item.href;
                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={cn(
                        'flex items-center space-x-2 text-sm font-medium transition-colors hover:text-primary',
                        isActive
                          ? 'text-foreground'
                          : 'text-muted-foreground'
                      )}
                    >
                      <Icon className="h-4 w-4" />
                      <span>{item.label}</span>
                    </Link>
                  );
                })}
              </div>
            )}
          </div>
          
          <div className="flex items-center gap-4">
            <ThemeToggle />
            {isAuthenticated ? (
              <>
                <span className="text-sm text-muted-foreground">
                  {user?.username} ({user?.role})
                </span>
                <Button variant="outline" size="sm" onClick={logout}>
                  <LogOut className="h-4 w-4 mr-2" />
                  Logout
                </Button>
              </>
            ) : (
              <Link href="/login">
                <Button variant="outline" size="sm">
                  <LogIn className="h-4 w-4 mr-2" />
                  Login
                </Button>
              </Link>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}