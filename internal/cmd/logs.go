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
	logs = viper.New()

	// Logs : docker inaccel logs
	Logs = &cobra.Command{
		Use:   "logs [OPTIONS] [PATTERN]",
		Short: "View output from containers",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(_ *cobra.Command, args []string) error {
			var cmd *system.Cmd

			if len(logs.GetString("service")) == 0 {
				cmd = system.Command("docker")
				cmd.Flag("host", internal.Host)
				cmd.Flag("log-level", viper.GetString("log-level"))
				cmd.Arg("ps")
				cmd.Flag("all", true)
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.container-number=%d", logs.GetInt("index")))
				cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
				cmd.Flag("filter", "label=com.inaccel.docker.default-logs-service=True")
				cmd.Flag("format", `{{ .Label "com.docker.compose.service" }}`)
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return internal.ExitToStatus(err)
				}

				services := strings.Fields(out)

				if len(services) > 0 {
					logs.Set("service", services[0])
				} else {
					cmd = system.Command("docker")
					cmd.Flag("host", internal.Host)
					cmd.Flag("log-level", viper.GetString("log-level"))
					cmd.Arg("ps")
					cmd.Flag("all", true)
					cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.container-number=%d", logs.GetInt("index")))
					cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
					cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
					cmd.Flag("format", `{{ .Label "com.docker.compose.service" }}`)
					cmd.Std(nil, nil, os.Stderr)

					out, err := cmd.Out(viper.GetBool("debug"))
					if err != nil {
						return internal.ExitToStatus(err)
					}

					services := strings.Fields(out)

					if len(services) == 0 {
						return fmt.Errorf("Error: No service (%d) found for %s", logs.GetInt("index"), viper.GetString("project-name"))
					} else if len(services) == 1 {
						logs.Set("service", services[0])
					} else {
						return fmt.Errorf("Error: A service (%d) must be specified for %s, choose one of: [%s]", logs.GetInt("index"), viper.GetString("project-name"), strings.Join(services, " "))
					}
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
			cmd.Flag("timestamps", logs.GetBool("timestamps"))
			cmd.Arg(fmt.Sprintf("%s_%s_%d", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_"), logs.GetString("service"), logs.GetInt("index")))
			switch len(args) {
			case 0:
				cmd.Std(nil, os.Stdout, os.Stderr)
			case 1:
				pattern, err := grep.Compile(args[0])
				if err != nil {
					return err
				}

				stdout := pattern.WriteCloser(os.Stdout, !logs.GetBool("no-color"), false)
				defer stdout.Close()

				cmd.Std(nil, stdout, os.Stderr)
			}

			if err := cmd.Run(viper.GetBool("debug")); err != nil {
				return internal.ExitToStatus(err)
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
	Logs.RegisterFlagCompletionFunc("index", func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var cmd *system.Cmd

		if len(logs.GetString("service")) > 0 {
			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Arg("ps")
			cmd.Flag("all", true)
			cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.service=%s", logs.GetString("service")))
			cmd.Flag("format", `{{ .Label "com.docker.compose.container-number" }}`)

			out, err := cmd.Out(false)
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for _, completion := range strings.Fields(out) {
				if strings.HasPrefix(completion, toComplete) {
					completions = append(completions, completion)
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveDefault
	})

	Logs.Flags().Bool("no-color", false, "Produce monochrome output")
	logs.BindPFlag("no-color", Logs.Flags().Lookup("no-color"))

	Logs.Flags().StringP("service", "s", "", "Service name")
	logs.BindPFlag("service", Logs.Flags().Lookup("service"))
	Logs.RegisterFlagCompletionFunc("service", func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var cmd *system.Cmd

		cmd = system.Command("docker")
		cmd.Flag("host", internal.Host)
		cmd.Arg("ps")
		cmd.Flag("all", true)
		cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
		cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
		cmd.Flag("format", `{{ .Label "com.docker.compose.service" }}`)

		out, err := cmd.Out(false)
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		var completions []string
		for _, completion := range strings.Fields(out) {
			if strings.HasPrefix(completion, toComplete) {
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	Logs.Flags().StringP("tail", "n", "10", "Number of lines to show from the end of the logs")
	logs.BindPFlag("tail", Logs.Flags().Lookup("tail"))
	Logs.RegisterFlagCompletionFunc("tail", func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for _, completion := range []string{"all"} {
			if strings.HasPrefix(completion, toComplete) {
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	Logs.Flags().BoolP("timestamps", "t", false, "Show timestamps")
	logs.BindPFlag("timestamps", Logs.Flags().Lookup("timestamps"))
}
