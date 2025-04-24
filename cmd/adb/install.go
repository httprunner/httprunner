package adb

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

var (
	replace   bool
	downgrade bool
	grant     bool
)

var installCmd = &cobra.Command{
	Use:   "install [flags] PACKAGE",
	Short: "push package to the device and install them automatically",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_adb_devices", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()
		_, err = getDevice(serial)
		if err != nil {
			return err
		}

		device, err := uixt.NewAndroidDevice(option.WithSerialNumber(serial))
		if err != nil {
			fmt.Println(err)
			return err
		}
		driver, err := device.NewDriver()
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = driver.GetDevice().Install(args[0],
			option.WithReinstall(replace),
			option.WithDowngrade(downgrade),
			option.WithGrantPermission(grant),
		)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("success")
		return nil
	},
}

func init() {
	installCmd.Flags().StringVarP(&serial, "serial", "s", "", "filter by device's serial")
	installCmd.Flags().BoolVarP(&replace, "replace", "r", false, "replace existing application")
	installCmd.Flags().BoolVarP(&downgrade, "downgrade", "d", false, "allow version code downgrade (debuggable packages only)")
	installCmd.Flags().BoolVarP(&grant, "grant", "g", false, "grant all runtime permissions")
	CmdAndroidRoot.AddCommand(installCmd)
}
