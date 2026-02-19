package cmd

import (
	"log/slog"
	"os"

	"github.com/muzzii255/testgen/generator"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate tests for the project",
	Long: `Analyzes your codebase and automatically generates comprehensive test suites.
Supports integration tests.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			slog.Error("error fetching cwd", "err", err)
			return
		}
		fileLoc, err := cmd.Flags().GetString("file")
		if err != nil {
			slog.Error("error parsing file flag", "err", err)
			return
		}
		jsonFile := generator.JsonFile{
			Filename: fileLoc,
			BaseDir:  cwd,
		}
		err = jsonFile.ReadFile()
		if err != nil {
			slog.Error("error reading file", "err", err)
			return

		}
		err = jsonFile.GenTest()
		if err != nil {
			slog.Error("error generating test files", "err", err)
			return

		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("file", "f", "", "Path to the recorded JSON file used to generate test cases.")
	generateCmd.MarkFlagRequired("file")
}
