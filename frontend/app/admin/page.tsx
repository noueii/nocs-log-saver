'use client';

import { useState, useEffect } from 'react';
import { Plus, Trash2, Server as ServerIcon, LogOut, RefreshCw, Copy } from 'lucide-react';
import { api, Server } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ProtectedRoute } from '@/components/protected-route';
import { useAuth } from '@/contexts/auth-context';
import { CS2ConfigModal } from '@/components/cs2-config-modal';

function AdminDashboard() {
  const { user, logout } = useAuth();
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);
  const [showApiKey, setShowApiKey] = useState<string | null>(null);
  const [newServer, setNewServer] = useState({
    name: '',
    description: '',
  });

  useEffect(() => {
    loadServers();
  }, []);

  const loadServers = async () => {
    try {
      const data = await api.getServers();
      setServers(data);
    } catch (error) {
      console.error('Failed to load servers:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAdd = async () => {
    if (!newServer.name) return;
    
    try {
      const added = await api.createServer(newServer);
      setServers([...servers, added]);
      setNewServer({ name: '', description: '' });
      // Show the API key for the newly created server
      setShowApiKey(added.api_key);
    } catch (error) {
      console.error('Failed to add server:', error);
    }
  };

  const handleToggle = async (server: Server) => {
    try {
      await api.updateServer(server.id, {
        name: server.name,
        description: server.description,
        is_active: !server.is_active,
      });
      setServers(servers.map(s => 
        s.id === server.id ? { ...s, is_active: !s.is_active } : s
      ));
    } catch (error) {
      console.error('Failed to toggle server:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this server?')) return;
    
    try {
      await api.deleteServer(id);
      setServers(servers.filter(s => s.id !== id));
    } catch (error) {
      console.error('Failed to delete server:', error);
    }
  };

  const handleRegenerateKey = async (server: Server) => {
    if (!confirm('Are you sure you want to regenerate the API key? The old key will stop working immediately.')) return;
    
    try {
      const result = await api.regenerateApiKey(server.id);
      setShowApiKey(result.api_key);
      // Update the server in the list
      const updatedServer = await api.getServer(server.id);
      setServers(servers.map(s => s.id === server.id ? updatedServer : s));
    } catch (error) {
      console.error('Failed to regenerate API key:', error);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <div className="container mx-auto p-6">
      {/* User Info Bar */}
      <div className="mb-6 flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">Admin Dashboard</h1>
          <p className="text-muted-foreground">Welcome, {user?.fullName || user?.username}</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm font-medium">{user?.email}</p>
            <p className="text-xs text-muted-foreground">Role: {user?.role}</p>
          </div>
          <Button variant="outline" onClick={logout}>
            <LogOut className="h-4 w-4 mr-2" />
            Logout
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ServerIcon className="h-5 w-5" />
            Server Management
          </CardTitle>
          <CardDescription>
            Manage CS2 servers that can send logs to the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Add new server form */}
          <div className="flex gap-2 mb-6">
            <input
              type="text"
              placeholder="Server Name"
              value={newServer.name}
              onChange={(e) => setNewServer({ ...newServer, name: e.target.value })}
              className="flex-1 px-3 py-2 border border-input bg-background rounded-md"
            />
            <input
              type="text"
              placeholder="Description (optional)"
              value={newServer.description}
              onChange={(e) => setNewServer({ ...newServer, description: e.target.value })}
              className="flex-1 px-3 py-2 border border-input bg-background rounded-md"
            />
            <Button onClick={handleAdd}>
              <Plus className="h-4 w-4 mr-1" />
              Add Server
            </Button>
          </div>

          {/* API Key Display */}
          {showApiKey && (
            <div className="mb-6 p-4 bg-primary/10 rounded-md">
              <p className="text-sm font-medium mb-2">API Key (save this, it won't be shown again):</p>
              <div className="flex items-center gap-2">
                <code className="flex-1 p-2 bg-background rounded text-xs font-mono">{showApiKey}</code>
                <Button size="sm" variant="outline" onClick={() => copyToClipboard(showApiKey)}>
                  <Copy className="h-4 w-4" />
                </Button>
                <Button size="sm" variant="outline" onClick={() => setShowApiKey(null)}>Hide</Button>
              </div>
            </div>
          )}

          {/* Servers table */}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Server ID</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Last IP</TableHead>
                <TableHead>API Key</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Seen</TableHead>
                <TableHead>CS2 Config</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={9} className="text-center">
                    Loading...
                  </TableCell>
                </TableRow>
              ) : servers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="text-center">
                    No servers configured yet
                  </TableCell>
                </TableRow>
              ) : (
                servers.map((server) => (
                  <TableRow key={server.id}>
                    <TableCell className="font-mono text-xs">{server.id}</TableCell>
                    <TableCell>{server.name}</TableCell>
                    <TableCell>{server.description || '-'}</TableCell>
                    <TableCell className="font-mono text-xs">{server.ip_address || '-'}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {server.api_key ? server.api_key.substring(0, 10) + '...' : '-'}
                    </TableCell>
                    <TableCell>
                      <button
                        onClick={() => handleToggle(server)}
                        className={`px-2 py-1 rounded text-xs ${
                          server.is_active
                            ? 'bg-green-500/20 text-green-500'
                            : 'bg-muted text-muted-foreground'
                        }`}
                      >
                        {server.is_active ? 'Active' : 'Inactive'}
                      </button>
                    </TableCell>
                    <TableCell>
                      {server.last_seen ? new Date(server.last_seen).toLocaleString() : 'Never'}
                    </TableCell>
                    <TableCell>
                      <CS2ConfigModal server={server} />
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleRegenerateKey(server)}
                          title="Regenerate API Key"
                        >
                          <RefreshCw className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(server.id)}
                          title="Delete Server"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
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

export default function AdminPage() {
  return (
    <ProtectedRoute requiredRole="admin">
      <AdminDashboard />
    </ProtectedRoute>
  );
}