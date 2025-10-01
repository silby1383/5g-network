import { NextRequest, NextResponse } from 'next/server';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

const NF_CONFIGS: Record<string, { port: number; config: string; binary: string }> = {
  nrf: { port: 8080, config: 'nf/nrf/config/nrf.yaml', binary: 'bin/nrf' },
  udr: { port: 8081, config: 'nf/udr/config/udr.yaml', binary: 'bin/udr' },
  udm: { port: 8082, config: 'nf/udm/config/udm.yaml', binary: 'bin/udm' },
  ausf: { port: 8083, config: 'nf/ausf/config/ausf.yaml', binary: 'bin/ausf' },
  amf: { port: 8084, config: 'nf/amf/config/amf.yaml', binary: 'bin/amf' },
  smf: { port: 8085, config: 'nf/smf/config/smf.yaml', binary: 'bin/smf' },
};

const PROJECT_ROOT = '/home/silby/5G';

// GET /api/nf - Get status of all NFs
export async function GET() {
  try {
    const nfStatuses = await Promise.all(
      Object.entries(NF_CONFIGS).map(async ([name, config]) => {
        try {
          // Check if process is running
          const { stdout: psOutput } = await execAsync(
            `ps aux | grep "${config.binary}" | grep -v grep || true`
          );
          const isRunning = psOutput.trim().length > 0;

          // Get health status if running
          let health = null;
          if (isRunning) {
            try {
              const response = await fetch(`http://localhost:${config.port}/health`, {
                signal: AbortSignal.timeout(2000),
              });
              health = response.ok ? await response.json() : null;
            } catch {
              health = { status: 'unreachable' };
            }
          }

          return {
            name: name.toUpperCase(),
            port: config.port,
            running: isRunning,
            health,
          };
        } catch (error) {
          return {
            name: name.toUpperCase(),
            port: config.port,
            running: false,
            health: null,
            error: error instanceof Error ? error.message : 'Unknown error',
          };
        }
      })
    );

    return NextResponse.json({ nfs: nfStatuses });
  } catch (error) {
    return NextResponse.json(
      { error: error instanceof Error ? error.message : 'Failed to get NF status' },
      { status: 500 }
    );
  }
}

// POST /api/nf - Start/Stop/Restart NF
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { action, nf } = body;

    if (!nf || !NF_CONFIGS[nf.toLowerCase()]) {
      return NextResponse.json({ error: 'Invalid NF name' }, { status: 400 });
    }

    const nfLower = nf.toLowerCase();
    const config = NF_CONFIGS[nfLower];

    switch (action) {
      case 'start':
        await execAsync(
          `cd ${PROJECT_ROOT} && ./${config.binary} --config ${config.config} > /tmp/${nfLower}.log 2>&1 &`
        );
        return NextResponse.json({ success: true, message: `${nf} started` });

      case 'stop':
        await execAsync(`pkill -f "${config.binary}"`);
        return NextResponse.json({ success: true, message: `${nf} stopped` });

      case 'restart':
        await execAsync(`pkill -f "${config.binary}"`);
        await new Promise(resolve => setTimeout(resolve, 1000));
        await execAsync(
          `cd ${PROJECT_ROOT} && ./${config.binary} --config ${config.config} > /tmp/${nfLower}.log 2>&1 &`
        );
        return NextResponse.json({ success: true, message: `${nf} restarted` });

      default:
        return NextResponse.json({ error: 'Invalid action' }, { status: 400 });
    }
  } catch (error) {
    return NextResponse.json(
      { error: error instanceof Error ? error.message : 'Operation failed' },
      { status: 500 }
    );
  }
}

