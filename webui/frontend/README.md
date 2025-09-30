# 5G Core Network Management WebUI

Modern Next.js-based management interface for the 5G Core Network.

## Features

- **Network Function Monitoring** - Real-time status of all NFs
- **Subscriber Management** - CRUD operations for subscriber data
- **Session Management** - View and manage PDU sessions
- **Analytics Dashboard** - Metrics, traces, and logs visualization
- **Policy Management** - Configure QoS policies and slicing

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Data Fetching**: TanStack Query
- **Validation**: Zod
- **Charts**: Recharts

## Getting Started

### Prerequisites

- Node.js 18+
- npm 9+

### Installation

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Open http://localhost:3000
```

### Build

```bash
# Production build
npm run build

# Start production server
npm start
```

## Development

```bash
# Type checking
npm run type-check

# Linting
npm run lint

# Format code
npm run format

# Check formatting
npm run format:check
```

## Project Structure

```
webui/frontend/
├── app/                # Next.js 13+ app directory
│   ├── layout.tsx     # Root layout
│   ├── page.tsx       # Home page
│   └── globals.css    # Global styles
├── components/        # React components
├── lib/              # Utilities and helpers
├── public/           # Static assets
├── package.json      # Dependencies
├── tsconfig.json     # TypeScript config
├── next.config.js    # Next.js config
└── tailwind.config.ts # Tailwind config
```

## Environment Variables

Create a `.env.local` file:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## API Integration

The WebUI communicates with the 5G Core Network via REST APIs:

- **NRF**: Network function discovery
- **UDR**: Subscriber data management
- **NWDAF**: Analytics and metrics

## Contributing

See the main project README for contribution guidelines.

## License

Apache License 2.0
