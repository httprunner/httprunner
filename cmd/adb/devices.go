package adb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
)

var listAndroidDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List all Android devices",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_adb_devices", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		deviceList, err := getAndroidDevices()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, d := range deviceList {
			fmt.Println(format(d.DeviceInfo()))
		}
		return nil
	},
}

func format(data map[string]string) string {
	result, _ := json.MarshalIndent(data, "", "\t")
	return string(result)
}

func getAndroidDevices() (devices []*gadb.Device, err error) {
	var adbClient gadb.Client
	if adbClient, err = gadb.NewClient(); err != nil {
		return nil, err
	}

	if devices, err = adbClient.DeviceList(); err != nil {
		return nil, err
	}
	return devices, nil
}

func init() {
	androidRootCmd.AddCommand(listAndroidDevicesCmd)
}
