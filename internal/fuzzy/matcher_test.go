package fuzzy

import (
	"strings"
	"testing"
)

// TestMatcher_ExactMatch tests that exact matches are found with 100% confidence
func TestMatcher_ExactMatch(t *testing.T) {
	matcher := NewMatcher(0.10)

	haystack := `package main

func main() {
	fmt.Println("Hello, World!")
}
`

	needle := `func main() {
	fmt.Println("Hello, World!")
}`

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find exact match")
	}

	if result.Confidence != 1.0 {
		t.Errorf("expected confidence 1.0 for exact match, got %f", result.Confidence)
	}

	if result.Index == -1 {
		t.Error("expected valid index")
	}

	if result.LineNumber != 3 {
		t.Errorf("expected line number 3, got %d", result.LineNumber)
	}
}

// TestMatcher_FuzzyMatch tests matching with whitespace differences
func TestMatcher_FuzzyMatch(t *testing.T) {
	matcher := NewMatcher(0.10)

	// Original code (2 spaces indentation)
	haystack := `func calculate(a, b int) int {
  return a + b
}`

	// LLM's version (4 spaces indentation)
	needle := `func calculate(a, b int) int {
    return a + b
}`

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find fuzzy match")
	}

	if result.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", result.Confidence)
	}

	t.Logf("Fuzzy match found with confidence: %f", result.Confidence)
}

// TestMatcher_TabsVsSpaces tests matching code with tabs vs spaces
func TestMatcher_TabsVsSpaces(t *testing.T) {
	matcher := NewMatcher(0.15) // slightly higher tolerance for tabs

	// Original (tabs)
	haystack := "func test() {\n\treturn 42\n}"

	// LLM's version (4 spaces)
	needle := "func test() {\n    return 42\n}"

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find match despite tabs vs spaces")
	}

	t.Logf("Tabs vs spaces match confidence: %f", result.Confidence)
}

// TestMatcher_NoMatch tests that completely different text doesn't match
func TestMatcher_NoMatch(t *testing.T) {
	matcher := NewMatcher(0.10)

	haystack := `func add(a, b int) int {
	return a + b
}`

	needle := `func multiply(x, y int) int {
	return x * y
}`

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Found {
		t.Errorf("expected no match for completely different code, but found match with confidence %f", result.Confidence)
	}
}

// TestMatcher_LineAnchoring tests searching near a specific line number
func TestMatcher_LineAnchoring(t *testing.T) {
	matcher := NewMatcher(0.10)

	haystack := strings.Repeat("// Line filler\n", 100) + // 100 lines of filler
		`func target() {
	return "found me"
}` + strings.Repeat("\n// More filler\n", 100)

	needle := `func target() {
	return "found me"
}`

	// Test with line anchor near the target (line 101)
	result, err := matcher.MatchWithLineAnchor(haystack, needle, 105)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find match with line anchoring")
	}

	if result.LineNumber < 100 || result.LineNumber > 110 {
		t.Errorf("expected line number around 101, got %d", result.LineNumber)
	}

	t.Logf("Found match at line %d with confidence %f", result.LineNumber, result.Confidence)
}

// TestMatcher_MultipleMatches tests finding multiple occurrences
func TestMatcher_MultipleMatches(t *testing.T) {
	matcher := NewMatcher(0.10)

	haystack := `
func test1() {
	return 42
}

func test2() {
	return 42
}

func test3() {
	return 42
}
`

	needle := `return 42`

	results, err := matcher.MatchMultiple(haystack, needle, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 matches, got %d", len(results))
	}

	// Verify matches are non-overlapping and in order
	for i := 1; i < len(results); i++ {
		if results[i].Index <= results[i-1].Index {
			t.Errorf("matches not in order: %d vs %d", results[i].Index, results[i-1].Index)
		}
	}
}

// TestMatcher_EmptyNeedle tests error handling for empty search string
func TestMatcher_EmptyNeedle(t *testing.T) {
	matcher := NewMatcher(0.10)

	_, err := matcher.Match("some text", "")
	if err == nil {
		t.Error("expected error for empty needle")
	}
}

// TestMatcher_EmptyHaystack tests searching in empty string
func TestMatcher_EmptyHaystack(t *testing.T) {
	matcher := NewMatcher(0.10)

	result, err := matcher.Match("", "needle")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Found {
		t.Error("expected no match in empty haystack")
	}
}

// TestMatcher_CaseInsensitive tests case-insensitive matching
func TestMatcher_CaseInsensitive(t *testing.T) {
	matcher := NewMatcher(0.10).WithCaseInsensitive()

	haystack := `func MyFunction() {
	return "HELLO"
}`

	needle := `func myfunction() {
	return "hello"
}`

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find case-insensitive match")
	}

	if result.Confidence < 0.90 {
		t.Errorf("expected high confidence for case-insensitive match, got %f", result.Confidence)
	}
}

