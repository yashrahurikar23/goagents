---
mode: 'agent'
tools: ['read_file', 'grep_search']
description: 'Error handling patterns and best practices.'
---

# Error Handling Guidelines

## Go Backend Patterns

### Repository Layer

- Return errors as-is from MongoDB operations
- Use `mongo.ErrNoDocuments` for not found cases
- Wrap errors with context: `fmt.Errorf("failed to create user: %w", err)`

### Service Layer

- Validate input before database operations
- Return typed errors for different failure modes
- Use error wrapping for debugging: `errors.Wrap(err, "validation failed")`

### Handler Layer

- Map errors to appropriate HTTP status codes
- Return JSON error responses with consistent structure
- Log errors with context for debugging

## Frontend Patterns

### API Calls

- Handle network errors gracefully
- Show user-friendly error messages
- Retry failed requests with exponential backoff

### State Management

- Handle async operation failures
- Provide loading and error states in UI
- Clear errors when operations succeed
