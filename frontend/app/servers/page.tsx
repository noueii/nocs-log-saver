'use client';

import { useState, useEffect } from 'react';
import { Server, Clock, Wifi, WifiOff } from 'lucide-react';
import { api } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

interface ServerData {
  id: string;
  name?: string;
  ip_address: string;
  last_seen: string;
  created_at: string;
  status?: 'online' | 'offline';
  log_count?: number;
  current_session?: string;
}

export default function ServersPage() {
  const [servers, setServers] = useState<ServerData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadServers();
    const interval = setInterval(loadServers, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const loadServers = async () => {
    try {
      const data = await api.getServers();
      const enrichedServers = data.map((server: any) => ({
        ...server,
        status: isServerOnline(server.last_seen) ? 'online' : 'offline',
      }));
      setServers(enrichedServers);
    } catch (error) {
      console.error('Failed to load servers:', error);
    } finally {
      setLoading(false);
    }
  };

  const isServerOnline = (lastSeen: string): boolean => {
    const lastSeenTime = new Date(lastSeen).getTime();
    const fiveMinutesAgo = Date.now() - 5 * 60 * 1000;
    return lastSeenTime > fiveMinutesAgo;
  };

  const formatLastSeen = (lastSeen: string): string => {
    const date = new Date(lastSeen);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)} hour${Math.floor(diffMins / 60) > 1 ? 's' : ''} ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5" />
            CS2 Servers
          </CardTitle>
          <CardDescription>
            Connected Counter-Strike 2 servers sending logs to the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Server ID</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>IP Address</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Seen</TableHead>
                <TableHead>Logs</TableHead>
                <TableHead>Session</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center">
                    Loading servers...
                  </TableCell>
                </TableRow>
              ) : servers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center">
                    No servers connected yet
                  </TableCell>
                </TableRow>
              ) : (
                servers.map((server) => (
                  <TableRow key={server.id}>
                    <TableCell className="font-mono">{server.id}</TableCell>
                    <TableCell>{server.name || '-'}</TableCell>
                    <TableCell className="font-mono">{server.ip_address}</TableCell>
                    <TableCell>
                      {server.status === 'online' ? (
                        <Badge variant="default" className="gap-1">
                          <Wifi className="h-3 w-3" />
                          Online
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="gap-1">
                          <WifiOff className="h-3 w-3" />
                          Offline
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="flex items-center gap-1">
                      <Clock className="h-3 w-3 text-muted-foreground" />
                      {formatLastSeen(server.last_seen)}
                    </TableCell>
                    <TableCell>{server.log_count || 0}</TableCell>
                    <TableCell>
                      {server.current_session ? (
                        <Badge variant="outline">{server.current_session}</Badge>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}