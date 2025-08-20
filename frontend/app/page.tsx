'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Server, Shield, Database, Activity } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export default function Home() {
  const [stats, setStats] = useState({
    totalServers: 0,
    totalLogs: 0,
    activeSessions: 0,
    whitelistedIPs: 0,
  });

  return (
    <div className="container mx-auto p-6">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-2">CS2 Log Saver Dashboard</h1>
        <p className="text-muted-foreground">
          Monitor and manage Counter-Strike 2 server logs in real-time
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Servers</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalServers}</div>
            <p className="text-xs text-muted-foreground">Connected CS2 servers</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Logs</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalLogs}</div>
            <p className="text-xs text-muted-foreground">Stored log entries</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Sessions</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.activeSessions}</div>
            <p className="text-xs text-muted-foreground">Ongoing matches</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Whitelisted IPs</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.whitelistedIPs}</div>
            <p className="text-xs text-muted-foreground">Authorized addresses</p>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>Manage your CS2 log collection system</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Link href="/admin">
              <Button>
                <Shield className="h-4 w-4 mr-2" />
                Manage IP Whitelist
              </Button>
            </Link>
            <Link href="/servers">
              <Button variant="outline">
                <Server className="h-4 w-4 mr-2" />
                View Servers
              </Button>
            </Link>
            <Link href="/logs">
              <Button variant="outline">
                <Database className="h-4 w-4 mr-2" />
                Browse Logs
              </Button>
            </Link>
            <Link href="/sessions">
              <Button variant="outline">
                <Activity className="h-4 w-4 mr-2" />
                View Sessions
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>

      {/* Instructions */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle>CS2 Server Configuration</CardTitle>
          <CardDescription>How to configure your CS2 server to send logs</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <p>Add the following to your CS2 server configuration:</p>
            <pre className="bg-secondary p-4 rounded-md overflow-x-auto">
              <code>{`log on
logaddress_add_http "http://your-domain.com/logs/YOUR_SERVER_ID"`}</code>
            </pre>
            <p className="text-sm text-muted-foreground">
              Replace YOUR_SERVER_ID with the ID of your server from the admin panel.
              Only configured servers with valid IDs can send logs.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
