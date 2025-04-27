package metrics

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// Metrics represents code metrics for a file
type Metrics struct {
	LOC           int // Lines of code
	CC            int // Cyclomatic complexity
	CogC          int // Cognitive complexity
	FuncCount     int // Total number of functions
	TestPassCount int // Number of functions that passed tests
}

// CalculateMetrics calculates all metrics for a given file
func CalculateMetrics(filePath string) (*Metrics, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the Go code
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	metrics := &Metrics{}

	// Calculate LOC
	metrics.LOC = calculateLOC(string(content))

	// Calculate cyclomatic complexity
	metrics.CC = calculateCyclomaticComplexity(f)

	// Calculate cognitive complexity
	metrics.CogC = calculateCognitiveComplexity(f)

	// Count functions
	metrics.FuncCount = countFunctions(f)

	return metrics, nil
}

// calculateLOC calculates the number of lines of code
func calculateLOC(content string) int {
	scanner := bufio.NewScanner(strings.NewReader(content))
	loc := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*") {
			loc++
		}
	}
	return loc
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity
func calculateCyclomaticComplexity(f *ast.File) int {
	v := &cyclomaticVisitor{complexity: 1}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			v.complexity++
		case *ast.ForStmt, *ast.RangeStmt:
			v.complexity++
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				v.complexity++
			}
		case *ast.SwitchStmt:
			if node.Body != nil {
				v.complexity += len(node.Body.List) - 1
				if len(node.Body.List) == 0 {
					v.complexity++
				}
			}
		case *ast.TypeSwitchStmt:
			if node.Body != nil {
				v.complexity += len(node.Body.List) - 1
				if len(node.Body.List) == 0 {
					v.complexity++
				}
			}
		case *ast.SelectStmt:
			if node.Body != nil {
				v.complexity += len(node.Body.List) - 1
				if len(node.Body.List) == 0 {
					v.complexity++
				}
			}
		}
		return true
	})

	return v.complexity
}

type cyclomaticVisitor struct {
	complexity int
}

// calculateCognitiveComplexity calculates the cognitive complexity
func calculateCognitiveComplexity(f *ast.File) int {
	complexity := 0
	nestingLevels := make(map[ast.Node]int)

	nodeStack := []ast.Node{}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt,
			*ast.TypeSwitchStmt, *ast.SelectStmt:
			level := 0
			if len(nodeStack) > 0 {
				parentNode := nodeStack[len(nodeStack)-1]
				level = nestingLevels[parentNode] + 1
			}
			nestingLevels[node] = level
			nodeStack = append(nodeStack, node)

		case *ast.BlockStmt:
			if len(nodeStack) > 0 && isControlFlowParent(nodeStack[len(nodeStack)-1], node) {
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
		}
		return true
	})

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++
			complexity += nestingLevels[node]

			if node.Else != nil {
				if _, isIf := node.Else.(*ast.IfStmt); isIf {
					complexity++
				} else {
					complexity++
				}
			}

		case *ast.ForStmt, *ast.RangeStmt:
			complexity++
			complexity += nestingLevels[node]

		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
			complexity += nestingLevels[node]

			var caseCount int
			if s, ok := node.(*ast.SwitchStmt); ok && s.Body != nil {
				caseCount = len(s.Body.List)
			} else if ts, ok := node.(*ast.TypeSwitchStmt); ok && ts.Body != nil {
				caseCount = len(ts.Body.List)
			}

			if caseCount > 0 {
				complexity += caseCount - 1
			}

		case *ast.SelectStmt:
			complexity++
			complexity += nestingLevels[node]

			if node.Body != nil {
				if len(node.Body.List) > 0 {
					complexity += len(node.Body.List) - 1
				}
			}

		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}

		case *ast.BranchStmt:
			if node.Tok == token.BREAK || node.Tok == token.CONTINUE || node.Tok == token.GOTO {
				complexity++
			}

		case *ast.ExprStmt:
			if call, ok := node.X.(*ast.CallExpr); ok {
				if fun, ok := call.Fun.(*ast.FuncLit); ok && fun.Body != nil {
					complexity++
				}
			}
		}
		return true
	})

	return complexity
}

// Проверяет, является ли blockNode блоком управляющей конструкции parentNode
func isControlFlowParent(parentNode ast.Node, blockNode *ast.BlockStmt) bool {
	switch node := parentNode.(type) {
	case *ast.IfStmt:
		return node.Body == blockNode
	case *ast.ForStmt:
		return node.Body == blockNode
	case *ast.RangeStmt:
		return node.Body == blockNode
	case *ast.SwitchStmt:
		return node.Body == blockNode
	case *ast.TypeSwitchStmt:
		return node.Body == blockNode
	case *ast.SelectStmt:
		return node.Body == blockNode
	}
	return false
}

// countFunctions counts the total number of functions in a file
func countFunctions(f *ast.File) int {
	count := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncDecl); ok {
			count++
		}
		return true
	})
	return count
}

// CalculateFunctionalEquivalence calculates the functional equivalence metric
func CalculateFunctionalEquivalence(passedTests, totalTests int) float64 {
	if totalTests == 0 {
		return 0
	}
	return float64(passedTests) / float64(totalTests) * 100
}

// CalculateDeltaMetrics calculates the delta metrics between original and metamorphic code
func CalculateDeltaMetrics(original, metamorphic *Metrics) (float64, float64, float64) {
	locDelta := float64(metamorphic.LOC-original.LOC) / float64(original.LOC) * 100
	ccDelta := float64(metamorphic.CC-original.CC) / float64(original.CC) * 100
	cogCDelta := float64(metamorphic.CogC-original.CogC) / float64(original.CogC) * 100
	return locDelta, ccDelta, cogCDelta
}
