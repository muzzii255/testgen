package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/muzzii255/testgen/proxy"

	"github.com/spf13/cobra"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record test scenarios for replay",
	Long:  `Records user interactions and system behaviors to create reusable test scenarios.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		target, _ := cmd.Flags().GetInt("target")
		targetURL := fmt.Sprintf("http://localhost:%d", target)
		recorder, err := proxy.NewRecorder(targetURL, "./recordings")
		if err != nil {
			slog.Error("failed to create recorder", "err", err)
			return
		}
		slog.Info("proxy recording", "port", port, "target", target)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: recorder}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", "err", err)
			}
		}()

		<-stop
		slog.Info("shutting down, saving recordings...")
		srv.Shutdown(context.Background())
		recorder.Save()
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
	recordCmd.Flags().IntP("port", "p", 9000, "Port to run proxy on")
	recordCmd.Flags().IntP("target", "t", 8080, "Target backend URL")
}
