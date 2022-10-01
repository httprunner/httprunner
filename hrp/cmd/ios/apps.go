package ios

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/uixt"
)

var listIOSAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List all iOS installed apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := uixt.IOSDevices(udid)
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			fmt.Println("no ios device found")
			os.Exit(1)
		}
		if len(devices) > 1 {
			return fmt.Errorf("multiple devices found, please specify udid")
		}
		device := devices[0]

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
	listIOSAppsCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's udid")
	listIOSAppsCmd.Flags().StringVarP(&appType, "type", "t", "user", "filter application type [user|system|pluginkit|all]")
	iosRootCmd.AddCommand(listIOSAppsCmd)
}
