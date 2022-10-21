package ios

import (
	"fmt"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
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

		var applicationType giDevice.ApplicationType
		switch appType {
		case "user":
			applicationType = giDevice.ApplicationTypeUser
		case "system":
			applicationType = giDevice.ApplicationTypeSystem
		case "internal":
			applicationType = giDevice.ApplicationTypeInternal
		case "all":
			applicationType = giDevice.ApplicationTypeAny
		}

		result, errList := device.InstallationProxyBrowse(
			giDevice.WithApplicationType(applicationType),
			giDevice.WithReturnAttributes("CFBundleVersion", "CFBundleDisplayName", "CFBundleIdentifier"))
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
