# 5G Network Management WebUI

Modern, real-time dashboard for managing and monitoring 5G Core Network Functions.

## 🚀 Quick Start

```bash
# From project root
cd webui/frontend
npm install
npm run dev
```

Access at: **http://localhost:3000**

## 📋 Features

### Network Functions Dashboard
- **Real-time Status Monitoring** - View running status of all NFs
- **Health Checks** - Live health status from each NF
- **Process Control** - Start, Stop, and Restart NFs with one click
- **Auto-refresh** - Updates every 5 seconds automatically

### NRF Status Viewer
- **Registration Statistics** - Total NFs, subscriptions, and counts by type
- **NF Instances** - Detailed view of all registered Network Functions
- **Service Details** - PLMN info, IP addresses, and service endpoints
- **Heartbeat Monitoring** - Track last heartbeat for each NF
- **Auto-refresh** - Updates every 10 seconds

## 🎨 Technologies

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first styling
- **shadcn/ui** - Beautiful, accessible components
- **Lucide React** - Modern icon library

## 📡 API Endpoints

### NF Management API

**GET /api/nf**
- Returns status of all Network Functions
- Includes process status and health checks

**POST /api/nf**
- Control NF lifecycle
- Body: `{ "action": "start|stop|restart", "nf": "NRF|UDR|UDM|AUSF|AMF|SMF" }`

### NRF API

**GET /api/nrf?endpoint=instances**
- Get all registered NF instances from NRF

**GET /api/nrf?endpoint=status**
- Get NRF statistics and counts

**GET /api/nrf?endpoint=health**
- Check NRF health status

**GET /api/nrf?endpoint=discover&nf-type=SMF**
- Discover specific NF types

## 🏗️ Architecture

```
webui/frontend/
├── app/
│   ├── api/
│   │   ├── nf/route.ts       # NF management API
│   │   └── nrf/route.ts      # NRF proxy API
│   ├── globals.css           # Global styles
│   ├── layout.tsx            # Root layout
│   └── page.tsx              # Main dashboard
├── components/
│   ├── ui/                   # shadcn/ui components
│   ├── nf-dashboard.tsx      # NF control panel
│   └── nrf-status.tsx        # NRF viewer
└── components.json           # shadcn config
```

## 🔧 Development

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## 🎯 Usage Examples

### Via Browser
1. Navigate to http://localhost:3000
2. Switch between "Network Functions" and "NRF Status" tabs
3. Click Start/Stop/Restart buttons to control NFs
4. View real-time updates automatically

### Via API (cURL)

```bash
# Get all NF statuses
curl http://localhost:3000/api/nf | jq .

# Start SMF
curl -X POST http://localhost:3000/api/nf \
  -H "Content-Type: application/json" \
  -d '{"action": "start", "nf": "SMF"}' | jq .

# Stop AMF
curl -X POST http://localhost:3000/api/nf \
  -H "Content-Type: application/json" \
  -d '{"action": "stop", "nf": "AMF"}' | jq .

# Get NRF registered instances
curl http://localhost:3000/api/nrf?endpoint=instances | jq .

# Get NRF statistics
curl http://localhost:3000/api/nrf?endpoint=status | jq .
```

## 🔒 Security Notes

⚠️ **Development Mode Only**
- Current implementation executes shell commands directly
- **DO NOT** use in production without proper security
- Future: Use Kubernetes API or process managers
- Future: Add authentication and RBAC

## 🚀 Future Enhancements

- [ ] Authentication (OAuth2/OIDC)
- [ ] Role-based access control
- [ ] Kubernetes integration
- [ ] Metrics & performance graphs
- [ ] Log viewer
- [ ] Configuration editor
- [ ] Alert notifications
- [ ] Multi-cluster support
- [ ] WebSocket for real-time updates
- [ ] Export reports

## 📊 Monitoring

The WebUI automatically monitors:
- Process running status (via `ps`)
- Health endpoint responses (via HTTP)
- NRF registration status
- Heartbeat timestamps
- Service availability

## 🐛 Troubleshooting

**WebUI won't start**
```bash
# Check if port 3000 is available
lsof -i :3000

# Kill existing Next.js process
pkill -f "next dev"

# Restart
npm run dev
```

**NFs not showing up**
- Ensure NFs are running: `ps aux | grep bin/`
- Check NF ports: `lsof -i :8080-8085`
- Verify NRF is accessible: `curl http://localhost:8080/health`

**Can't control NFs**
- Check file permissions on binaries
- Verify project path in API routes matches your setup
- Check logs: `tail -f /tmp/*.log`

## 📝 License

Part of the 5G Network project.
