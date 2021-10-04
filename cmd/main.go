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

		inaccel.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
		viper.BindPFlag("env", inaccel.Flags().Lookup("env"))

		inaccel.Flags().String("env-file", "", "Specify an alternate environment file")
		inaccel.MarkFlagFilename("env-file")
		viper.BindPFlag("env-file", inaccel.Flags().Lookup("env-file"))

		inaccel.Flags().StringSlice("profile", []string{}, "Specify a profile to enable")
		viper.BindPFlag("profile", inaccel.Flags().Lookup("profile"))
		viper.BindEnv("profile", "INACCEL_PROFILES")

		inaccel.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
		viper.BindPFlag("project-name", inaccel.Flags().Lookup("project-name"))
		viper.BindEnv("project-name", "INACCEL_PROJECT_NAME")

		inaccel.Flags().Bool("pull", false, "Always attempt to pull a newer version of the image")
		viper.BindPFlag("pull", inaccel.Flags().Lookup("pull"))

		inaccel.Flags().StringP("tag", "t", "latest", "Tag and optionally a name in the 'name:tag' format")
		viper.BindPFlag("tag", inaccel.Flags().Lookup("tag"))

		inaccel.Flags().BoolP("version", "v", false, "Print version information and quit")
		viper.BindPFlag("version", inaccel.Flags().Lookup("version"))

		inaccel.AddCommand(cmd.Config, cmd.Down, cmd.Exec, cmd.Logs, cmd.Ps, cmd.Run, cmd.Up)

		return inaccel
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "InAccel <info@inaccel.com>",
		Version:       version,
		URL:           "https://inaccel.com",
	})
}
