package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/aenix-io/talm/pkg/commands"
	"github.com/siderolabs/talos/cmd/talosctl/cmd/common"
	"github.com/siderolabs/talos/pkg/machinery/constants"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:               "talm",
	Short:             "Just like Helm, but for Talos",
	Long:              ``,
	SilenceErrors:     true,
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}

func Execute() error {
	rootCmd.PersistentFlags().StringVar(
		&commands.GlobalArgs.Talosconfig,
		"talosconfig",
		"",
		fmt.Sprintf("The path to the Talos configuration file. Defaults to '%s' env variable if set, otherwise '%s' and '%s' in order.",
			constants.TalosConfigEnvVar,
			filepath.Join("$HOME", constants.TalosDir, constants.TalosconfigFilename),
			filepath.Join(constants.ServiceAccountMountPath, constants.TalosconfigFilename),
		),
	)
	rootCmd.PersistentFlags().StringVar(&commands.GlobalArgs.CmdContext, "context", "", "Context to be used in command")
	rootCmd.PersistentFlags().StringSliceVarP(&commands.GlobalArgs.Nodes, "nodes", "n", []string{}, "target the specified nodes")
	rootCmd.PersistentFlags().StringSliceVarP(&commands.GlobalArgs.Endpoints, "endpoints", "e", []string{}, "override default endpoints in Talos configuration")
	rootCmd.PersistentFlags().StringVar(&commands.GlobalArgs.Cluster, "cluster", "", "Cluster to connect to if a proxy endpoint is used.")

	cmd, err := rootCmd.ExecuteContextC(context.Background())
	if err != nil && !common.SuppressErrors {
		fmt.Fprintln(os.Stderr, err.Error())

		errorString := err.Error()
		// TODO: this is a nightmare, but arg-flag related validation returns simple `fmt.Errorf`, no way to distinguish
		//       these errors
		if strings.Contains(errorString, "arg(s)") || strings.Contains(errorString, "flag") || strings.Contains(errorString, "command") {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, cmd.UsageString())
		}
	}

	return err
}

func init() {
	rootCmd.PersistentFlags().StringVar(&commands.Config.RootDir, "root", ".", "root directory of the project")
	cobra.OnInitialize(initConfig)

	for _, cmd := range commands.Commands {
		rootCmd.AddCommand(cmd)
	}
}

func initConfig() {
	cmd, _, _ := rootCmd.Find(os.Args[1:])
	if cmd != nil && cmd.Use != "init" {
		configFile := filepath.Join(commands.Config.RootDir, "Chart.yaml")
		if err := loadConfig(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}
	}
}

func loadConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading configuration file: %w", err)
	}
	if err := yaml.Unmarshal(data, &commands.Config); err != nil {
		return fmt.Errorf("error unmarshalling configuration: %w", err)
	}
	commands.GlobalArgs.Talosconfig = commands.Config.GlobalOptions.Talosconfig
	return nil
}
