package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"dev.azure.com/msazure/One/_git/symphony/gitops/internal/server"
	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleInterupt(cancel)
	newCmdGitOps().ExecuteContext(ctx)
}

func handleInterupt(cancel context.CancelFunc) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(c)

	go func() {
		<-c
		cancel()
	}()
}

func newCmdGitOps() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitops",
		Short: "GitOps",
		Long:  "GitOps",
		Run:   run,
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) {
	server.NewServer().Start()
}
