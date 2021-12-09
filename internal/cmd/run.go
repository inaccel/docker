package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/inaccel/docker/internal"
	"github.com/inaccel/docker/pkg/system"
	"github.com/inaccel/docker/pkg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	run = viper.New()

	// Run : docker inaccel run
	Run = &cobra.Command{
		Use:   "run [OPTIONS] SERVICE [COMMAND] [ARG...]",
		Short: "Run a one-off command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			rootless, err := internal.Rootless()
			if err != nil {
				return err
			}

			var cmd *system.Cmd

			if viper.GetBool("pull") {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("pull")
				cmd.Arg(fmt.Sprintf("%s:%s", viper.GetString("project-name"), viper.GetString("tag")))
				cmd.Std(nil, os.Stdout, os.Stderr)

				if err := cmd.Run(viper.GetBool("debug")); err != nil {
					return internal.ExitToStatus(err)
				}
			}

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("run")
			cmd.Flag("env", fmt.Sprintf("%s=%s", "DOCKER_HOST_PATH", internal.Host.Path))
			cmd.Flag("env", fmt.Sprintf("%s=%s", "DOCKER_HOST_SCHEME", internal.Host.Scheme))
			if rootless {
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_CACHE_HOME", xdg.CacheHome))
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_CONFIG_DIRS", strings.Join(xdg.ConfigDirs, ":")))
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_CONFIG_HOME", xdg.ConfigHome))
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_DATA_DIRS", strings.Join(xdg.DataDirs, ":")))
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_DATA_HOME", xdg.DataHome))
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_RUNTIME_DIR", xdg.RuntimeDir))
				cmd.Flag("env", fmt.Sprintf("%s=%s", "XDG_STATE_HOME", xdg.StateHome))
			}
			cmd.Flag("env", viper.GetStringSlice("env"))
			if len(viper.GetString("env-file")) > 0 {
				cmd.Flag("env-file", viper.GetString("env-file"))
			} else if _, err := os.Stat(".env"); err == nil {
				cmd.Flag("env-file", ".env")
			}
			cmd.Flag("interactive", true)
			cmd.Flag("rm", true)
			cmd.Flag("tty", true)
			cmd.Flag("volume", fmt.Sprintf("%s:%s", internal.Host.Path, "/var/run/docker.sock"))
			cmd.Arg(fmt.Sprintf("%s:%s", viper.GetString("project-name"), viper.GetString("tag")))
			cmd.Flag("ansi", "always")
			cmd.Flag("project-name", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_"))
			cmd.Arg("run")
			cmd.Flag("entrypoint", run.GetString("entrypoint"))
			cmd.Flag("e", run.GetStringSlice("env"))
			cmd.Flag("no-deps", run.GetBool("no-deps"))
			cmd.Flag("publish", run.GetStringSlice("publish"))
			cmd.Flag("rm", true)
			cmd.Flag("service-ports", true)
			cmd.Flag("user", run.GetString("user"))
			cmd.Flag("volume", run.GetStringSlice("volume"))
			cmd.Flag("workdir", run.GetString("workdir"))
			cmd.Arg(args...)
			cmd.Std(os.Stdin, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return internal.ExitToStatus(err)
			}

			return nil
		},
	}
)

func init() {
	Run.Flags().String("entrypoint", "", "Override the entrypoint of the container")
	run.BindPFlag("entrypoint", Run.Flags().Lookup("entrypoint"))

	Run.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
	run.BindPFlag("env", Run.Flags().Lookup("env"))

	Run.Flags().Bool("no-deps", false, "Don't start linked services")
	run.BindPFlag("no-deps", Run.Flags().Lookup("no-deps"))

	Run.Flags().StringSliceP("publish", "p", []string{}, "Publish a container's port(s) to the host")
	run.BindPFlag("publish", Run.Flags().Lookup("publish"))

	Run.Flags().StringP("user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	run.BindPFlag("user", Run.Flags().Lookup("user"))

	Run.Flags().StringSliceP("volume", "v", []string{}, "Bind mount a volume")
	run.BindPFlag("volume", Run.Flags().Lookup("volume"))

	Run.Flags().StringP("workdir", "w", "", "Working directory inside the container")
	run.BindPFlag("workdir", Run.Flags().Lookup("workdir"))
}
