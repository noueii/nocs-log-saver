'use client';

import { useState } from 'react';
import { Upload, FileText, CheckCircle, XCircle, AlertCircle, Copy, Loader2 } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

interface ParseResult {
  line_number: number;
  content: string;
  success: boolean;
  event_type?: string;
  event_data?: any;
  error?: string;
}

interface ParseResponse {
  total_lines: number;
  parsed_count: number;
  failed_count: number;
  results: ParseResult[];
}

const SAMPLE_LOGS = `[2025-08-19T15:12:44Z] 18a5c248-c891-42a6-b72e-af0b184937c1: 08/19/2025 - 18:13:02.735 - "SHESKY<7><[U:1:215888626]><CT>" picked up "hegrenade"
[2025-08-19T15:12:45Z] 18a5c248-c891-42a6-b72e-af0b184937c1: L 08/19/2025 - 18:13:03.123 - "Player<1><STEAM_1:0:123456><CT>" killed "Enemy<2><STEAM_1:0:654321><T>" with "ak47"
[2025-08-19T15:12:46Z] 18a5c248-c891-42a6-b72e-af0b184937c1: L 08/19/2025 - 18:13:04.456 - World triggered "Round_Start"
[2025-08-19T15:12:47Z] 18a5c248-c891-42a6-b72e-af0b184937c1: L 08/19/2025 - 18:13:05.789 - "Player<1><STEAM_1:0:123456><CT>" say "gg wp"
18a5c248-c891-42a6-b72e-af0b184937c1: L 08/19/2025 - 18:13:06.012 - "Player<1><STEAM_1:0:123456><CT>" purchased "m4a1"
08/19/2025 - 18:13:07.345 - "Player<1><STEAM_1:0:123456><CT>" threw flashbang [1234 5678 90]
L 08/19/2025 - 18:13:08.678 - World triggered "Round_End"
[2025-08-19T15:12:48Z] 18a5c248-c891-42a6-b72e-af0b184937c1: Invalid log line that will fail parsing
L 08/19/2025 - 18:13:09.901 - Game Over: competitive de_mirage score 16:14 after 30 min
[2025-08-19T15:12:49Z] 18a5c248-c891-42a6-b72e-af0b184937c1: L 08/19/2025 - 18:13:10.234 - "Player<1><STEAM_1:0:123456>" connected, address "192.168.1.100:27005"`;

