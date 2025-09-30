# WebUI Setup Guide

## ✅ Problem Solved!

The root `package.json` has been created to manage the project as a **monorepo** with npm workspaces.

## 📁 Project Structure

```
5G/
├── package.json              ← Root package.json (workspaces)
├── webui/
│   └── frontend/
│       ├── package.json      ← WebUI package.json
│       ├── app/
│       ├── components/
│       └── ...
└── ...
```

## 🚀 Quick Start Options

### Option 1: From Root Directory
```bash
# Install all workspace dependencies
npm install

# Run WebUI dev server
npm run dev:webui

# Build WebUI
npm run build:webui

# Lint WebUI
npm run lint:webui
```

### Option 2: From WebUI Directory
```bash
# Go to WebUI directory
cd webui/frontend

# Install dependencies
npm install

# Run dev server
npm run dev

# Open http://localhost:3000
```

## 📋 Available Root Scripts

| Command | Description |
|---------|-------------|
| `npm run dev:webui` | Start WebUI development server |
| `npm run build:webui` | Build WebUI for production |
| `npm run start:webui` | Start WebUI production server |
| `npm run lint:webui` | Lint WebUI code |
| `npm run format:webui` | Format WebUI code |
| `npm run install:webui` | Install WebUI dependencies |
| `npm run clean` | Remove WebUI build artifacts |
| `npm run clean:all` | Remove all node_modules |

## 🔧 Setup Script

The `scripts/setup-dev-env.sh` now works correctly:

```bash
./scripts/setup-dev-env.sh
```

It will:
1. Check prerequisites
2. Install Go tools
3. Install eBPF dependencies
4. ✅ Install WebUI dependencies (now works!)
5. Set up Git hooks
6. Create directories
7. Generate configs
8. Build eBPF programs

## 🎯 Next Steps

### 1. Install Dependencies
```bash
# From root
npm install

# Or from WebUI directory
cd webui/frontend && npm install
```

### 2. Start Development
```bash
# Start WebUI
npm run dev:webui
```

### 3. Access WebUI
Open browser to: http://localhost:3000

## 🐛 Troubleshooting

### Error: ENOENT package.json
**Fixed!** The root `package.json` now exists.

### Dependencies not installing
```bash
# Clean and reinstall
npm run clean:all
npm install
```

### Port 3000 in use
```bash
# Use different port
cd webui/frontend
PORT=3001 npm run dev
```

## 📚 Development Workflow

```bash
# 1. Install everything
npm install

# 2. Start WebUI in development mode
npm run dev:webui

# 3. Make changes to files in webui/frontend/

# 4. Changes auto-reload in browser

# 5. Format code before commit
npm run format:webui

# 6. Lint code
npm run lint:webui

# 7. Build for production
npm run build:webui
```

## ✨ What's Configured

- ✅ Next.js 14 with App Router
- ✅ TypeScript with strict mode
- ✅ Tailwind CSS with custom theme
- ✅ ESLint + Prettier
- ✅ TanStack Query for data fetching
- ✅ Zustand for state management
- ✅ Zod for validation
- ✅ Recharts for visualization
- ✅ Production optimization
- ✅ Security headers

---

**Everything is ready to go!** 🚀
