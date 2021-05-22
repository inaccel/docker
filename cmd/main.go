package main

import (
	"net"
	"net/url"

	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/inaccel/docker/internal"
	"github.com/inaccel/docker/internal/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

func main() {
	plugin.Run(func(cli command.Cli) *cobra.Command {
		inaccel := &cobra.Command{
			Use:     "inaccel",
			Short:   "Simplifying FPGA management in Docker",
			Version: version,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				if err := plugin.PersistentPreRunE(cmd, args); err != nil {
					return err
				}

				viper.BindPFlag("debug", cmd.Parent().Parent().Flags().Lookup("debug"))
				viper.BindPFlag("log-level", cmd.Parent().Parent().Flags().Lookup("log-level"))

				endpoint := cli.DockerEndpoint()
				internal.Host, _ = url.Parse(endpoint.Host)

				switch internal.Host.Scheme {
				case "unix":
					return nil
				default:
					return net.UnknownNetworkError(internal.Host.Scheme)
				}
			},
		}

		inaccel.AddCommand(cmd.Down, cmd.Exec, cmd.Logs, cmd.Ps, cmd.Run, cmd.Up)

		return inaccel
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "InAccel <info@inaccel.com>",
		Version:       version,
		URL:           "https://inaccel.com",
	})
}
