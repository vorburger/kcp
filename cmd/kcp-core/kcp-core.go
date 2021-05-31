package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.etcd.io/etcd/clientv3"

	"github.com/kcp-dev/kcp/pkg/cmd/help"
	"github.com/kcp-dev/kcp/pkg/etcd"

	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/controlplane"
	"k8s.io/kubernetes/pkg/controlplane/options"
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

			dir := filepath.Join(".", ".kcp")
			if fi, err := os.Stat(dir); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
				if err := os.Mkdir(dir, 0755); err != nil {
					return err
				}
			} else {
				if !fi.IsDir() {
					return fmt.Errorf("%q is a file, please delete or select another location", dir)
				}
			}
			s := &etcd.Server{
				Dir: filepath.Join(dir, "data"),
			}
			ctx := context.TODO()

			return s.Run(func(cfg etcd.ClientInfo) error {
				c, err := clientv3.New(clientv3.Config{
					Endpoints: cfg.Endpoints,
					TLS:       cfg.TLS,
				})
				if err != nil {
					return err
				}
				defer c.Close()
				r, err := c.Cluster.MemberList(context.Background())
				if err != nil {
					return err
				}
				for _, member := range r.Members {
					fmt.Fprintf(os.Stderr, "Connected to etcd %d %s\n", member.GetID(), member.GetName())
				}

				serverOptions := options.NewServerRunOptions()
				host, port, err := net.SplitHostPort(listen)
				if err != nil {
					return fmt.Errorf("--listen must be of format host:port: %w", err)
				}

				if host != "" {
					serverOptions.SecureServing.BindAddress = net.ParseIP(host)
				}
				if port != "" {
					p, err := strconv.Atoi(port)
					if err != nil {
						return err
					}
					serverOptions.SecureServing.BindPort = p
				}

				serverOptions.SecureServing.ServerCert.CertDirectory = s.Dir
				serverOptions.InsecureServing = nil
				serverOptions.Etcd.StorageConfig.Transport = storagebackend.TransportConfig{
					ServerList:    cfg.Endpoints,
					CertFile:      cfg.CertFile,
					KeyFile:       cfg.KeyFile,
					TrustedCAFile: cfg.TrustedCAFile,
				}
				cpOptions, err := controlplane.Complete(serverOptions)
				if err != nil {
					return err
				}

				server, err := controlplane.CreateServerChain(cpOptions, ctx.Done())
				if err != nil {
					return err
				}

				var clientConfig clientcmdapi.Config
				clientConfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
					"loopback": {Token: server.LoopbackClientConfig.BearerToken},
				}
				clientConfig.Clusters = map[string]*clientcmdapi.Cluster{
					// admin is the virtual cluster running by default
					"admin": {
						Server:                   server.LoopbackClientConfig.Host,
						CertificateAuthorityData: server.LoopbackClientConfig.CAData,
						TLSServerName:            server.LoopbackClientConfig.TLSClientConfig.ServerName,
					},
					// user is a virtual cluster that is lazily instantiated
					"user": {
						Server:                   server.LoopbackClientConfig.Host + "/clusters/user",
						CertificateAuthorityData: server.LoopbackClientConfig.CAData,
						TLSServerName:            server.LoopbackClientConfig.TLSClientConfig.ServerName,
					},
				}
				clientConfig.Contexts = map[string]*clientcmdapi.Context{
					"admin": {Cluster: "admin", AuthInfo: "loopback"},
					"user":  {Cluster: "user", AuthInfo: "loopback"},
				}
				clientConfig.CurrentContext = "admin"
				if err := clientcmd.WriteToFile(clientConfig, filepath.Join(s.Dir, "admin.kubeconfig")); err != nil {
					return err
				}

				prepared := server.PrepareRun()

				return prepared.Run(ctx.Done())
			})
		},
	}
	startCmd.Flags().AddFlag(pflag.PFlagFromGoFlag(flag.CommandLine.Lookup("v")))
	startCmd.Flags().StringVar(&listen, "listen", ":6443", "Address:port to bind to")
	cmd.AddCommand(startCmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
