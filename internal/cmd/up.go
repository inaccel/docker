package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/inaccel/docker/internal"
	"github.com/inaccel/docker/pkg/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	up = viper.New()

	Up = &cobra.Command{
		Use:   "up [OPTIONS]",
		Short: "Create and start containers",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			var cmd *system.Cmd

			if up.GetBool("pull") {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("pull")
				if strings.Contains(up.GetString("tag"), ":") {
					cmd.Arg(up.GetString("tag"))
				} else {
					cmd.Arg(fmt.Sprintf("%s:%s", "inaccel/fpga-operator", up.GetString("tag")))
				}
				cmd.Std(nil, os.Stdout, os.Stderr)

				if err := cmd.Run(up.GetBool("debug")); err != nil {
					return err
				}
			}

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("run")
			cmd.Flag("env", up.GetStringSlice("env"))
			if len(up.GetString("env-file")) > 0 {
				cmd.Flag("env-file", up.GetString("env-file"))
			} else if _, err := os.Stat(".env"); err == nil {
				cmd.Flag("env-file", ".env")
			}
			cmd.Flag("interactive", true)
			cmd.Flag("volume", fmt.Sprintf("%s:%s", internal.Host.Path, "/var/run/docker.sock"))
			if strings.Contains(up.GetString("tag"), ":") {
				cmd.Arg(up.GetString("tag"))
			} else {
				cmd.Arg(fmt.Sprintf("%s:%s", "inaccel/fpga-operator", up.GetString("tag")))
			}
			cmd.Arg("docker-compose")
			cmd.Flag("profile", up.GetStringSlice("profile"))
			cmd.Flag("project-name", up.GetString("project-name"))
			cmd.Arg("up")
			cmd.Flag("detach", true)
			cmd.Std(nil, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	Up.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
	up.BindPFlag("env", Up.Flags().Lookup("env"))

	Up.Flags().String("env-file", "", "Specify an alternate environment file")
	Up.MarkFlagFilename("env-file")
	up.BindPFlag("env-file", Up.Flags().Lookup("env-file"))

	Up.Flags().StringSlice("profile", []string{}, "Specify a profile to enable")
	up.BindPFlag("profile", Up.Flags().Lookup("profile"))

	Up.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
	up.BindPFlag("project-name", Up.Flags().Lookup("project-name"))

	Up.Flags().Bool("pull", false, "Always attempt to pull a newer version of the image")
	up.BindPFlag("pull", Up.Flags().Lookup("pull"))

	Up.Flags().StringP("tag", "t", "latest", "Tag and optionally a name in the 'name:tag' format")
	up.BindPFlag("tag", Up.Flags().Lookup("tag"))
}
