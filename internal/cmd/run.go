package cmd

import (
	"fmt"
	"os"
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

			if run.GetBool("pull") {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("pull")
				if strings.ContainsAny(run.GetString("tag"), "/:") {
					cmd.Arg(run.GetString("tag"))
				} else {
					cmd.Arg(fmt.Sprintf("inaccel/%s:%s", internal.Config, run.GetString("tag")))
				}
				cmd.Std(nil, os.Stdout, os.Stderr)

				if err := cmd.Run(viper.GetBool("debug")); err != nil {
					return err
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
			cmd.Flag("env", run.GetStringSlice("env"))
			if len(run.GetString("env-file")) > 0 {
				cmd.Flag("env-file", run.GetString("env-file"))
			} else if _, err := os.Stat(".env"); err == nil {
				cmd.Flag("env-file", ".env")
			}
			cmd.Flag("interactive", true)
			cmd.Flag("rm", true)
			cmd.Flag("tty", true)
			cmd.Flag("volume", fmt.Sprintf("%s:%s", internal.Host.Path, "/var/run/docker.sock"))
			if strings.ContainsAny(run.GetString("tag"), "/:") {
				cmd.Arg(run.GetString("tag"))
			} else {
				cmd.Arg(fmt.Sprintf("inaccel/%s:%s", internal.Config, run.GetString("tag")))
			}
			cmd.Flag("ansi", "always")
			cmd.Flag("project-name", run.GetString("project-name"))
			cmd.Arg("run")
			cmd.Flag("entrypoint", run.GetString("entrypoint"))
			cmd.Flag("rm", true)
			cmd.Flag("service-ports", true)
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
	Run.Flags().String("entrypoint", "", "Override the entrypoint of the service")
	run.BindPFlag("entrypoint", Run.Flags().Lookup("entrypoint"))

	Run.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
	run.BindPFlag("env", Run.Flags().Lookup("env"))

	Run.Flags().String("env-file", "", "Specify an alternate environment file")
	Run.MarkFlagFilename("env-file")
	run.BindPFlag("env-file", Run.Flags().Lookup("env-file"))

	Run.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
	run.BindPFlag("project-name", Run.Flags().Lookup("project-name"))
	run.BindEnv("project-name", "INACCEL_PROJECT_NAME")

	Run.Flags().Bool("pull", false, "Always attempt to pull a newer version of the image")
	run.BindPFlag("pull", Run.Flags().Lookup("pull"))

	Run.Flags().StringP("tag", "t", "latest", "Tag and optionally a name in the 'name:tag' format")
	run.BindPFlag("tag", Run.Flags().Lookup("tag"))
}
