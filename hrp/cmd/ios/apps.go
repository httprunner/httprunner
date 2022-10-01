package ios

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var listAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List all iOS installed apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		apps, err := device.AppList()
		if err != nil {
			return errors.Wrap(err, "get ios apps failed")
		}
		for _, app := range apps {
			if appType != "all" && strings.ToLower(app.Type) != appType {
				continue
			}

			fmt.Printf("%-10.10s %-30.30s %-50.50s %-s\n",
				app.Type, app.DisplayName, app.CFBundleIdentifier, app.Version)
		}
		return nil
	},
}

var appType string

func init() {
	listAppsCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's udid")
	listAppsCmd.Flags().StringVarP(&appType, "type", "t", "user", "filter application type [user|system|pluginkit|all]")
	iosRootCmd.AddCommand(listAppsCmd)
}
