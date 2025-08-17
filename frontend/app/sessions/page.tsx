'use client';

import { useState, useEffect } from 'react';
import { Activity, Clock, Users, Trophy, Target, Shield } from 'lucide-react';
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

interface Session {
  id: string;
  server_id: string;
  server_name?: string;
  match_id?: string;
  phase: 'warmup' | 'live' | 'halftime' | 'overtime' | 'postgame';
  status: 'active' | 'completed';
  started_at: string;
  ended_at?: string;
  map?: string;
  score?: {
    ct: number;
    t: number;
  };
  round?: number;
  max_rounds?: number;
  player_count?: number;
}

export default function SessionsPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'active' | 'completed'>('all');

  useEffect(() => {
    loadSessions();
    const interval = setInterval(loadSessions, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  const loadSessions = async () => {
    try {
      // In a real app, this would fetch from the API
      // const data = await api.getSessions();
      // For now, using mock data
      const mockSessions: Session[] = [
        {
          id: 'session-1',
          server_id: 'server-1',
          server_name: 'EU West #1',
          match_id: 'match-123',
          phase: 'live',
          status: 'active',
          started_at: new Date(Date.now() - 1800000).toISOString(), // 30 minutes ago
          map: 'de_dust2',
          score: { ct: 8, t: 5 },
          round: 13,
          max_rounds: 30,
          player_count: 10,
        },
        {
          id: 'session-2',
          server_id: 'server-2',
          server_name: 'US East #2',
          match_id: 'match-124',
          phase: 'warmup',
          status: 'active',
          started_at: new Date(Date.now() - 300000).toISOString(), // 5 minutes ago
          map: 'de_mirage',
          player_count: 6,
        },
        {
          id: 'session-3',
          server_id: 'server-3',
          server_name: 'Asia #1',
          match_id: 'match-122',
          phase: 'postgame',
          status: 'completed',
          started_at: new Date(Date.now() - 7200000).toISOString(), // 2 hours ago
          ended_at: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
          map: 'de_inferno',
          score: { ct: 16, t: 14 },
          round: 30,
          max_rounds: 30,
        },
      ];
      setSessions(mockSessions);
    } catch (error) {
      console.error('Failed to load sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredSessions = sessions.filter(session => {
    if (filter === 'all') return true;
    return session.status === filter;
  });

  const formatDuration = (started: string, ended?: string): string => {
    const start = new Date(started).getTime();
    const end = ended ? new Date(ended).getTime() : Date.now();
    const duration = Math.floor((end - start) / 1000);
    
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  const getPhaseColor = (phase: string): string => {
    switch (phase) {
      case 'warmup':
        return 'secondary';
      case 'live':
        return 'default';
      case 'halftime':
        return 'outline';
      case 'overtime':
        return 'destructive';
      case 'postgame':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  const getPhaseIcon = (phase: string) => {
    switch (phase) {
      case 'warmup':
        return <Clock className="h-3 w-3" />;
      case 'live':
        return <Target className="h-3 w-3" />;
      case 'halftime':
        return <Clock className="h-3 w-3" />;
      case 'overtime':
        return <Trophy className="h-3 w-3" />;
      case 'postgame':
        return <Trophy className="h-3 w-3" />;
      default:
        return <Activity className="h-3 w-3" />;
    }
  };

  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Game Sessions
          </CardTitle>
          <CardDescription>
            Monitor active and completed Counter-Strike 2 game sessions
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Filter buttons */}
          <div className="flex gap-2 mb-6">
            <button
              onClick={() => setFilter('all')}
              className={`px-4 py-2 rounded-md transition-colors ${
                filter === 'all'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary hover:bg-secondary/80'
              }`}
            >
              All Sessions
            </button>
            <button
              onClick={() => setFilter('active')}
              className={`px-4 py-2 rounded-md transition-colors ${
                filter === 'active'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary hover:bg-secondary/80'
              }`}
            >
              Active
            </button>
            <button
              onClick={() => setFilter('completed')}
              className={`px-4 py-2 rounded-md transition-colors ${
                filter === 'completed'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary hover:bg-secondary/80'
              }`}
            >
              Completed
            </button>
          </div>

          {/* Sessions Table */}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Server</TableHead>
                <TableHead>Map</TableHead>
                <TableHead>Phase</TableHead>
                <TableHead>Score</TableHead>
                <TableHead>Round</TableHead>
                <TableHead>Players</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center">
                    Loading sessions...
                  </TableCell>
                </TableRow>
              ) : filteredSessions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center">
                    No sessions found
                  </TableCell>
                </TableRow>
              ) : (
                filteredSessions.map((session) => (
                  <TableRow key={session.id}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{session.server_name || session.server_id}</div>
                        <div className="text-xs text-muted-foreground">{session.match_id || '-'}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="font-mono">{session.map || '-'}</div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getPhaseColor(session.phase) as any} className="gap-1">
                        {getPhaseIcon(session.phase)}
                        {session.phase}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {session.score ? (
                        <div className="flex items-center gap-2">
                          <div className="flex items-center gap-1">
                            <Shield className="h-3 w-3 text-blue-500" />
                            <span className="font-bold">{session.score.ct}</span>
                          </div>
                          <span className="text-muted-foreground">:</span>
                          <div className="flex items-center gap-1">
                            <span className="font-bold">{session.score.t}</span>
                            <Shield className="h-3 w-3 text-orange-500" />
                          </div>
                        </div>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>
                      {session.round && session.max_rounds ? (
                        <div className="text-sm">
                          {session.round}/{session.max_rounds}
                        </div>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>
                      {session.player_count ? (
                        <div className="flex items-center gap-1">
                          <Users className="h-3 w-3 text-muted-foreground" />
                          {session.player_count}
                        </div>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3 text-muted-foreground" />
                        {formatDuration(session.started_at, session.ended_at)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={session.status === 'active' ? 'default' : 'secondary'}
                      >
                        {session.status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Active Sessions Summary */}
      {!loading && filteredSessions.some(s => s.status === 'active') && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
          {filteredSessions
            .filter(s => s.status === 'active')
            .map(session => (
              <Card key={session.id}>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">{session.server_name || session.server_id}</CardTitle>
                  <CardDescription>{session.map || 'Unknown Map'}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Phase</span>
                      <Badge variant={getPhaseColor(session.phase) as any} className="gap-1">
                        {getPhaseIcon(session.phase)}
                        {session.phase}
                      </Badge>
                    </div>
                    {session.score && (
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-muted-foreground">Score</span>
                        <span className="font-bold">{session.score.ct} - {session.score.t}</span>
                      </div>
                    )}
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Duration</span>
                      <span>{formatDuration(session.started_at)}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
        </div>
      )}
    </div>
  );
}