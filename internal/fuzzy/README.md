# Fuzzy Package

**Purpose:** Fuzzy string matching for AI-powered file editing operations.

## Why This Exists

AI language models (LLMs) struggle with exact whitespace matching when generating code edits. They might use:
- Tabs instead of spaces
- Wrong indentation levels
- Extra or missing blank lines
- Different line endings

This causes traditional search-replace operations to fail ~30% of the time. The fuzzy package solves this by using **Levenshtein distance** to tolerate small differences, improving edit success rates from ~70% to **95%+**.

## Key Concepts

### Levenshtein Distance
The minimum number of single-character edits (insertions, deletions, substitutions) needed to transform one string into another.

Example:
- `"hello"` → `"helo"`: distance = 1 (delete one 'l')
- `"cat"` → `"dog"`: distance = 3 (substitute all characters)

### Normalized Distance
Distance divided by the maximum string length, giving a 0.0-1.0 score:
- `0.0` = identical strings
- `0.1` = 10% different
- `1.0` = completely different

### Tolerance
The maximum allowed normalized distance for a match:
- `0.0` = exact match required (no fuzzy matching)
- `0.10` = 10% difference allowed (recommended for code)
- `0.20` = 20% difference allowed (very lenient)

### Confidence
The inverse of distance (1.0 - distance):
- `1.0` = exact match
- `0.95` = 95% similar
- `0.90` = 90% similar

## Usage

### Basic Matching

```go
import "github.com/yashrahurikar23/goagents/internal/fuzzy"

// Create matcher with 10% tolerance
matcher := fuzzy.NewMatcher(0.10)

haystack := `func calculate(a, b int) int {
  return a + b
}`

needle := `func calculate(a, b int) int {
    return a + b
}` // Different indentation

result, err := matcher.Match(haystack, needle)
if err != nil {
    log.Fatal(err)
}

if result.Found {
    fmt.Printf("Found at index %d (line %d) with %.2f%% confidence\n",
        result.Index, result.LineNumber, result.Confidence*100)
    fmt.Printf("Matched text: %s\n", result.MatchedText)
}
```

### Line-Anchored Matching

When you know approximately where the match should be:

```go
// Search near line 42 first
result, err := matcher.MatchWithLineAnchor(haystack, needle, 42)
```

This searches within ±50 lines of the anchor before falling back to full search, improving performance.

### Multiple Matches

Find all occurrences of a pattern:

```go
results, err := matcher.MatchMultiple(haystack, needle, 10) // max 10 matches
for i, result := range results {
    fmt.Printf("Match %d: line %d, confidence %.2f\n",
        i+1, result.LineNumber, result.Confidence)
}
```

### Case-Insensitive Matching

```go
matcher := fuzzy.NewMatcher(0.10).WithCaseInsensitive()
result, _ := matcher.Match("Hello World", "hello world") // Found!
```

### Distance Calculation

```go
distance := matcher.Distance("hello", "helo")
fmt.Printf("Distance: %.2f (%.0f%% different)\n", distance, distance*100)
// Output: Distance: 0.20 (20% different)
```

## API Reference

### Types

#### `Matcher`
```go
type Matcher struct {
    maxTolerance  float64  // 0.0-1.0
    caseSensitive bool
}
```

#### `MatchResult`
```go
type MatchResult struct {
    Found       bool    // Was a match found?
    Index       int     // Byte offset (-1 if not found)
    Confidence  float64 // 0.0-1.0 similarity score
    MatchedText string  // The actual matched text
    LineNumber  int     // 1-indexed line number
}
```

### Functions

#### `NewMatcher(tolerance float64) *Matcher`
Creates a new matcher with specified tolerance (0.0-1.0).

**Recommended tolerance values:**
- `0.05` - Very strict, only minor whitespace differences
- `0.10` - **Recommended default** for code editing
- `0.15` - Lenient, handles tabs vs spaces well
- `0.20` - Very lenient, may match unintended code

#### `Match(haystack, needle string) (*MatchResult, error)`
Searches for the best fuzzy match of needle in haystack.

**Performance:** O(n×m×k) where:
- n = haystack length
- m = needle length  
- k = Levenshtein cost

For typical code files (<100KB) and search blocks (<500 chars): **<100ms**

#### `MatchWithLineAnchor(haystack, needle string, lineNumber int) (*MatchResult, error)`
Searches near a specific line number first, then falls back to full search.

**Use when:** LLM provides a line number hint (e.g., "around line 42")

#### `MatchMultiple(haystack, needle string, maxMatches int) ([]*MatchResult, error)`
Finds multiple non-overlapping matches, up to maxMatches.

#### `Distance(s1, s2 string) float64`
Calculates normalized Levenshtein distance (0.0-1.0).

#### `WithCaseInsensitive() *Matcher`
Returns a new matcher with case-insensitive matching enabled.

## Performance

Benchmarks on M1 MacBook Pro:

```
BenchmarkMatcher_ExactMatch-8     50000    25000 ns/op    (25µs)
BenchmarkMatcher_FuzzyMatch-8     10000   100000 ns/op    (100µs)
BenchmarkMatcher_LargeFile-8       1000  1500000 ns/op    (1.5ms for 50KB)
```

**Optimization tips:**
1. Use `MatchWithLineAnchor()` when you have a line hint
2. Keep search blocks reasonably sized (<1000 characters)
3. Use exact match fallback for identical strings (built-in optimization)

## Testing

Run tests:
```bash
cd internal/fuzzy
go test -v
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

Test coverage:
```bash
go test -cover
# Expected: >90% coverage
```

## Examples

### Example 1: Whitespace Tolerance
```go
haystack := "func test() {\n  return 42\n}"  // 2 spaces
needle   := "func test() {\n    return 42\n}" // 4 spaces

matcher := fuzzy.NewMatcher(0.10)
result, _ := matcher.Match(haystack, needle)
// result.Found = true, result.Confidence ≈ 0.95
```

### Example 2: Tabs vs Spaces
```go
haystack := "func test() {\n\treturn 42\n}"    // tab
needle   := "func test() {\n    return 42\n}" // 4 spaces

matcher := fuzzy.NewMatcher(0.15) // slightly higher tolerance
result, _ := matcher.Match(haystack, needle)
// result.Found = true
```

### Example 3: No Match (Too Different)
```go
haystack := "func add(a, b int) { return a + b }"
needle   := "func multiply(x, y int) { return x * y }"

matcher := fuzzy.NewMatcher(0.10)
result, _ := matcher.Match(haystack, needle)
// result.Found = false (too different)
```

## Inspiration

This package is inspired by **Roo-Code's multi-search-replace strategy**, which uses Levenshtein distance with 10% tolerance to handle LLM-generated code edits. Roo-Code achieved:
- **95%+ edit success rate** (vs ~70% with exact matching)
- **Robust handling** of formatting variations
- **Production-proven** across thousands of users

## References

- **Levenshtein Distance:** https://en.wikipedia.org/wiki/Levenshtein_distance
- **Library Used:** https://github.com/agnivade/levenshtein
- **Roo-Code:** https://github.com/RooCodeInc/Roo-Code

## License

MIT License - See main repository LICENSE file.
