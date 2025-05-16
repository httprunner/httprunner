package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/scaffold"
)

var CmdScaffold = &cobra.Command{
	Use:     "startproject $project_name",
	Aliases: []string{"scaffold"},
	Short:   "Create a scaffold project",
	Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
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
	CmdScaffold.Flags().BoolVarP(&force, "force", "f", false, "force to overwrite existing project")
	CmdScaffold.Flags().BoolVar(&genPythonPlugin, "py", true, "generate hashicorp python plugin")
	CmdScaffold.Flags().BoolVar(&genGoPlugin, "go", false, "generate hashicorp go plugin")
	CmdScaffold.Flags().BoolVar(&ignorePlugin, "ignore-plugin", false, "ignore function plugin")
	CmdScaffold.Flags().BoolVar(&empty, "empty", false, "generate empty project")
}
