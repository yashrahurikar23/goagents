#!/bin/bash
# GitHub Secho "🔗 Adding GitHub remote..."
git remote add origin git@github.com:yashrahurikar23/goagents.gitup Script for GoAgents
# Run this after creating the GitHub repository

cd /Users/yashrahurikar/yash/projects/tweeny/agentspace/goagents

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Not in goagents directory"
    exit 1
fi

echo "📦 Setting up GoAgents repository..."

# Initialize git if not already initialized
if [ ! -d ".git" ]; then
    echo "Initializing git repository..."
    git init
    git branch -M main
fi

# Add remote origin
echo "Adding GitHub remote..."
git remote add origin https://github.com/yashrahurikar/goagents.git

# Or if remote already exists, update it
git remote set-url origin https://github.com/yashrahurikar/goagents.git 2>/dev/null || true

# Check remote
echo "✅ Remote configured:"
git remote -v

# Stage all files
echo "Staging files..."
git add .

# Create initial commit
echo "Creating commit..."
git commit -m "Initial release: GoAgents v0.1.0

- Core agent types: FunctionAgent, ReActAgent, ConversationalAgent
- LLM providers: OpenAI and Ollama (local AI support)
- Tool system with calculator example
- Memory management with 4 strategies
- 100+ tests passing
- Complete documentation and examples

Let's Go, Agents! 🚀"

# Push to GitHub
echo "Pushing to GitHub..."
git push -u origin main

# Create and push tag
echo "Creating release tag v0.1.0..."
git tag -a v0.1.0 -m "v0.1.0 - Initial Release

Production-ready AI agent framework for Go.

Features:
- 3 agent types (Function, ReAct, Conversational)
- OpenAI and Ollama LLM providers
- Tool system with custom tool support
- Memory management strategies
- 100+ tests passing
- Complete documentation

Let's Go, Agents! 🚀"

git push origin v0.1.0

echo ""
echo "✅ Success! Your package is now live on GitHub! 🎉"
echo ""
echo "Repository: https://github.com/yashrahurikar23/goagents"
echo "Documentation will be available at: https://pkg.go.dev/github.com/yashrahurikar23/goagents"
