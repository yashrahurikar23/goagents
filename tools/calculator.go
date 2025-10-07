package tools

import (
	"context"
	"fmt"

	"github.com/yashrahurikar23/goagents/core"
)

// Calculator is a simple calculator tool that performs basic arithmetic operations
type Calculator struct{}

// NewCalculator creates a new calculator tool
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Name returns the tool's name
func (c *Calculator) Name() string {
	return "calculator"
}

// Description returns the tool's description
func (c *Calculator) Description() string {
	return "A calculator tool that can perform basic arithmetic operations: add, subtract, multiply, divide. Use this when you need to perform calculations."
}

// Schema returns the tool's parameter schema
func (c *Calculator) Schema() *core.ToolSchema {
	return &core.ToolSchema{
		Name:        "calculator",
		Description: "Performs basic arithmetic operations",
		Parameters: []core.Parameter{
			{
				Name:        "operation",
				Type:        "string",
				Description: "The operation to perform: add, subtract, multiply, divide",
				Required:    true,
				Enum:        []interface{}{"add", "subtract", "multiply", "divide"},
			},
			{
				Name:        "a",
				Type:        "number",
				Description: "The first number",
				Required:    true,
			},
			{
				Name:        "b",
				Type:        "number",
				Description: "The second number",
				Required:    true,
			},
		},
	}
}

// Execute performs the calculation
func (c *Calculator) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Extract operation
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	// Extract numbers
	aVal, ok := args["a"]
	if !ok {
		return nil, fmt.Errorf("parameter 'a' is required")
	}

	bVal, ok := args["b"]
	if !ok {
		return nil, fmt.Errorf("parameter 'b' is required")
	}

	// Convert to float64
	var a, b float64
	switch v := aVal.(type) {
	case float64:
		a = v
	case int:
		a = float64(v)
	case int64:
		a = float64(v)
	default:
		return nil, fmt.Errorf("parameter 'a' must be a number")
	}

	switch v := bVal.(type) {
	case float64:
		b = v
	case int:
		b = float64(v)
	case int64:
		b = float64(v)
	default:
		return nil, fmt.Errorf("parameter 'b' must be a number")
	}

	// Perform operation
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return result, nil
}
