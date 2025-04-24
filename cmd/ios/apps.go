package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/uixt"
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_apps", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		device.GetDeviceInfo()
		var applicationType uixt.ApplicationType
		switch appType {
		case "user":
			applicationType = uixt.ApplicationTypeUser
		case "system":
			applicationType = uixt.ApplicationTypeSystem
		case "internal":
			applicationType = uixt.ApplicationTypeInternal
		case "all":
			applicationType = uixt.ApplicationTypeAny
		}

		result, err := device.ListApps(applicationType)
		if err != nil {
			return fmt.Errorf("get app list failed %v", err)
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
	CmdIOSRoot.AddCommand(listAppsCmd)
}
