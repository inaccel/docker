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
	logs = viper.New()

	Logs = &cobra.Command{
		Use:   "logs [OPTIONS]",
		Short: "View output from containers",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			if len(viper.GetString("service")) == 0 {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("ps")
				cmd.Flag("all", true)
				cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", logs.GetString("project-name")))
				cmd.Flag("filter", "label=com.inaccel.docker.default-logs-service=True")
				cmd.Flag("format", `{{ .Label "com.docker.compose.service" }}`)
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return err
				}

				services := strings.Fields(out)

				if len(services) > 0 {
					logs.Set("service", services[0])
				} else {
					return fmt.Errorf("Error: No service (%d) found for %s", logs.GetInt("index"), logs.GetString("project-name"))
				}
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("logs")
			cmd.Flag("follow", logs.GetBool("follow"))
			cmd.Flag("tail", logs.GetString("tail"))
			cmd.Arg(fmt.Sprintf("%s_%s_%d", logs.GetString("project-name"), logs.GetString("service"), logs.GetInt("index")))
			cmd.Std(os.Stdin, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	Logs.Flags().BoolP("follow", "f", false, "Follow log output")
	logs.BindPFlag("follow", Logs.Flags().Lookup("follow"))

	Logs.Flags().Int("index", 1, "Index of the container if there are multiple instances of a service")
	logs.BindPFlag("index", Logs.Flags().Lookup("index"))

	Logs.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
	logs.BindPFlag("project-name", Logs.Flags().Lookup("project-name"))

	Logs.Flags().StringP("service", "s", "", "Service name")
	logs.BindPFlag("service", Logs.Flags().Lookup("service"))

	Logs.Flags().StringP("tail", "n", "all", "Number of lines to show from the end of the logs")
	logs.BindPFlag("tail", Logs.Flags().Lookup("tail"))
}