export default function ParseTestPage() {
  const [logs, setLogs] = useState('');
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<ParseResponse | null>(null);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('all');
  const [eventTypeFilter, setEventTypeFilter] = useState<string>('all');

  const handleParse = async () => {
    if (!logs.trim()) {
      setError('Please enter some logs to parse');
      return;
    }

    setLoading(true);
    setError('');
    setResponse(null);

    try {
      const token = localStorage.getItem('token');
      const res = await fetch('http://localhost:9090/api/parse-test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ logs }),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || 'Failed to parse logs');
      }

      const data = await res.json();
      setResponse(data);
      
      // Auto-switch to appropriate tab
      if (data.failed_count > 0 && data.parsed_count === 0) {
        setActiveTab('failed');
      } else if (data.parsed_count > 0 && data.failed_count === 0) {
        setActiveTab('parsed');
      }
    } catch (err: any) {
      setError(err.message || 'An error occurred while parsing logs');
    } finally {
      setLoading(false);
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setLogs(e.target?.result as string);
      };
      reader.readAsText(file);
    }
  };

  const loadSampleLogs = () => {
    setLogs(SAMPLE_LOGS);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getFilteredResults = () => {
    if (!response) return [];
    
    let results = response.results;
    
    // Filter by tab
    switch (activeTab) {
      case 'parsed':
        results = results.filter(r => r.success);
        break;
      case 'failed':
        results = results.filter(r => !r.success);
        break;
    }
    
    // Filter by event type if in parsed tab
    if (activeTab === 'parsed' && eventTypeFilter !== 'all') {
      results = results.filter(r => r.event_type === eventTypeFilter);
    }
    
    return results;
  };
  
  const getUniqueEventTypes = () => {
    if (!response) return [];
    
    const eventTypes = new Map<string, number>();
    response.results
      .filter(r => r.success && r.event_type)
      .forEach(r => {
        const count = eventTypes.get(r.event_type!) || 0;
        eventTypes.set(r.event_type!, count + 1);
      });
    
    return Array.from(eventTypes.entries())
      .map(([type, count]) => ({ type, count }))
      .sort((a, b) => b.count - a.count);
  };

  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            CS2 Log Parser Test
          </CardTitle>
          <CardDescription>
            Upload or paste CS2 server logs to test the parsing functionality
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Input Section */}
            <div>
              <div className="flex gap-2 mb-2">
                <Button
                  variant="outline"
                  onClick={loadSampleLogs}
                  size="sm"
                >
                  Load Sample Logs
                </Button>
                <label htmlFor="file-upload">
                  <span className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-3 cursor-pointer">
                    <Upload className="h-4 w-4 mr-2" />
                    Upload File
                  </span>
                  <input
                    id="file-upload"
                    type="file"
                    accept=".txt,.log"
                    onChange={handleFileUpload}
                    className="hidden"
                  />
                </label>
              </div>
              
              <Textarea
                placeholder="Paste your CS2 server logs here..."
                value={logs}
                onChange={(e) => setLogs(e.target.value)}
                className="min-h-[200px] font-mono text-xs"
              />
            </div>

            {/* Parse Button */}
            <div className="flex justify-between items-center">
              <div className="text-sm text-muted-foreground">
                {logs && `${logs.split('\n').filter(l => l.trim()).length} lines`}
              </div>
              <Button
                onClick={handleParse}
                disabled={loading || !logs.trim()}
                className="gap-2"
              >
                {loading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Parsing...
                  </>
                ) : (
                  <>
                    <FileText className="h-4 w-4" />
                    Parse Logs
                  </>
                )}
              </Button>
            </div>

            {/* Error Alert */}
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {/* Results Section */}
            {response && (
              <div className="space-y-4">
                {/* Statistics */}
                <div className="grid grid-cols-3 gap-4">
                  <Card>
                    <CardContent className="pt-6">
                      <div className="text-2xl font-bold">{response.total_lines}</div>
                      <p className="text-xs text-muted-foreground">Total Lines</p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="pt-6">
                      <div className="text-2xl font-bold text-green-600">
                        {response.parsed_count}
                      </div>
                      <p className="text-xs text-muted-foreground">Successfully Parsed</p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="pt-6">
                      <div className="text-2xl font-bold text-red-600">
                        {response.failed_count}
                      </div>
                      <p className="text-xs text-muted-foreground">Failed to Parse</p>
                    </CardContent>
                  </Card>
                </div>

                {/* Results Tabs */}
                <Tabs value={activeTab} onValueChange={(value) => {
                  setActiveTab(value);
                  setEventTypeFilter('all'); // Reset filter when changing tabs
                }}>
                  <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="all">
                      All ({response.total_lines})
                    </TabsTrigger>
                    <TabsTrigger value="parsed">
                      Parsed ({response.parsed_count})
                    </TabsTrigger>
                    <TabsTrigger value="failed">
                      Failed ({response.failed_count})
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value={activeTab} className="mt-4">
                    {/* Event Type Filter for Parsed Logs */}
                    {activeTab === 'parsed' && getUniqueEventTypes().length > 0 && (
                      <div className="mb-4">
                        <div className="flex gap-2 flex-wrap">
                          <Button
                            variant={eventTypeFilter === 'all' ? 'default' : 'outline'}
                            size="sm"
                            onClick={() => setEventTypeFilter('all')}
                          >
                            All Types ({response.parsed_count})
                          </Button>
                          {getUniqueEventTypes().map(({ type, count }) => (
                            <Button
                              key={type}
                              variant={eventTypeFilter === type ? 'default' : 'outline'}
                              size="sm"
                              onClick={() => setEventTypeFilter(type)}
                            >
                              {type} ({count})
                            </Button>
                          ))}
                        </div>
                      </div>
                    )}
                    
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[60px]">Line</TableHead>
                          <TableHead className="w-[80px]">Status</TableHead>
                          <TableHead className="w-[120px]">Event Type</TableHead>
                          <TableHead>Content</TableHead>
                          <TableHead className="w-[100px]">Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {getFilteredResults().map((result) => (
                          <TableRow key={result.line_number}>
                            <TableCell className="font-mono text-xs">
                              {result.line_number}
                            </TableCell>
                            <TableCell>
                              {result.success ? (
                                <Badge variant="default" className="gap-1">
                                  <CheckCircle className="h-3 w-3" />
                                  Parsed
                                </Badge>
                              ) : (
                                <Badge variant="destructive" className="gap-1">
                                  <XCircle className="h-3 w-3" />
                                  Failed
                                </Badge>
                              )}
                            </TableCell>
                            <TableCell>
                              {result.event_type ? (
                                <Badge variant="outline">{result.event_type}</Badge>
                              ) : (
                                <span className="text-muted-foreground">-</span>
                              )}
                            </TableCell>
                            <TableCell>
                              <div className="max-w-md">
                                <div className="font-mono text-xs truncate">
                                  {result.content}
                                </div>
                                {result.error && (
                                  <div className="text-xs text-red-600 mt-1">
                                    Error: {result.error}
                                  </div>
                                )}
                              </div>
                            </TableCell>
                            <TableCell>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => copyToClipboard(result.content)}
                                title="Copy log line"
                              >
                                <Copy className="h-4 w-4" />
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>

                    {getFilteredResults().length === 0 && (
                      <div className="text-center py-8 text-muted-foreground">
                        {activeTab === 'parsed' && eventTypeFilter !== 'all' 
                          ? `No logs with event type "${eventTypeFilter}"`
                          : `No ${activeTab === 'parsed' ? 'parsed' : activeTab === 'failed' ? 'failed' : ''} logs to display`}
                      </div>
                    )}
                    
                    {/* Show filtered count if filtering */}
                    {activeTab === 'parsed' && eventTypeFilter !== 'all' && getFilteredResults().length > 0 && (
                      <div className="text-sm text-muted-foreground mt-4">
                        Showing {getFilteredResults().length} of {response.results.filter(r => r.success).length} parsed logs
                      </div>
                    )}
                  </TabsContent>
                </Tabs>

                {/* Parsed Data Details */}
                {activeTab === 'parsed' && getFilteredResults().some(r => r.event_data) && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">
                        Parsed Event Data
                        {eventTypeFilter !== 'all' && (
                          <span className="ml-2 font-normal text-muted-foreground">
                            ({eventTypeFilter} events)
                          </span>
                        )}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        {getFilteredResults()
                          .filter(r => r.event_data)
                          .map((result) => (
                            <div key={result.line_number} className="border rounded-lg p-3">
                              <div className="flex justify-between items-start mb-2">
                                <Badge variant="outline">Line {result.line_number}</Badge>
                                <Badge>{result.event_type}</Badge>
                              </div>
                              <pre className="bg-muted p-2 rounded text-xs overflow-x-auto">
                                {JSON.stringify(result.event_data, null, 2)}
                              </pre>
                            </div>
                          ))}
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}