/*
Copyright © 2026 muzzii255 muzammil.jvd@gmail.com
*/
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.testgen.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
