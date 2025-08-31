package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cooper/quiki/test/framework"
)

func main() {
	var (
		testDir   = flag.String("dir", "suites", "directory containing test suites")
		quikiPath = flag.String("quiki", "../quiki", "path to quiki binary")
		verbose   = flag.Bool("v", false, "verbose output")
		filter    = flag.String("filter", "", "filter tests by name")
		suite     = flag.String("suite", "", "run specific suite file")
	)
	flag.Parse()

	// ensure quiki binary exists
	if _, err := os.Stat(*quikiPath); os.IsNotExist(err) {
		fmt.Printf("error: quiki binary not found at %s\n", *quikiPath)
		fmt.Println("hint: run 'go build' first")
		os.Exit(1)
	}

	runner := framework.NewRunner(*quikiPath)
	runner.Verbose = *verbose
	runner.Filter = *filter

	var results []*framework.SuiteResult
	var err error

	if *suite != "" {
		// run single suite
		testSuite, err := runner.LoadSuite(*suite)
		if err != nil {
			fmt.Printf("error: failed to load suite %s: %v\n", *suite, err)
			os.Exit(1)
		}
		result := runner.RunSuite(testSuite)
		results = []*framework.SuiteResult{result}
	} else {
		// run all suites
		results, err = runner.RunAllSuites(*testDir)
		if err != nil {
			fmt.Printf("error: failed to run test suites: %v\n", err)
			os.Exit(1)
		}
	}

	runner.PrintResults(results)

	// exit with error code if any tests failed
	for _, result := range results {
		if result.Failed > 0 {
			os.Exit(1)
		}
	}
}
