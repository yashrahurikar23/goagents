---
mode: 'agent'
tools: ['githubRepo', 'codebase']
description: 'Guidelines to add code comments.'
---

# Code Comments Guidelines

## Focus: WHY, Not WHAT

**Primary Rule**: Explain WHY code exists, not WHAT it does.

- ❌ Bad: `// Increment counter by 1`
- ✅ Good: `// WHY: Prevent race conditions in concurrent backing operations`

- ❌ Bad: `// Check if user is admin`
- ✅ Good: `// WHY: Only admins can modify system constants to maintain data integrity`

## Module-Level Documentation

At the start of each module (file), include comprehensive documentation:

```go
/*
Package repository provides the data access layer for [system name].

PURPOSE:
[What this module handles and why it's needed in the system]

WHY THIS EXISTS:
[Business rationale for separate system/module]

KEY DESIGN DECISIONS:
- [Decision 1 with explanation]
- [Decision 2 with explanation]

METHODS:
- Method1: [Brief description of purpose]
- Method2: [Brief description of purpose]
*/
```

## Method-Level Documentation

For each method/function, explain the rationale:

```go
// MethodName [brief what it does]
//
// WHY THIS WAY:
// - Reason 1 explaining design choice
// - Reason 2 explaining implementation approach
//
// BUSINESS LOGIC:
// - Rule 1 that drives the behavior
// - Rule 2 that constrains the operation
//
// WHEN TO USE:
// - Use case 1
// - Use case 2
func MethodName() {
    // WHY: [Inline explanation for complex logic]
    // TODO: [Future improvements or known limitations]
}
```

## Documentation Patterns

### Atomic Operations

```go
// WHY: Use $push and $inc together to ensure consistency
// If one fails, both fail - prevents partial state
```

### Business Rules

```go
// WHY: Auto-close expired campaigns to prevent stale data
// Business rule: campaigns can't accept backing after end date
```

### Performance Decisions

```go
// WHY: Store computed field instead of calculating on read
// Enables efficient filtering/sorting without expensive operations
```

### Error Handling

```go
// WHY: Return specific error types for proper HTTP status codes
// Allows caller to distinguish between not found vs server error
```

## What to Avoid

- Repetitive comments that just restate the code
- Obvious explanations (`i++ // increment i`)
- Implementation details that change frequently
- Comments that become outdated with code changes

## Quality Checklist

- [ ] Every module has comprehensive header documentation
- [ ] Each method explains WHY the approach was chosen
- [ ] Business rules are clearly documented
- [ ] Performance trade-offs are explained
- [ ] Future maintainers can understand the reasoning
