/*
Copyright © 2026 NAME HERE muzammil.jvd@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/muzzii255/testgen/generator"

	"github.com/spf13/cobra"
)

var fileLoc string

var generateCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate tests for the project",
	Long: `Analyzes your codebase and automatically generates comprehensive test suites.
Supports integration tests.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Println(cwd)
		jsonFile := generator.JsonFile{
			Filename: fileLoc,
			BaseDir:  cwd,
		}
		err = jsonFile.ReadFile()
		if err != nil {
			panic(err)
		}
		err = jsonFile.GenTest()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&fileLoc, "file", "f", "", "Path to the recorded JSON file used to generate test cases.")
	generateCmd.MarkFlagRequired("file")
}
