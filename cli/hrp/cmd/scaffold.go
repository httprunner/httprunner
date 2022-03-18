package cmd

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/hrp/internal/scaffold"
)

var scaffoldCmd = &cobra.Command{
	Use:   "startproject $project_name",
	Short: "create a scaffold project",
	Args:  cobra.ExactValidArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !ignorePlugin && !genPythonPlugin && !genGoPlugin {
			return errors.New("please select function plugin type")
		}

		var pluginType scaffold.PluginType
		if ignorePlugin {
			pluginType = scaffold.Ignore
		} else if genGoPlugin {
			pluginType = scaffold.Go
		} else {
			pluginType = scaffold.Py // default
		}
		err := scaffold.CreateScaffold(args[0], pluginType)
		if err != nil {
			log.Error().Err(err).Msg("create scaffold project failed")
			os.Exit(1)
		}
		return nil
	},
}

var (
	ignorePlugin    bool
	genPythonPlugin bool
	genGoPlugin     bool
)

func init() {
	rootCmd.AddCommand(scaffoldCmd)
	scaffoldCmd.Flags().BoolVar(&genPythonPlugin, "py", true, "generate hashicorp python plugin")
	scaffoldCmd.Flags().BoolVar(&genGoPlugin, "go", false, "generate hashicorp go plugin")
	scaffoldCmd.Flags().BoolVar(&ignorePlugin, "ignore-plugin", false, "ignore function plugin")
}
