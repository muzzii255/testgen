package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "testgen",
	Short: "Test generation and recording CLI tool",
	Long: `testgen is a CLI tool for generating and recording tests.

It helps developers create comprehensive test suites quickly.
The tool supports generating tests from code analysis and recording
test scenarios for replay.

Examples:
  testgen generate          # Generate tests for the current project
  testgen record           # Record test scenarios`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
