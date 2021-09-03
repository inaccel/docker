package cmd

import (
	"fmt"
	"os"

	"github.com/inaccel/docker/internal"
	"github.com/inaccel/docker/pkg/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ps = viper.New()

	// Ps : docker inaccel ps
	Ps = &cobra.Command{
		Use:   "ps [OPTIONS]",
		Short: "List containers",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			var cmd *system.Cmd

			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Flag("log-level", viper.GetString("log-level"))
			cmd.Arg("ps")
			cmd.Flag("all", true)
			cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", ps.GetString("project-name")))
			cmd.Flag("format", `table {{ .ID }}\t{{ .Image }}\t{{ .RunningFor }}\t{{ .Status }}\t{{ if eq ( .Label "com.docker.compose.service" ) "service" }}SERVICE{{ else }}{{ .Label "com.docker.compose.service" }}{{ end }}\t{{ if eq ( .Label "com.docker.compose.container-number" ) "container number" }}INDEX{{ else }}{{ .Label "com.docker.compose.container-number" }}{{ end }}`)
			cmd.Flag("no-trunc", ps.GetBool("no-trunc"))
			cmd.Flag("quiet", ps.GetBool("quiet"))
			cmd.Std(nil, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	Ps.Flags().Bool("no-trunc", false, "Don't truncate output")
	ps.BindPFlag("no-trunc", Ps.Flags().Lookup("no-trunc"))

	Ps.Flags().StringP("project-name", "p", "inaccel", "Specify an alternate project name")
	ps.BindPFlag("project-name", Ps.Flags().Lookup("project-name"))
	ps.BindEnv("project-name", "INACCEL_PROJECT_NAME")

	Ps.Flags().BoolP("quiet", "q", false, "Only display container IDs")
	ps.BindPFlag("quiet", Ps.Flags().Lookup("quiet"))
}
