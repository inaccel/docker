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
	up = viper.New()

	// Up : docker inaccel up
	Up = &cobra.Command{
		Use:   "up [OPTIONS] [SERVICE]",
		Short: "Create and start containers",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			rootless, err := internal.Rootless()
			if err != nil {
				return err
			}

			var cmd *system.Cmd

			if up.GetBool("pull") {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("pull")
				if strings.ContainsAny(up.GetString("tag"), "/:") {
					cmd.Arg(up.GetString("tag"))
				} else {
					cmd.Arg(fmt.Sprintf("inaccel/%s:%s", internal.Config, up.GetString("tag")))
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
			cmd.Flag("env", up.GetStringSlice("env"))
			if len(up.GetString("env-file")) > 0 {
				cmd.Flag("env-file", up.GetString("env-file"))
			} else if _, err := os.Stat(".env"); err == nil {
				cmd.Flag("env-file", ".env")
			}
			cmd.Flag("interactive", true)
			cmd.Flag("rm", true)
			cmd.Flag("volume", fmt.Sprintf("%s:%s", internal.Host.Path, "/var/run/docker.sock"))
			if strings.ContainsAny(up.GetString("tag"), "/:") {
				cmd.Arg(up.GetString("tag"))
			} else {
				cmd.Arg(fmt.Sprintf("inaccel/%s:%s", internal.Config, up.GetString("tag")))
			}
			cmd.Flag("ansi", "always")
			cmd.Flag("profile", up.GetStringSlice("profile"))
			cmd.Flag("project-name", up.GetString("project-name"))
			cmd.Arg("up")
			cmd.Flag("detach", true)
			if len(args) > 0 {
				cmd.Arg(args[0])
			}
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
	up.BindEnv("profile", "INACCEL_PROFILES")

	Up.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
	up.BindPFlag("project-name", Up.Flags().Lookup("project-name"))
	up.BindEnv("project-name", "INACCEL_PROJECT_NAME")

	Up.Flags().Bool("pull", false, "Always attempt to pull a newer version of the image")
	up.BindPFlag("pull", Up.Flags().Lookup("pull"))

	Up.Flags().StringP("tag", "t", "latest", "Tag and optionally a name in the 'name:tag' format")
	up.BindPFlag("tag", Up.Flags().Lookup("tag"))
}
