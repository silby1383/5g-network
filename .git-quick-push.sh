#!/bin/bash

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🚀 Quick Git Push Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Configure git if needed
if ! git config user.name > /dev/null 2>&1; then
    echo "📝 Please enter your name for git commits:"
    read -p "Name: " git_name
    git config user.name "$git_name"
fi

if ! git config user.email > /dev/null 2>&1; then
    echo "📝 Please enter your email for git commits:"
    read -p "Email: " git_email
    git config user.email "$git_email"
fi

echo ""
echo "✅ Git user configured:"
echo "   Name: $(git config user.name)"
echo "   Email: $(git config user.email)"
echo ""

# Add all files
echo "📦 Adding all files..."
git add .

# Show what will be committed
echo ""
echo "📋 Files to commit:"
git status --short | head -20
echo ""

# Make initial commit
echo "💾 Creating initial commit..."
if git commit -m "feat: initial 5G network project setup

- Complete project structure with AI agent guide
- Network function interfaces and implementations
- Simulated UPF data plane (with clean interface for future eBPF)
- gNodeB with CU/DU split and simulated radio
- eBPF-based OpenTelemetry tracing infrastructure
- WebUI foundation (Next.js 14 + TypeScript + Tailwind)
- Kubernetes deployment configs (Kind + Helm)
- Build system (Makefile with comprehensive targets)
- Comprehensive documentation (20+ markdown files)
- Development scripts for quick start

Components included:
- AMF entry point
- gNodeB CU implementation
- UPF simulated data plane
- Common interfaces (F1, data plane)
- eBPF loader for tracing
- WebUI frontend scaffolding
- Helm charts for deployment

Ready for development of remaining NFs following AI-AGENT-GUIDE.md"; then
    echo "✅ Commit created successfully"
else
    echo "❌ Commit failed"
    exit 1
fi

# Rename branch to main (only works after commit exists)
echo "🔀 Renaming branch to main..."
if git branch -M main; then
    echo "✅ Branch renamed to main"
else
    echo "❌ Branch rename failed"
    exit 1
fi

# Push to GitHub
echo ""
echo "☁️  Pushing to GitHub..."
echo "   Repository: https://github.com/silby1383/5g-network"
echo ""
git push -u origin main

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Successfully pushed to GitHub!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌐 View your repository at:"
echo "   https://github.com/silby1383/5g-network"
echo ""
echo "🎯 Next steps:"
echo "   1. Start development: See QUICK-DEV-START.md"
echo "   2. Follow AI-AGENT-GUIDE.md for implementation"
echo "   3. Use feature branches: git checkout -b feature/nrf"
echo ""
