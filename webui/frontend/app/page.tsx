import { NFDashboard } from '@/components/nf-dashboard';
import { NRFStatus } from '@/components/nrf-status';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

export default function Home() {
  return (
    <main className="container mx-auto p-6 space-y-8">
      <div className="space-y-2">
        <h1 className="text-4xl font-bold tracking-tight">5G Network Management</h1>
        <p className="text-muted-foreground">
          Monitor and manage your 5G Core Network Functions
        </p>
      </div>

      <Tabs defaultValue="nf" className="space-y-4">
        <TabsList>
          <TabsTrigger value="nf">Network Functions</TabsTrigger>
          <TabsTrigger value="nrf">NRF Status</TabsTrigger>
        </TabsList>

        <TabsContent value="nf" className="space-y-4">
          <NFDashboard />
        </TabsContent>

        <TabsContent value="nrf" className="space-y-4">
          <NRFStatus />
        </TabsContent>
      </Tabs>
    </main>
  );
}
