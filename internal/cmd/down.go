package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/inaccel/docker/internal"
	"github.com/inaccel/docker/pkg/grep"
	"github.com/inaccel/docker/pkg/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	down = viper.New()

	// Down : docker inaccel down
	Down = &cobra.Command{
		Use:   "down [SERVICE]",
		Short: "Stop and remove containers and networks",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("ps")
			cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
			if len(args) > 0 {
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.service=%s", args[0]))
			}
			cmd.Flag("format", `{{ .ID }}`)
			cmd.Std(nil, nil, os.Stderr)

			out, err := cmd.Out(viper.GetBool("debug"))
			if err != nil {
				return internal.ExitToStatus(err)
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
					return internal.ExitToStatus(err)
				}
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("container", "prune")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
			cmd.Flag("force", true)
			cmd.Std(nil, grep.MustCompile("^$|Total reclaimed space").WriteCloser(os.Stdout, false, true), os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return internal.ExitToStatus(err)
			}

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("network", "prune")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
			cmd.Flag("force", true)
			cmd.Std(nil, grep.MustCompile("^$|Total reclaimed space").WriteCloser(os.Stdout, false, true), os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return internal.ExitToStatus(err)
			}

			if down.GetBool("volumes") {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("volume", "ls")
				cmd.Flag("filter", "dangling=true")
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
				cmd.Flag("format", `{{ .Name }}`)
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return internal.ExitToStatus(err)
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
						return internal.ExitToStatus(err)
					}

					fmt.Fprintln(os.Stdout, "Deleted Volumes:")
					fmt.Fprint(os.Stdout, out)
				}
			}

			return nil
		},
	}
)

func init() {
	Down.Flags().BoolP("volumes", "v", false, "Remove volumes attached to containers")
	down.BindPFlag("volumes", Down.Flags().Lookup("volumes"))
}