// TestMatcher_WindowsLineEndings tests handling of \r\n line endings
func TestMatcher_WindowsLineEndings(t *testing.T) {
	matcher := NewMatcher(0.10)

	// Unix line endings
	haystack := "line1\nline2\nline3\n"

	// Windows line endings
	needle := "line1\r\nline2\r\nline3\r\n"

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find match despite different line endings")
	}

	if result.Confidence != 1.0 {
		t.Errorf("expected exact match after line ending normalization, got confidence %f", result.Confidence)
	}
}

// TestMatcher_Distance tests the Distance method
func TestMatcher_Distance(t *testing.T) {
	matcher := NewMatcher(0.10)

	tests := []struct {
		s1       string
		s2       string
		maxDist  float64
		minDist  float64
		describe string
	}{
		{"hello", "hello", 0.0, 0.0, "identical strings"},
		{"hello", "helo", 0.21, 0.19, "one character different"},  // distance = 0.20 (1 edit out of 5 chars)
		{"hello", "world", 0.81, 0.79, "completely different"},    // distance = 0.80 (4 edits out of 5 chars)
		{"  hello", "hello", 0.29, 0.28, "whitespace difference"}, // distance = 0.285714 (2 edits out of 7 chars)
	}

	for _, tt := range tests {
		distance := matcher.Distance(tt.s1, tt.s2)
		if distance < tt.minDist || distance > tt.maxDist {
			t.Errorf("%s: expected distance between %f and %f, got %f",
				tt.describe, tt.minDist, tt.maxDist, distance)
		}
		t.Logf("%s: distance = %f", tt.describe, distance)
	}
}

// TestMatcher_ToleranceLevels tests different tolerance levels
func TestMatcher_ToleranceLevels(t *testing.T) {
	haystack := "func test() { return 42 }"
	needle := "func test() {  return 42  }" // extra spaces

	tests := []struct {
		tolerance  float64
		shouldFind bool
	}{
		{0.00, false}, // exact match only - should fail
		{0.05, false}, // very strict - should fail
		{0.10, true},  // default - should succeed
		{0.20, true},  // lenient - should succeed
	}

	for _, tt := range tests {
		matcher := NewMatcher(tt.tolerance)
		result, err := matcher.Match(haystack, needle)
		if err != nil {
			t.Fatalf("unexpected error with tolerance %f: %v", tt.tolerance, err)
		}

		if result.Found != tt.shouldFind {
			t.Errorf("tolerance %f: expected found=%v, got found=%v (confidence=%f)",
				tt.tolerance, tt.shouldFind, result.Found, result.Confidence)
		}
	}
}

// TestMatcher_RealWorldCode tests with actual code examples
func TestMatcher_RealWorldCode(t *testing.T) {
	matcher := NewMatcher(0.10)

	// Real Go code
	haystack := `package calculator

import "fmt"

// Calculate performs basic arithmetic
func Calculate(op string, a, b float64) (float64, error) {
    switch op {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("unknown operation: %s", op)
    }
}
`

	// LLM's version with slightly different formatting
	needle := `func Calculate(op string, a, b float64) (float64, error) {
    switch op {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("unknown operation: %s", op)
    }
}`

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find function in real code")
	}

	if result.Confidence < 0.95 {
		t.Errorf("expected high confidence for real code match, got %f", result.Confidence)
	}

	t.Logf("Real code match: line %d, confidence %f", result.LineNumber, result.Confidence)
}

// TestMatcher_LineNumberAccuracy tests that line numbers are accurate
func TestMatcher_LineNumberAccuracy(t *testing.T) {
	matcher := NewMatcher(0.10)

	haystack := `line 1
line 2
line 3
target line
line 5
line 6`

	needle := "target line"

	result, err := matcher.Match(haystack, needle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Found {
		t.Error("expected to find target")
	}

	if result.LineNumber != 4 {
		t.Errorf("expected line number 4, got %d", result.LineNumber)
	}
}

// BenchmarkMatcher_ExactMatch benchmarks exact match performance
func BenchmarkMatcher_ExactMatch(b *testing.B) {
	matcher := NewMatcher(0.10)
	haystack := strings.Repeat("func test() { return 42 }\n", 1000) // 1000 lines
	needle := "func test() { return 42 }"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match(haystack, needle)
	}
}

// BenchmarkMatcher_FuzzyMatch benchmarks fuzzy match performance
func BenchmarkMatcher_FuzzyMatch(b *testing.B) {
	matcher := NewMatcher(0.10)
	haystack := strings.Repeat("func test() { return 42 }\n", 1000)
	needle := "func test() {  return 42  }" // slightly different

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match(haystack, needle)
	}
}

// BenchmarkMatcher_LargeFile benchmarks performance on large files
func BenchmarkMatcher_LargeFile(b *testing.B) {
	matcher := NewMatcher(0.10)
	haystack := strings.Repeat("// comment\nfunc test() { return 42 }\n", 5000) // ~50KB
	needle := "func test() { return 42 }"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match(haystack, needle)
	}
}
