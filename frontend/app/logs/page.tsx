'use client';

import { useState, useEffect, useCallback } from 'react';
import { Database, FileText, Search, Download, Server, ChevronDown } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

interface LogEntry {
  id: string;
  server_id: string;
  server_name?: string;
  content: string;
  type: 'raw' | 'parsed' | 'failed';
  created_at: string;
  event_type?: string;
  event_data?: any;
  error_message?: string;
  player_name?: string;
  team?: string;
}

export default function LogsPage() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [servers, setServers] = useState<{id: string, name: string}[]>([]);
  const [eventTypes, setEventTypes] = useState<{type: string, count: number}[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState({
    search: '',
    type: 'all',
    serverId: '',
    eventType: '',
  });
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);
  const [pagination, setPagination] = useState({
    limit: 100,
    offset: 0,
    hasMore: true,
  });

  useEffect(() => {
    loadServers();
    loadEventTypes();
  }, []);

  useEffect(() => {
    // Reset pagination when filters change
    setPagination(prev => ({ ...prev, offset: 0 }));
    loadLogs(true);
  }, [filter.type, filter.serverId, filter.eventType]);
  
  useEffect(() => {
    // Reload event types when server changes
    if (filter.type === 'parsed') {
      loadEventTypes();
    }
  }, [filter.serverId, filter.type]);

  const loadServers = async () => {
    try {
      const response = await fetch('http://localhost:9090/api/servers');
      const data = await response.json();
      setServers(data);
    } catch (error) {
      console.error('Failed to load servers:', error);
    }
  };
  
  const loadEventTypes = async () => {
    try {
      const params = new URLSearchParams();
      if (filter.serverId) params.append('server_id', filter.serverId);
      
      const response = await fetch(`http://localhost:9090/api/event-types?${params.toString()}`);
      const data = await response.json();
      setEventTypes(data);
    } catch (error) {
      console.error('Failed to load event types:', error);
    }
  };

  const loadLogs = useCallback(async (reset = false) => {
    try {
      setLoading(true);
      // Build query params
      const params = new URLSearchParams();
      if (filter.serverId) params.append('server_id', filter.serverId);
      if (filter.type !== 'all') params.append('type', filter.type);
      if (filter.eventType && filter.type === 'parsed') params.append('event_type', filter.eventType);
      params.append('limit', pagination.limit.toString());
      params.append('offset', reset ? '0' : pagination.offset.toString());
      
      const response = await fetch(`http://localhost:9090/api/logs?${params.toString()}`);
      const data = await response.json();
      
      // Map the API response to our LogEntry interface
      const newLogs: LogEntry[] = data.map((log: any) => ({
        id: log.id,
        server_id: log.server_id,
        server_name: servers.find(s => s.id === log.server_id)?.name || log.server_id,
        content: log.content,
        type: log.type || 'raw',
        created_at: log.created_at,
        event_type: log.event_type,
        event_data: log.event_data,
        error_message: log.error_message,
        player_name: log.player_name,
        team: log.team,
      }));
      
      if (reset) {
        setLogs(newLogs);
      } else {
        setLogs(prev => [...prev, ...newLogs]);
      }
      
      setPagination(prev => ({
        ...prev,
        hasMore: newLogs.length === pagination.limit,
      }));
    } catch (error) {
      console.error('Failed to load logs:', error);
      if (reset) setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [filter.serverId, filter.type, filter.eventType, servers, pagination.limit, pagination.offset]);

  const filteredLogs = logs.filter(log => {
    if (filter.search && !log.content.toLowerCase().includes(filter.search.toLowerCase())) {
      return false;
    }
    return true;
  });

  const loadMoreLogs = () => {
    setPagination(prev => ({
      ...prev,
      offset: prev.offset + prev.limit,
    }));
    loadLogs(false);
  };

  const downloadLogs = async (all = false) => {
    const params = new URLSearchParams();
    if (filter.serverId) params.append('server_id', filter.serverId);
    if (filter.type !== 'all') params.append('type', filter.type);
    if (filter.eventType && filter.type === 'parsed') params.append('event_type', filter.eventType);
    params.append('download', 'true');
    
    // If downloading all, set limit to 0 (backend will interpret as all)
    if (all) {
      params.append('limit', '0');
    } else {
      params.append('limit', '100');
    }
    
    window.open(`http://localhost:9090/api/logs?${params.toString()}`, '_blank');
  };

  const formatTimestamp = (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  };

  const getTypeBadgeVariant = (type: string) => {
    switch (type) {
      case 'parsed':
        return 'default';
      case 'failed':
        return 'destructive';
      default:
        return 'secondary';
    }
  };

  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            CS2 Server Logs
          </CardTitle>
          <CardDescription>
            Browse and search through all collected server logs
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Filters */}
          <div className="flex gap-4 mb-6">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search logs..."
                value={filter.search}
                onChange={(e) => setFilter({ ...filter, search: e.target.value })}
                className="pl-10"
              />
            </div>
            
            {/* Server selector */}
            <Select
              value={filter.serverId}
              onValueChange={(value) => setFilter({ ...filter, serverId: value === 'all' ? '' : value })}
            >
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="All Servers" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Servers</SelectItem>
                {servers.map(server => (
                  <SelectItem key={server.id} value={server.id}>
                    {server.name || server.id}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            
            {/* Event Type selector - only show for parsed logs */}
            {filter.type === 'parsed' && (
              <Select
                value={filter.eventType}
                onValueChange={(value) => setFilter({ ...filter, eventType: value === 'all' ? '' : value })}
              >
                <SelectTrigger className="w-[200px]">
                  <SelectValue placeholder="All Event Types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Event Types</SelectItem>
                  {eventTypes.map(eventType => (
                    <SelectItem key={eventType.type} value={eventType.type}>
                      {eventType.type} ({eventType.count})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            
            <div className="flex gap-2">
              <Button
                variant={filter.type === 'all' ? 'default' : 'outline'}
                onClick={() => setFilter({ ...filter, type: 'all', eventType: '' })}
              >
                All
              </Button>
              <Button
                variant={filter.type === 'raw' ? 'default' : 'outline'}
                onClick={() => setFilter({ ...filter, type: 'raw', eventType: '' })}
              >
                Raw
              </Button>
              <Button
                variant={filter.type === 'parsed' ? 'default' : 'outline'}
                onClick={() => setFilter({ ...filter, type: 'parsed' })}
              >
                Parsed
              </Button>
              <Button
                variant={filter.type === 'failed' ? 'default' : 'outline'}
                onClick={() => setFilter({ ...filter, type: 'failed' })}
              >
                Failed
              </Button>
            </div>
            
            {/* Download buttons */}
            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => downloadLogs(false)}
                className="gap-2"
                title="Download currently visible logs"
              >
                <Download className="h-4 w-4" />
                Download (100)
              </Button>
              <Button
                variant="outline"
                onClick={() => downloadLogs(true)}
                className="gap-2"
                title="Download all logs matching current filters"
              >
                <Download className="h-4 w-4" />
                Download All
              </Button>
            </div>
          </div>

          {/* Logs Table */}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Server</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Event</TableHead>
                <TableHead>Content</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center">
                    Loading logs...
                  </TableCell>
                </TableRow>
              ) : filteredLogs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center">
                    No logs found
                  </TableCell>
                </TableRow>
              ) : (
                filteredLogs.map((log) => (
                  <TableRow 
                    key={log.id}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => setSelectedLog(log)}
                  >
                    <TableCell className="whitespace-nowrap">
                      {formatTimestamp(log.created_at)}
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium">{log.server_name || log.server_id}</div>
                        <div className="text-xs text-muted-foreground">{log.server_id}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getTypeBadgeVariant(log.type)}>
                        {log.type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {log.event_type ? (
                        <Badge variant="outline">{log.event_type}</Badge>
                      ) : log.type === 'failed' && log.error_message ? (
                        <span className="text-xs text-destructive">Parse error</span>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell className="max-w-md truncate">
                      {log.content}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
          
          {/* Load More Button */}
          {!loading && pagination.hasMore && filteredLogs.length > 0 && (
            <div className="mt-4 flex justify-center">
              <Button
                variant="outline"
                onClick={loadMoreLogs}
                className="gap-2"
              >
                <ChevronDown className="h-4 w-4" />
                Load More Logs
              </Button>
            </div>
          )}
          
          {/* Log Count */}
          <div className="mt-4 text-sm text-muted-foreground text-center">
            Showing {filteredLogs.length} logs
            {pagination.hasMore && ' (more available)'}
          </div>
        </CardContent>
      </Card>

      {/* Log Detail Modal/Card */}
      {selectedLog && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Log Details
            </CardTitle>
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-4 top-4"
              onClick={() => setSelectedLog(null)}
            >
              ×
            </Button>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="text-sm font-medium text-muted-foreground">Log ID</div>
                <div className="font-mono">{selectedLog.id}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-muted-foreground">Server</div>
                <div>{selectedLog.server_name || selectedLog.server_id}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-muted-foreground">Timestamp</div>
                <div>{formatTimestamp(selectedLog.created_at)}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-muted-foreground">Type</div>
                <Badge variant={getTypeBadgeVariant(selectedLog.type)}>
                  {selectedLog.type}
                </Badge>
              </div>
              {selectedLog.event_type && (
                <div>
                  <div className="text-sm font-medium text-muted-foreground">Event Type</div>
                  <div>{selectedLog.event_type}</div>
                </div>
              )}
              {selectedLog.type === 'failed' && selectedLog.error_message && (
                <div>
                  <div className="text-sm font-medium text-muted-foreground">Parse Error</div>
                  <div className="text-destructive">{selectedLog.error_message}</div>
                </div>
              )}
              <div>
                <div className="text-sm font-medium text-muted-foreground">Content</div>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto whitespace-pre-wrap">
                  {selectedLog.content}
                </pre>
              </div>
              {selectedLog.event_data && (
                <div>
                  <div className="text-sm font-medium text-muted-foreground">Parsed Data</div>
                  <pre className="bg-muted p-4 rounded-md overflow-x-auto whitespace-pre-wrap">
                    {JSON.stringify(selectedLog.event_data, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}