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

			if viper.GetBool("pull") {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("pull")
				cmd.Arg(fmt.Sprintf("%s:%s", viper.GetString("project-name"), viper.GetString("tag")))
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
			cmd.Flag("env", viper.GetStringSlice("env"))
			if len(viper.GetString("env-file")) > 0 {
				cmd.Flag("env-file", viper.GetString("env-file"))
			} else if _, err := os.Stat(".env"); err == nil {
				cmd.Flag("env-file", ".env")
			}
			cmd.Flag("interactive", true)
			cmd.Flag("rm", true)
			cmd.Flag("volume", fmt.Sprintf("%s:%s", internal.Host.Path, "/var/run/docker.sock"))
			cmd.Arg(fmt.Sprintf("%s:%s", viper.GetString("project-name"), viper.GetString("tag")))
			cmd.Flag("ansi", "always")
			cmd.Flag("profile", viper.GetStringSlice("profile"))
			cmd.Flag("project-name", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_"))
			cmd.Arg("up")
			cmd.Flag("always-recreate-deps", up.GetBool("always-recreate-deps"))
			cmd.Flag("detach", true)
			cmd.Flag("force-recreate", up.GetBool("force-recreate"))
			cmd.Flag("no-deps", up.GetBool("no-deps"))
			cmd.Flag("no-recreate", up.GetBool("no-recreate"))
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
	Up.Flags().Bool("always-recreate-deps", false, "Recreate dependent containers")
	up.BindPFlag("always-recreate-deps", Up.Flags().Lookup("always-recreate-deps"))

	Up.Flags().Bool("force-recreate", false, "Recreate containers even if their configuration and image haven't changed")
	up.BindPFlag("force-recreate", Up.Flags().Lookup("force-recreate"))

	Up.Flags().Bool("no-deps", false, "Don't start linked services")
	up.BindPFlag("no-deps", Up.Flags().Lookup("no-deps"))

	Up.Flags().Bool("no-recreate", false, "If containers already exist, don't recreate them")
	up.BindPFlag("no-recreate", Up.Flags().Lookup("no-recreate"))
}
