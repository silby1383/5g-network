'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { PlayCircle, StopCircle, RefreshCw, Activity } from 'lucide-react';

interface NFStatus {
  name: string;
  port: number;
  running: boolean;
  health: { status: string } | null;
  error?: string;
}

export function NFDashboard() {
  const [nfs, setNfs] = useState<NFStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchStatus = async () => {
    try {
      const response = await fetch('/api/nf');
      const data = await response.json();
      setNfs(data.nfs || []);
    } catch (error) {
      console.error('Failed to fetch NF status:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const handleAction = async (nf: string, action: string) => {
    setActionLoading(`${nf}-${action}`);
    try {
      const response = await fetch('/api/nf', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ nf, action }),
      });
      
      if (response.ok) {
        // Wait a bit for the process to start/stop
        await new Promise(resolve => setTimeout(resolve, 1000));
        await fetchStatus();
      }
    } catch (error) {
      console.error(`Failed to ${action} ${nf}:`, error);
    } finally {
      setActionLoading(null);
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center p-8">Loading...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Network Functions</h2>
        <Button
          variant="outline"
          size="sm"
          onClick={fetchStatus}
          disabled={loading}
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {nfs.map((nf) => (
          <Card key={nf.name}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">{nf.name}</CardTitle>
                <Badge variant={nf.running ? 'default' : 'secondary'}>
                  {nf.running ? 'Running' : 'Stopped'}
                </Badge>
              </div>
              <CardDescription>Port: {nf.port}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {nf.running && nf.health && (
                <div className="flex items-center gap-2 text-sm">
                  <Activity className="h-4 w-4 text-green-500" />
                  <span>Health: {nf.health.status}</span>
                </div>
              )}

              <div className="flex gap-2">
                {!nf.running ? (
                  <Button
                    size="sm"
                    onClick={() => handleAction(nf.name, 'start')}
                    disabled={actionLoading === `${nf.name}-start`}
                    className="flex-1"
                  >
                    <PlayCircle className="h-4 w-4 mr-2" />
                    Start
                  </Button>
                ) : (
                  <>
                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={() => handleAction(nf.name, 'stop')}
                      disabled={actionLoading === `${nf.name}-stop`}
                      className="flex-1"
                    >
                      <StopCircle className="h-4 w-4 mr-2" />
                      Stop
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleAction(nf.name, 'restart')}
                      disabled={actionLoading === `${nf.name}-restart`}
                      className="flex-1"
                    >
                      <RefreshCw className="h-4 w-4 mr-2" />
                      Restart
                    </Button>
                  </>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

