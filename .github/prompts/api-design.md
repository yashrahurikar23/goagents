---
mode: 'agent'
tools: ['run_in_terminal', 'runTests']
description: 'REST API design patterns and Go Gin framework best practices.'
---

# API Design Guidelines

## REST Principles

### Resource Naming

- Use plural nouns: `/projects`, `/courses`, `/jobs`
- Hierarchical relationships: `/projects/{id}/backers`
- Actions as sub-resources: `/projects/{id}/publish`

### HTTP Methods

- `GET`: Retrieve resources (idempotent)
- `POST`: Create new resources
- `PUT`: Update entire resource (idempotent)
- `PATCH`: Partial updates
- `DELETE`: Remove resources (idempotent)

## Go Gin Patterns

### Handler Structure

```go
func (h *Handler) GetProject(c *gin.Context) {
    id := c.Param("id")
    // Business logic here
    c.JSON(200, project)
}
```

### Middleware Usage

- Authentication middleware for protected routes
- Validation middleware for request bodies
- Logging middleware for all requests
- CORS middleware for cross-origin requests

### Error Responses

- Use appropriate HTTP status codes
- Include error details in response body
- Log errors with context for debugging
- Don't expose internal errors to clients

## Request/Response Patterns

### Request Validation

- Use struct tags for validation rules
- Validate input before processing
- Return 400 for invalid requests
- Provide clear validation error messages

### Response Format

- Consistent JSON structure across endpoints
- Include metadata (pagination, counts)
- Use snake_case for JSON field names
- Handle null values appropriately

