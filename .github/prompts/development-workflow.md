---
mode: 'agent'
tools: ['run_in_terminal', 'run_task']
description: 'Guidelines for running and testing the application.'
---

# Development Workflow Guidelines

## Testing Strategy

- Run tests before committing: `npm test` (client) and `go test ./...` (server)
- Integration tests require both services running
- E2E tests use Playwright, run with `npm run test:e2e`

## Build Process

- Client: `pnpm build` in nalanda-client/
- Server: `uv build` in nalanda-content-engine/ or `go build` in nalanda-go-server/
- Docker builds: Use multi-stage for optimized images

## Deployment

- Use environment-specific configs (.env files)
- Database migrations run automatically on startup
- Health checks required before traffic routing
