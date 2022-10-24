package ios

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

type Application struct {
	CFBundleVersion     string `json:"version"`
	CFBundleDisplayName string `json:"name"`
	CFBundleIdentifier  string `json:"bundleId"`
}

var listAppsCmd = &cobra.Command{
	Use:              "apps",
	Short:            "List all iOS installed apps",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		var applicationType gidevice.ApplicationType
		switch appType {
		case "user":
			applicationType = gidevice.ApplicationTypeUser
		case "system":
			applicationType = gidevice.ApplicationTypeSystem
		case "internal":
			applicationType = gidevice.ApplicationTypeInternal
		case "all":
			applicationType = gidevice.ApplicationTypeAny
		}

		result, errList := device.InstallationProxyBrowse(
			gidevice.WithApplicationType(applicationType),
			gidevice.WithReturnAttributes("CFBundleVersion", "CFBundleDisplayName", "CFBundleIdentifier"))
		if errList != nil {
			return fmt.Errorf("get app list failed")
		}

		for _, app := range result {
			a := Application{}
			mapstructure.Decode(app, &a)

			fmt.Printf("%-30.30s %-50.50s %-s\n",
				a.CFBundleDisplayName, a.CFBundleIdentifier, a.CFBundleVersion)
		}
		return nil
	},
}

var appType string

func init() {
	listAppsCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	listAppsCmd.Flags().StringVarP(&appType, "type", "t", "user", "filter application type [user|system|internal|all]")
	iosRootCmd.AddCommand(listAppsCmd)
}
