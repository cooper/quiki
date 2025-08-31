package framework

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestCase represents a single formatter test
type TestCase struct {
	Name        string   `json:"name"`
	Input       string   `json:"input"`
	Expected    string   `json:"expected"`
	Description string   `json:"description"`
	Skip        bool     `json:"skip,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// TestSuite groups related test cases
type TestSuite struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Tests       []TestCase `json:"tests"`
}

// TestResult holds the outcome of a test
type TestResult struct {
	Name     string
	Passed   bool
	Expected string
	Actual   string
	Error    string
	Duration time.Duration
}

// SuiteResult holds results for an entire suite
type SuiteResult struct {
	Name     string
	Results  []TestResult
	Duration time.Duration
	Passed   int
	Failed   int
	Skipped  int
}

// Runner executes test suites
type Runner struct {
	QuikiPath string
	Verbose   bool
	Filter    string
}

// NewRunner creates a test runner
func NewRunner(quikiPath string) *Runner {
	return &Runner{
		QuikiPath: quikiPath,
	}
}

// LoadSuite loads a test suite from json file
func (r *Runner) LoadSuite(path string) (*TestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &suite, nil
}

// RunSuite executes all tests in a suite
func (r *Runner) RunSuite(suite *TestSuite) *SuiteResult {
	start := time.Now()
	result := &SuiteResult{
		Name:    suite.Name,
		Results: make([]TestResult, 0, len(suite.Tests)),
	}

	for _, test := range suite.Tests {
		if test.Skip {
			result.Skipped++
			continue
		}

		if r.Filter != "" && !strings.Contains(test.Name, r.Filter) {
			result.Skipped++
			continue
		}

		testResult := r.runTest(test)
		result.Results = append(result.Results, testResult)

		if testResult.Passed {
			result.Passed++
		} else {
			result.Failed++
		}
	}

	result.Duration = time.Since(start)
	return result
}

// runTest executes a single test case
func (r *Runner) runTest(test TestCase) TestResult {
	start := time.Now()
	result := TestResult{
		Name:     test.Name,
		Expected: test.Expected,
		Duration: time.Since(start),
	}

	// run quiki with the input
	cmd := exec.Command(r.QuikiPath, "-i")
	cmd.Stdin = strings.NewReader(test.Input)

	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("command failed: %v", err)
		return result
	}

	// extract just the content from the html output
	actual := r.extractContent(string(output))
	result.Actual = actual
	result.Passed = (actual == test.Expected)
	result.Duration = time.Since(start)

	return result
}

// extractContent pulls the actual content from quiki's html output
func (r *Runner) extractContent(html string) string {
	// find content between <p class="q-p"> and </p>
	start := strings.Index(html, `<p class="q-p">`)
	if start == -1 {
		return strings.TrimSpace(html)
	}
	start += len(`<p class="q-p">`)

	end := strings.Index(html[start:], "</p>")
	if end == -1 {
		return strings.TrimSpace(html[start:])
	}

	content := html[start : start+end]
	// clean up whitespace
	content = strings.TrimSpace(content)
	// normalize multiple spaces/newlines
	content = strings.ReplaceAll(content, "\n            ", "")
	content = strings.ReplaceAll(content, "\n        ", "")
	content = strings.ReplaceAll(content, "\n    ", "")

	return content
}

// RunAllSuites finds and runs all test suites in a directory
func (r *Runner) RunAllSuites(testDir string) ([]*SuiteResult, error) {
	pattern := filepath.Join(testDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("finding test files: %w", err)
	}

	var results []*SuiteResult
	for _, file := range files {
		suite, err := r.LoadSuite(file)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", file, err)
		}

		result := r.RunSuite(suite)
		results = append(results, result)
	}

	return results, nil
}

// PrintResults outputs test results in a clean format
func (r *Runner) PrintResults(results []*SuiteResult) {
	totalPassed := 0
	totalFailed := 0
	totalSkipped := 0

	fmt.Printf("quiki test suite results\n\n")

	for _, suite := range results {
		if suite.Failed > 0 {
			fmt.Printf("FAIL %s: %d passed, %d failed", suite.Name, suite.Passed, suite.Failed)
		} else {
			fmt.Printf("PASS %s: %d passed", suite.Name, suite.Passed)
		}

		if suite.Skipped > 0 {
			fmt.Printf(", %d skipped", suite.Skipped)
		}
		fmt.Printf(" (%.1fms)\n", float64(suite.Duration.Nanoseconds())/1e6)

		if r.Verbose || suite.Failed > 0 {
			for _, test := range suite.Results {
				if test.Passed {
					if r.Verbose {
						fmt.Printf("  PASS %s\n", test.Name)
					}
				} else {
					fmt.Printf("  FAIL %s\n", test.Name)
					fmt.Printf("    expected: %q\n", test.Expected)
					fmt.Printf("    actual:   %q\n", test.Actual)
					if test.Error != "" {
						fmt.Printf("    error:    %s\n", test.Error)
					}
				}
			}
		}
		fmt.Println()

		totalPassed += suite.Passed
		totalFailed += suite.Failed
		totalSkipped += suite.Skipped
	}

	if totalFailed == 0 {
		fmt.Printf("SUCCESS: all %d tests passed", totalPassed)
	} else {
		fmt.Printf("FAILED: %d failed, %d passed", totalFailed, totalPassed)
	}

	if totalSkipped > 0 {
		fmt.Printf(", %d skipped", totalSkipped)
	}
	fmt.Println()
}
