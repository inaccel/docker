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
	down = viper.New()

	Down = &cobra.Command{
		Use:   "down [OPTIONS] [SERVICE]",
		Short: "Stop and remove containers, networks and volumes",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("ps")
			cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", down.GetString("project-name")))
			if len(args) > 0 {
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.service=%s", args[0]))
			}
			cmd.Flag("format", `{{ .ID }}`)
			cmd.Std(nil, nil, os.Stderr)

			out, err := cmd.Out(viper.GetBool("debug"))
			if err != nil {
				return err
			}

			ids := strings.Fields(out)

			if len(ids) > 0 {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("stop")
				cmd.Arg(ids...)
				cmd.Std(nil, nil, os.Stderr)

				if err := cmd.Run(viper.GetBool("debug")); err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("system", "prune")
			cmd.Flag("all", true)
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", down.GetString("project-name")))
			cmd.Flag("force", true)
			cmd.Std(nil, nil, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return err
			}

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("volume", "ls")
			cmd.Flag("filter", "dangling=true")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", down.GetString("project-name")))
			cmd.Flag("format", `{{ .Name }}`)
			cmd.Std(nil, nil, os.Stderr)

			out, err := cmd.Out(viper.GetBool("debug"))
			if err != nil {
				return err
			}

			names := strings.Fields(out)

			if len(names) > 0 {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("volume", "rm")
				cmd.Arg(names...)
				cmd.Std(nil, nil, os.Stderr)

				if err := cmd.Run(viper.GetBool("debug")); err != nil {
					return err
				}
			}

			return nil
		},
	}
)

func init() {
	Down.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
	down.BindPFlag("project-name", Down.Flags().Lookup("project-name"))
	down.BindEnv("project-name", "INACCEL_PROJECT_NAME")
}
