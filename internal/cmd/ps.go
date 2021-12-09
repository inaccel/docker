package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
			if !ps.GetBool("quiet") {
				cmd.Flag("format", `table {{ .ID }}\t{{ .Image }}\t{{ .RunningFor }}\t{{ .Status }}\t{{ if eq ( .Label "com.docker.compose.service" ) "service" }}SERVICE{{ else }}{{ .Label "com.docker.compose.service" }}{{ end }}\t{{ if eq ( .Label "com.docker.compose.container-number" ) "container number" }}INDEX{{ else }}{{ .Label "com.docker.compose.container-number" }}{{ end }}`)
			}
			cmd.Flag("no-trunc", ps.GetBool("no-trunc"))
			cmd.Flag("quiet", ps.GetBool("quiet"))
			cmd.Std(nil, os.Stdout, os.Stderr)

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return internal.ExitToStatus(err)
			}

			return nil
		},
	}
)

func init() {
	Ps.Flags().Bool("no-trunc", false, "Don't truncate output")
	ps.BindPFlag("no-trunc", Ps.Flags().Lookup("no-trunc"))

	Ps.Flags().BoolP("quiet", "q", false, "Only display container IDs")
	ps.BindPFlag("quiet", Ps.Flags().Lookup("quiet"))
}
