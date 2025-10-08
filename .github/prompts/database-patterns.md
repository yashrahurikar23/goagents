---
mode: 'agent'
tools: ['run_in_terminal', 'dbclient-execute-query']
description: 'Database design patterns and MongoDB best practices.'
---

# Database Guidelines

## MongoDB Patterns

### Atomic Operations

- Use `$inc`, `$push`, `$addToSet` for concurrent safety
- Combine multiple operations in single `UpdateOne` call
- Avoid read-modify-write patterns that cause race conditions

### Indexing Strategy

- Index fields used in filters: `status`, `user_id`, `created_at`
- Compound indexes for common query patterns
- Text indexes for search functionality
- Consider index size vs query performance trade-offs

### Data Modeling

- Embed related data when read together frequently
- Use references for large or frequently changing data
- Store computed fields to avoid expensive aggregations
- Use arrays for one-to-many relationships within documents

## Query Patterns

### Efficient Lookups

- Use `_id` for primary key queries (automatic index)
- Filter by indexed fields to avoid collection scans
- Use projection to return only needed fields
- Limit results for large datasets

### Aggregation Pipeline

- Use for complex data transformations
- Prefer aggregation over multiple queries
- Stage order matters for performance
- Use `$match` early to reduce document processing

