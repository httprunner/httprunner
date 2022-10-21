package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/scaffold"
)

var scaffoldCmd = &cobra.Command{
	Use:     "startproject $project_name",
	Aliases: []string{"scaffold"},
	Short:   "create a scaffold project",
	Args:    cobra.ExactValidArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !ignorePlugin && !genPythonPlugin && !genGoPlugin {
			return errors.New("please specify function plugin type")
		}

		var pluginType scaffold.PluginType
		if empty {
			pluginType = scaffold.Empty
		} else if ignorePlugin {
			pluginType = scaffold.Ignore
		} else if genGoPlugin {
			pluginType = scaffold.Go
		} else {
			pluginType = scaffold.Py // default
		}

		err := scaffold.CreateScaffold(args[0], pluginType, venv, force)
		if err != nil {
			log.Error().Err(err).Msg("create scaffold project failed")
			return err
		}
		log.Info().Str("projectName", args[0]).Msg("create scaffold success")
		return nil
	},
}

var (
	empty           bool
	ignorePlugin    bool
	genPythonPlugin bool
	genGoPlugin     bool
	force           bool
)

func init() {
	rootCmd.AddCommand(scaffoldCmd)
	scaffoldCmd.Flags().BoolVarP(&force, "force", "f", false, "force to overwrite existing project")
	scaffoldCmd.Flags().BoolVar(&genPythonPlugin, "py", true, "generate hashicorp python plugin")
	scaffoldCmd.Flags().BoolVar(&genGoPlugin, "go", false, "generate hashicorp go plugin")
	scaffoldCmd.Flags().BoolVar(&ignorePlugin, "ignore-plugin", false, "ignore function plugin")
	scaffoldCmd.Flags().BoolVar(&empty, "empty", false, "generate empty project")
}
