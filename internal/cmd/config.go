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
	config = viper.New()

	// Config : docker inaccel config
	Config = &cobra.Command{
		Use:   "config [OPTIONS]",
		Short: "Validate and view the config file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
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
				if strings.ContainsAny(viper.GetString("tag"), "/:") {
					cmd.Arg(viper.GetString("tag"))
				} else {
					cmd.Arg(fmt.Sprintf("inaccel/%s:%s", internal.Config, viper.GetString("tag")))
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
			cmd.Flag("env", viper.GetStringSlice("env"))
			if len(viper.GetString("env-file")) > 0 {
				cmd.Flag("env-file", viper.GetString("env-file"))
			} else if _, err := os.Stat(".env"); err == nil {
				cmd.Flag("env-file", ".env")
			}
			cmd.Flag("interactive", true)
			cmd.Flag("rm", true)
			cmd.Flag("volume", fmt.Sprintf("%s:%s", internal.Host.Path, "/var/run/docker.sock"))
			if strings.ContainsAny(viper.GetString("tag"), "/:") {
				cmd.Arg(viper.GetString("tag"))
			} else {
				cmd.Arg(fmt.Sprintf("inaccel/%s:%s", internal.Config, viper.GetString("tag")))
			}
			cmd.Flag("ansi", "always")
			cmd.Arg("config")
			cmd.Flag("profiles", config.GetBool("profiles"))
			cmd.Flag("quiet", config.GetBool("quiet"))
			cmd.Flag("resolve-image-digests", config.GetBool("resolve-image-digests"))
			cmd.Flag("services", config.GetBool("services"))
			cmd.Std(nil, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	Config.Flags().Bool("profiles", false, "Print the profile names, one per line")
	config.BindPFlag("profiles", Config.Flags().Lookup("profiles"))

	Config.Flags().BoolP("quiet", "q", false, "Only validate the configuration, don't print anything")
	config.BindPFlag("quiet", Config.Flags().Lookup("quiet"))

	Config.Flags().Bool("resolve-image-digests", false, "Pin image tags to digests")
	config.BindPFlag("resolve-image-digests", Config.Flags().Lookup("resolve-image-digests"))

	Config.Flags().Bool("services", false, "Print the service names, one per line")
	config.BindPFlag("services", Config.Flags().Lookup("services"))
}
