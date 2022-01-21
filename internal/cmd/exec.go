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
				cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
				cmd.Flag("filter", "label=com.inaccel.docker.default-exec-service=True")
				cmd.Flag("format", `{{ .Label "com.docker.compose.service" }}`)
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return internal.ExitToStatus(err)
				}

				services := strings.Fields(out)

				if len(services) > 0 {
					exec.Set("service", services[0])
				} else {
					cmd = system.Command("docker")
					cmd.Flag("host", internal.Host)
					cmd.Flag("log-level", viper.GetString("log-level"))
					cmd.Arg("ps")
					cmd.Flag("all", true)
					cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.container-number=%d", exec.GetInt("index")))
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
						return fmt.Errorf("Error: No service (%d) found for %s", exec.GetInt("index"), viper.GetString("project-name"))
					} else if len(services) == 1 {
						exec.Set("service", services[0])
					} else {
						return fmt.Errorf("Error: A service (%d) must be specified for %s, choose one of: [%s]", exec.GetInt("index"), viper.GetString("project-name"), strings.Join(services, " "))
					}
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
				cmd.Arg(fmt.Sprintf("%s_%s_%d", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_"), exec.GetString("service"), exec.GetInt("index")))
				cmd.Std(nil, nil, os.Stderr)

				out, err := cmd.Out(viper.GetBool("debug"))
				if err != nil {
					return internal.ExitToStatus(err)
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
			cmd.Arg(fmt.Sprintf("%s_%s_%d", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_"), exec.GetString("service"), exec.GetInt("index")))
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
	Exec.Flags().Int("index", 1, "Index of the container if there are multiple instances of a service")
	exec.BindPFlag("index", Exec.Flags().Lookup("index"))
	Exec.RegisterFlagCompletionFunc("index", func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var cmd *system.Cmd

		if len(exec.GetString("service")) > 0 {
			cmd = system.Command("docker")
			cmd.Flag("host", internal.Host)
			cmd.Arg("ps")
			cmd.Flag("filter", "label=com.docker.compose.oneoff=False")
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.project=%s", regexp.MustCompile("[^-0-9_a-z]").ReplaceAllString(strings.ToLower(viper.GetString("project-name")), "_")))
			cmd.Flag("filter", fmt.Sprintf("label=com.docker.compose.service=%s", exec.GetString("service")))
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

	Exec.Flags().StringP("service", "s", "", "Service name")
	exec.BindPFlag("service", Exec.Flags().Lookup("service"))
	Exec.RegisterFlagCompletionFunc("service", func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var cmd *system.Cmd

		cmd = system.Command("docker")
		cmd.Flag("host", internal.Host)
		cmd.Arg("ps")
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

	Exec.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
	exec.BindPFlag("env", Exec.Flags().Lookup("env"))

	Exec.Flags().StringP("user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	exec.BindPFlag("user", Exec.Flags().Lookup("user"))

	Exec.Flags().StringP("workdir", "w", "", "Working directory inside the container")
	exec.BindPFlag("workdir", Exec.Flags().Lookup("workdir"))
}
