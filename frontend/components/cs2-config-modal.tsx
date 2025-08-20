'use client';

import { useState } from 'react';
import { Copy, Check, Terminal, Server as ServerIcon, Key, Link2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Server } from '@/lib/api';

interface CS2ConfigModalProps {
  server: Server;
}

export function CS2ConfigModal({ server }: CS2ConfigModalProps) {
  const [copied, setCopied] = useState(false);
  
  const baseUrl = process.env.NEXT_PUBLIC_LOG_INGESTION_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9090';
  const logUrl = `${baseUrl}/logs/${server.id}?key=${server.api_key}`;
  
  // Split into two lines for better display
  const configLines = [
    'log on',
    `logaddress_add_http "${logUrl}"`
  ];
  const fullConfig = configLines.join('\n');

  const copyConfig = () => {
    navigator.clipboard.writeText(fullConfig);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <Terminal className="h-4 w-4 mr-1" />
          CS2 Config
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-xl flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" />
            CS2 Server Configuration
          </DialogTitle>
          <DialogDescription>
            Copy this configuration to your CS2 server
          </DialogDescription>
        </DialogHeader>
        
        <div className="space-y-4 mt-4">
          {/* Quick Copy Section */}
          <div className="bg-secondary/50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-3 gap-2">
              <span className="text-sm font-medium truncate">Server: {server.name}</span>
              <Button
                variant="default"
                size="sm"
                onClick={copyConfig}
                className="shrink-0"
              >
                {copied ? (
                  <>
                    <Check className="h-4 w-4 mr-1" />
                    Copied!
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4 mr-1" />
                    Copy Config
                  </>
                )}
              </Button>
            </div>
            <div className="bg-background/60 rounded border">
              <div className="p-3 overflow-x-auto scrollbar-thin">
                <code className="text-xs block font-mono">
                  <div>log on</div>
                  <div className="whitespace-nowrap">logaddress_add_http "{logUrl}"</div>
                </code>
              </div>
            </div>
          </div>

          {/* Collapsible Details */}
          <details className="group">
            <summary className="cursor-pointer text-sm font-medium flex items-center gap-2 hover:text-primary transition-colors">
              <span className="group-open:rotate-90 transition-transform">▶</span>
              View Details
            </summary>
            
            <div className="mt-3 space-y-3 pl-5">
              {/* Server ID */}
              <div className="flex items-start gap-2">
                <ServerIcon className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                <div className="flex-1 min-w-0 overflow-hidden">
                  <p className="text-xs text-muted-foreground mb-1">Server ID</p>
                  <div className="bg-secondary/50 p-2 rounded overflow-hidden">
                    <code className="text-xs block overflow-x-auto whitespace-nowrap">
                      {server.id}
                    </code>
                  </div>
                </div>
              </div>

              {/* API Key */}
              <div className="flex items-start gap-2">
                <Key className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                <div className="flex-1 min-w-0 overflow-hidden">
                  <p className="text-xs text-muted-foreground mb-1">API Key</p>
                  <div className="bg-secondary/50 p-2 rounded overflow-hidden">
                    <code className="text-xs block overflow-x-auto whitespace-nowrap">
                      {server.api_key}
                    </code>
                  </div>
                </div>
              </div>

              {/* Full URL */}
              <div className="flex items-start gap-2">
                <Link2 className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                <div className="flex-1 min-w-0 overflow-hidden">
                  <p className="text-xs text-muted-foreground mb-1">Log URL</p>
                  <div className="bg-secondary/50 p-2 rounded overflow-hidden">
                    <code className="text-xs block overflow-x-auto whitespace-nowrap">
                      {logUrl}
                    </code>
                  </div>
                </div>
              </div>
            </div>
          </details>

          {/* Simple Instructions */}
          <div className="text-xs text-muted-foreground pt-2 border-t">
            <p className="mb-1">📝 Add to your server.cfg or execute via RCON</p>
            <p>🔒 Keep your API key secure</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}