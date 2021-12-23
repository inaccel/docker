package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/google/shlex"
	"github.com/inaccel/docker/internal"
	"github.com/inaccel/docker/internal/cmd"
	"github.com/inaccel/docker/pkg/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

func main() {
	plugin.Run(func(cli command.Cli) *cobra.Command {
		inaccel := &cobra.Command{
			Use:     "inaccel",
			Short:   "Simplifying FPGA management in Docker",
			Args:    cobra.NoArgs,
			Version: version,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				if err := plugin.PersistentPreRunE(cmd, args); err != nil {
					return err
				}

				viper.BindPFlag("debug", cmd.Root().Flags().Lookup("debug"))
				viper.BindPFlag("log-level", cmd.Root().Flags().Lookup("log-level"))

				endpoint := cli.DockerEndpoint()
				internal.Host, _ = url.Parse(endpoint.Host)

				switch internal.Host.Scheme {
				case "unix":
					return nil
				default:
					return net.UnknownNetworkError(internal.Host.Scheme)
				}
			},
			Run: func(cmd *cobra.Command, _ []string) {
				executor := func(input string) {
					args, err := shlex.Split(input)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						return
					}

					if len(args) == 0 {
						return
					}

					inaccel := system.Command(os.Args[0])
					inaccel.Arg(os.Args[1:]...)
					inaccel.Arg(args...)
					inaccel.Env(os.Environ()...)
					inaccel.Std(os.Stdin, os.Stdout, os.Stderr)

					if err := inaccel.Run(false); err != nil {
						fmt.Fprintln(os.Stderr, err)
						return
					}
				}

				completer := func(document prompt.Document) []prompt.Suggest {
					if document.GetCharRelativeToCursor(1) > 0 && document.GetCharRelativeToCursor(1) != ' ' && document.GetCharRelativeToCursor(1) != '=' {
						return nil
					}
					input := document.CurrentLineBeforeCursor()

					args, err := shlex.Split(input)
					if err != nil {
						return nil
					}

					if len(args) == 0 || strings.HasSuffix(input, " ") {
						args = append(args, "")
					}

					inaccel := system.Command(os.Args[0])
					inaccel.Arg(cobra.ShellCompRequestCmd)
					inaccel.Arg(os.Args[1:]...)
					inaccel.Arg(args...)
					inaccel.Env(os.Environ()...)

					out, err := inaccel.Out(false)
					if err != nil {
						return nil
					}

					var suggestions []prompt.Suggest
					for _, line := range strings.Split(out, "\n") {
						if len(line) > 0 && !strings.HasPrefix(line, ":") {
							suggestion := strings.SplitN(line, "\t", 2)
							switch len(suggestion) {
							case 1:
								suggestions = append(suggestions, prompt.Suggest{
									Text: suggestion[0],
								})
							case 2:
								suggestions = append(suggestions, prompt.Suggest{
									Text:        suggestion[0],
									Description: suggestion[1],
								})
							}
						}
					}
					return suggestions
				}

				fmt.Fprintln(os.Stdout, "Use Ctrl-D (i.e. EOF) to quit")

				ps1, ok := os.LookupEnv("INACCEL_PS1")
				if !ok {
					args := make([]string, len(os.Args))
					args[0] = cmd.Root().Name()
					copy(args[1:], os.Args[1:])
					ps1 = fmt.Sprintf("$ %s ", strings.Join(args, " "))
				}

				prompt.New(
					executor,
					completer,
					prompt.OptionCompletionOnDown(),
					prompt.OptionCompletionWordSeparator(" ="),
					prompt.OptionDescriptionBGColor(prompt.DefaultColor),
					prompt.OptionDescriptionTextColor(prompt.Blue),
					prompt.OptionInputBGColor(prompt.DefaultColor),
					prompt.OptionInputTextColor(prompt.DefaultColor),
					prompt.OptionPrefix(ps1),
					prompt.OptionPrefixBackgroundColor(prompt.DefaultColor),
					prompt.OptionPrefixTextColor(prompt.DarkBlue),
					prompt.OptionPreviewSuggestionBGColor(prompt.DefaultColor),
					prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
					prompt.OptionScrollbarBGColor(prompt.DefaultColor),
					prompt.OptionScrollbarThumbColor(prompt.DarkGray),
					prompt.OptionSelectedDescriptionBGColor(prompt.DarkBlue),
					prompt.OptionSelectedDescriptionTextColor(prompt.White),
					prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
					prompt.OptionSelectedSuggestionTextColor(prompt.Black),
					prompt.OptionShowCompletionAtStart(),
					prompt.OptionSuggestionBGColor(prompt.DefaultColor),
					prompt.OptionSuggestionTextColor(prompt.DefaultColor),
				).Run()
			},
		}

		inaccel.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")
		viper.BindPFlag("env", inaccel.Flags().Lookup("env"))

		inaccel.Flags().String("env-file", "", "Specify an alternate environment file")
		inaccel.MarkFlagFilename("env-file")
		viper.BindPFlag("env-file", inaccel.Flags().Lookup("env-file"))

		inaccel.Flags().StringSlice("profile", []string{}, "Specify a profile to enable")
		viper.BindPFlag("profile", inaccel.Flags().Lookup("profile"))
		viper.BindEnv("profile", "INACCEL_PROFILES")

		inaccel.Flags().StringP("project-name", "p", fmt.Sprintf("inaccel/%s", internal.Config), "Specify an alternate project name")
		viper.BindPFlag("project-name", inaccel.Flags().Lookup("project-name"))
		viper.BindEnv("project-name", "INACCEL_PROJECT_NAME")

		inaccel.Flags().Bool("pull", false, "Always attempt to pull a newer version of the project")
		viper.BindPFlag("pull", inaccel.Flags().Lookup("pull"))

		inaccel.Flags().StringP("tag", "t", "latest", "Specify the project tag to use")
		viper.BindPFlag("tag", inaccel.Flags().Lookup("tag"))

		inaccel.Flags().BoolP("version", "v", false, "Print version information and quit")
		viper.BindPFlag("version", inaccel.Flags().Lookup("version"))

		inaccel.AddCommand(cmd.Config, cmd.Down, cmd.Exec, cmd.Logs, cmd.Ps, cmd.Run, cmd.Up)

		return inaccel
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "InAccel <info@inaccel.com>",
		Version:       version,
		URL:           "https://inaccel.com",
	})
}
