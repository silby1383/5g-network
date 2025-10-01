import { NextResponse } from 'next/server';

const NRF_URL = 'http://localhost:8080';

// GET /api/nrf - Get NRF status and registered NFs
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const endpoint = searchParams.get('endpoint') || 'instances';

  try {
    let url = '';
    
    switch (endpoint) {
      case 'instances': {
        url = `${NRF_URL}/nnrf-nfm/v1/nf-instances`;
        break;
      }
      case 'status': {
        url = `${NRF_URL}/status`;
        break;
      }
      case 'health': {
        url = `${NRF_URL}/health`;
        break;
      }
      case 'discover': {
        const nfType = searchParams.get('nf-type');
        url = `${NRF_URL}/nnrf-disc/v1/nf-instances${nfType ? `?target-nf-type=${nfType}` : ''}`;
        break;
      }
      default:
        return NextResponse.json({ error: 'Invalid endpoint' }, { status: 400 });
    }

    const response = await fetch(url, {
      signal: AbortSignal.timeout(5000),
    });

    if (!response.ok) {
      return NextResponse.json(
        { error: `NRF returned status ${response.status}` },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    return NextResponse.json(
      { 
        error: error instanceof Error ? error.message : 'Failed to connect to NRF',
        offline: true 
      },
      { status: 503 }
    );
  }
}

