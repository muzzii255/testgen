package cmd

import (
	"fmt"
	"net/http"

	"github.com/muzzii255/testgen/proxy"

	"github.com/spf13/cobra"
)

var (
	port   int
	target int
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record test scenarios for replay",
	Long:  `Records user interactions and system behaviors to create reusable test scenarios.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetURL := fmt.Sprintf("http://localhost:%d", target)
		recorder, _ := proxy.NewRecorder(targetURL, "./recordings")
		fmt.Println("piece of shit recording on :", port)
		fmt.Println("target :", target)
		addr := fmt.Sprintf(":%d", port)
		http.ListenAndServe(addr, recorder)
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)

	recordCmd.Flags().IntVarP(&port, "port", "p", 9000, "Port to run proxy on")
	recordCmd.Flags().IntVarP(&target, "target", "t", 8080, "Target backend URL")
}
