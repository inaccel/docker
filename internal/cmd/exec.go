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
	exec = viper.New()

	// Exec : docker inaccel exec
	Exec = &cobra.Command{
		Use:   "exec [OPTIONS] COMMAND [ARG...]",
		Short: "Execute a command in a running container",
		Args:  cobra.ArbitraryArgs,
		PreRunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			if len(exec.GetString("service")) == 0 {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("ps")
				cmd.Flag("all", true)
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.container-number=%d", exec.GetInt("index")))
				cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", viper.GetString("project-name")))
				cmd.Flag("filter", "label=com.inaccel.docker.default-exec-service=True")
				cmd.Flag("format", `{{ .Label "com.docker.compose.service" }}`)
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return err
				}

				services := strings.Fields(out)

				if len(services) > 0 {
					exec.Set("service", services[0])
				} else {
					return fmt.Errorf("Error: No service (%d) found for %s", exec.GetInt("index"), viper.GetString("project-name"))
				}
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			if err := cobra.MinimumNArgs(1)(nil, args); err != nil {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("inspect")
				cmd.Flag("format", `{{ index .Config.Labels "com.inaccel.docker.default-exec-command" }}`)
				cmd.Arg(fmt.Sprintf("%s_%s_%d", viper.GetString("project-name"), exec.GetString("service"), exec.GetInt("index")))
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return err
				}

				args = strings.Fields(out)

				if err := cobra.MinimumNArgs(1)(nil, args); err != nil {
					return err
				}
			}

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("exec")
			cmd.Flag("env", exec.GetStringSlice("env"))
			cmd.Flag("interactive", true)
			cmd.Flag("tty", true)
			cmd.Flag("user", exec.GetString("user"))
			cmd.Flag("workdir", exec.GetString("workdir"))
			cmd.Arg(fmt.Sprintf("%s_%s_%d", viper.GetString("project-name"), exec.GetString("service"), exec.GetInt("index")))
			cmd.Arg(args...)
			cmd.Std(os.Stdin, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	Exec.Flags().Int("index", 1, "Index of the container if there are multiple instances of a service")
	exec.BindPFlag("index", Exec.Flags().Lookup("index"))

	Exec.Flags().StringP("service", "s", "", "Service name")
	exec.BindPFlag("service", Exec.Flags().Lookup("service"))

	Exec.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
	exec.BindPFlag("env", Exec.Flags().Lookup("env"))

	Exec.Flags().StringP("user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	exec.BindPFlag("user", Exec.Flags().Lookup("user"))

	Exec.Flags().StringP("workdir", "w", "", "Working directory inside the container")
	exec.BindPFlag("workdir", Exec.Flags().Lookup("workdir"))
}
