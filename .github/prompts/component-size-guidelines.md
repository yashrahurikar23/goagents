# Component Size and Architecture Guidelines

## Overview

This document outlines guidelines for component size, architecture, and organization to ensure maintainable, readable, and scalable React components.

## Component Size Limits

### Maximum Lines of Code (LOC)

- **Functional Components**: 200-250 lines maximum
- **Class Components**: 300 lines maximum (discouraged, prefer functional)
- **Custom Hooks**: 150 lines maximum
- **Utility Functions**: 100 lines maximum

### Rationale

- Components exceeding these limits become difficult to understand and maintain
- Large components often violate Single Responsibility Principle
- Breaking down large components improves testability and reusability

## Component Organization

### File Structure

```text
src/components/
├── feature-name/
│   ├── ComponentName.tsx          # Main component
│   ├── ComponentName.test.tsx     # Unit tests
│   ├── subcomponent/
│   │   ├── SubComponent.tsx       # Sub-components
│   │   └── index.ts              # Clean exports
│   ├── hooks/
│   │   └── useComponentLogic.ts   # Custom hooks
│   ├── types.ts                  # TypeScript types
│   └── index.ts                  # Main exports
```

### Component Types

#### 1. Page Components (100-200 LOC)

- Route-level components
- Coordinate data fetching and state management
- Compose smaller components
- Focus on orchestration, not implementation

#### 2. Feature Components (50-150 LOC)

- Self-contained features
- Handle specific business logic
- May contain sub-components
- Reusable within the application

#### 3. UI Components (20-80 LOC)

- Pure presentation components
- Focus on styling and user interaction
- Highly reusable
- Minimal or no business logic

#### 4. Layout Components (30-100 LOC)

- Structure and positioning
- Responsive design handling
- Container components

## Refactoring Triggers

### When to Break Down a Component

1. **Exceeds size limits** (200-250 LOC)
2. **Multiple responsibilities** (violates SRP)
3. **Complex state management** (>5 state variables)
4. **Long render methods** (>50 lines)
5. **Deep nesting** (>3 levels)
6. **Multiple conditional renders** (>5 branches)

### Refactoring Patterns

#### 1. Extract Custom Hooks

```typescript
// Before: Component with complex logic
function LargeComponent() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Complex data fetching logic
  }, []);

  // Component logic...
}

// After: Extract to custom hook
function useDataFetching() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Complex data fetching logic
  }, []);

  return { data, loading };
}

function LargeComponent() {
  const { data, loading } = useDataFetching();
  // Component logic...
}
```

#### 2. Extract Sub-components

```typescript
// Before: Large component with multiple sections
function Dashboard() {
  return (
    <div>
      <div className="header">...</div>
      <div className="stats">...</div>
      <div className="chart">...</div>
      <div className="table">...</div>
    </div>
  );
}

// After: Extract sub-components
function DashboardHeader() { /* ... */ }
function DashboardStats() { /* ... */ }
function DashboardChart() { /* ... */ }
function DashboardTable() { /* ... */ }

function Dashboard() {
  return (
    <div>
      <DashboardHeader />
      <DashboardStats />
      <DashboardChart />
      <DashboardTable />
    </div>
  );
}
```

#### 3. Extract Configuration Objects

```typescript
// Before: Inline configuration
const statusColors = {
  running: 'green',
  failed: 'red',
  pending: 'yellow'
};

// After: Extract to separate file
export const STATUS_CONFIG = {
  running: { color: 'green', label: 'Running' },
  failed: { color: 'red', label: 'Failed' },
  pending: { color: 'yellow', label: 'Pending' }
} as const;
```

## State Management Guidelines

### Local State vs Global State

- **Local State**: Component-specific state (< 5 variables)
- **Global State**: Shared state across components
- **Server State**: Data from APIs (use React Query/SWR)

### State Organization

- Group related state variables
- Use reducer for complex state transitions
- Extract stateful logic to custom hooks

## Performance Considerations

### Component Splitting Benefits

- **Code Splitting**: Smaller bundles
- **Tree Shaking**: Unused code elimination
- **Caching**: Better memoization opportunities
- **Testing**: Isolated unit tests

### Memoization Guidelines

- Use `React.memo()` for expensive renders
- Use `useMemo()` for expensive calculations
- Use `useCallback()` for event handlers passed to children

## Testing Guidelines

### Component Testing

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **E2E Tests**: Test complete user flows

### Test File Organization

```
### Test File Organization

```text
ComponentName.test.tsx
__tests__/
  ├── unit/
  ├── integration/
  └── e2e/
```
```

## Code Quality Metrics

### Maintainability Index

- Target: > 70
- Measures: Complexity, LOC, Halstead metrics

### Cyclomatic Complexity

- Target: < 10 per function
- Measures: Decision points in code

### Code Coverage

- Target: > 80%
- Measures: Lines/functions/branches covered by tests

## Migration Strategy

### Gradual Refactoring

1. Identify large components (>250 LOC)
2. Analyze responsibilities and dependencies
3. Extract custom hooks first
4. Extract sub-components
5. Update imports and tests
6. Verify functionality

### Breaking Changes

- Update component APIs gradually
- Use deprecation warnings
- Maintain backward compatibility during migration

## Tools and Automation

### Linting Rules

```json
### Linting Rules

```json
{
  "rules": {
    "max-lines-per-function": ["error", 50],
    "complexity": ["error", 10],
    "max-depth": ["error", 3]
  }
}
```
```

### Automated Checks

- Pre-commit hooks for size limits
- CI/CD checks for complexity metrics
- Bundle size monitoring

## Examples

### Good Component Structure

```typescript
// Page Component (orchestration)
function ProjectDetailsPage() {
  const { project, loading } = useProjectDetails();
  const { services } = useProjectServices();

  if (loading) return <LoadingSpinner />;

  return (
    <div>
      <ProjectHeader project={project} />
      <ProjectTabs>
        <ProjectOverview services={services} />
        <ProjectDeployments />
      </ProjectTabs>
    </div>
  );
}

// Feature Component (business logic)
function ProjectOverview({ services }) {
  return (
    <div>
      <EnvironmentAnalytics />
      <ServicesList services={services} />
    </div>
  );
}

// UI Component (presentation)
function ServiceCard({ service }) {
  return (
    <Card>
      <CardTitle>{service.name}</CardTitle>
      <Badge variant={service.status}>
        {service.status}
      </Badge>
    </Card>
  );
}
```

## Conclusion

Following these guidelines ensures:

- **Maintainable Code**: Easier to understand and modify
- **Better Performance**: Smaller components, better memoization
- **Improved Testing**: Isolated, focused tests
- **Enhanced Reusability**: Modular, composable components
- **Scalable Architecture**: Clear separation of concerns

Regular code reviews and automated checks help maintain these standards across the codebase.
