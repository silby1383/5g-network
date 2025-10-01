'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { RefreshCw, Server, Clock } from 'lucide-react';

interface NFInstance {
  nfInstanceId: string;
  nfType: string;
  nfStatus: string;
  heartBeatTimer: number;
  plmnId: { mcc: string; mnc: string };
  ipv4Addresses: string[];
  nfServices: Array<{
    serviceName: string;
    ipv4EndPoints: string[];
  }>;
  lastHeartbeat: string;
}

interface NRFStatus {
  nf_instance_id: string;
  nf_name: string;
  stats: {
    total_nfs: number;
    total_subscriptions: number;
    nfs_by_type: Record<string, number>;
    nfs_by_status: Record<string, number>;
  };
  version: string;
}

export function NRFStatus() {
  const [instances, setInstances] = useState<NFInstance[]>([]);
  const [status, setStatus] = useState<NRFStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [offline, setOffline] = useState(false);

  const fetchData = async () => {
    try {
      // Fetch registered instances
      const instancesRes = await fetch('/api/nrf?endpoint=instances');
      const instancesData = await instancesRes.json();
      
      if (instancesData.offline) {
        setOffline(true);
        return;
      }
      
      setInstances(instancesData.nfInstances || []);
      setOffline(false);

      // Fetch NRF status
      const statusRes = await fetch('/api/nrf?endpoint=status');
      const statusData = await statusRes.json();
      setStatus(statusData);
    } catch (error) {
      console.error('Failed to fetch NRF data:', error);
      setOffline(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return <div className="flex items-center justify-center p-8">Loading NRF status...</div>;
  }

  if (offline) {
    return (
      <Card className="border-red-200">
        <CardHeader>
          <CardTitle className="text-red-600">NRF Offline</CardTitle>
          <CardDescription>Unable to connect to NRF at localhost:8080</CardDescription>
        </CardHeader>
        <CardContent>
          <Button onClick={fetchData}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry Connection
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">NRF Registration Status</h2>
        <Button variant="outline" size="sm" onClick={fetchData}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* NRF Statistics */}
      {status && (
        <Card>
          <CardHeader>
            <CardTitle>NRF Statistics</CardTitle>
            <CardDescription>Network Repository Function - {status.version}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Total NFs</p>
                <p className="text-2xl font-bold">{status.stats.total_nfs}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Subscriptions</p>
                <p className="text-2xl font-bold">{status.stats.total_subscriptions}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">NF Types</p>
                <p className="text-2xl font-bold">{Object.keys(status.stats.nfs_by_type).length}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Registered</p>
                <p className="text-2xl font-bold">{status.stats.nfs_by_status.REGISTERED || 0}</p>
              </div>
            </div>

            {Object.keys(status.stats.nfs_by_type).length > 0 && (
              <div className="mt-4 pt-4 border-t">
                <p className="text-sm font-medium mb-2">NFs by Type:</p>
                <div className="flex flex-wrap gap-2">
                  {Object.entries(status.stats.nfs_by_type).map(([type, count]) => (
                    <Badge key={type} variant="secondary">
                      {type}: {count}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Registered NF Instances */}
      <div>
        <h3 className="text-lg font-semibold mb-3">Registered Network Functions</h3>
        {instances.length === 0 ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              No network functions registered yet
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {instances.map((instance) => (
              <Card key={instance.nfInstanceId}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{instance.nfType}</CardTitle>
                    <Badge
                      variant={instance.nfStatus === 'REGISTERED' ? 'default' : 'secondary'}
                    >
                      {instance.nfStatus}
                    </Badge>
                  </div>
                  <CardDescription className="font-mono text-xs">
                    {instance.nfInstanceId}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center gap-2 text-sm">
                    <Server className="h-4 w-4 text-muted-foreground" />
                    <span>{instance.ipv4Addresses?.[0] || 'N/A'}</span>
                  </div>

                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span>
                      Heartbeat: {instance.heartBeatTimer}s
                    </span>
                  </div>

                  <div className="text-sm">
                    <p className="font-medium mb-1">PLMN:</p>
                    <p className="text-muted-foreground">
                      MCC: {instance.plmnId.mcc}, MNC: {instance.plmnId.mnc}
                    </p>
                  </div>

                  {instance.nfServices && instance.nfServices.length > 0 && (
                    <div className="text-sm">
                      <p className="font-medium mb-1">Services:</p>
                      <div className="space-y-1">
                        {instance.nfServices.map((service, idx) => (
                          <div key={idx} className="text-muted-foreground">
                            {service.serviceName} - {service.ipv4EndPoints?.[0] || 'N/A'}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  <div className="text-xs text-muted-foreground pt-2 border-t">
                    Last heartbeat: {new Date(instance.lastHeartbeat).toLocaleString()}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

