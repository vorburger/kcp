package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kcp-dev/kcp/pkg/cmd/help"
	"github.com/kcp-dev/kcp/pkg/kcp"
)

var (
	listen string
)

func main() {
	help.FitTerminal()
	cmd := &cobra.Command{
		Use:   "kcp-core",
		Short: "Kube for Control Plane (KCP)",
		Long: help.Doc(`
			KCP Core (Mini) is the easiest way to launch a generic Kubernetes Control Plane API server.
			
			To get started, launch a new cluster with 'kcp start', which will
			initialize your personal control plane and write an admin kubeconfig file
			to disk.
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the control plane process",
		Long: help.Doc(`
			Start the control plane process

			The server process listens on port 6443 and will act like a Kubernetes
			API server. It will initialize any necessary data to the provided start
			location or as a '.kcp' directory in the current directory. An admin
			kubeconfig file will be generated at initialization time that may be
			used to access the control plane.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			//flag.CommandLine.Lookup("v").Value.Set("9")
			kcp := &kcp.KCP{
				Listen: listen,
			}
			return kcp.Run()
		},
	}
	startCmd.Flags().AddFlag(pflag.PFlagFromGoFlag(flag.CommandLine.Lookup("v")))
	startCmd.Flags().StringVar(&listen, "listen", ":6443", "Address:port to bind to")
	cmd.AddCommand(startCmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
