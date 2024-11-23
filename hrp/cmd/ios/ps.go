package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

var psCmd = &cobra.Command{
	Use:              "ps",
	Short:            "show running processes",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_ps", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		runningProcesses, err := device.ListProcess(!isAll)
		if err != nil {
			return err
		}
		for _, p := range runningProcesses {
			fmt.Printf("%4d %-"+fmt.Sprintf("%d", len(runningProcesses))+"s %20s %s\n",
				p.Pid, p.Name, time.Since(p.StartDate).String(), bundleID)
		}
		return nil
	},
}

var isAll bool

func init() {
	psCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	psCmd.Flags().BoolVarP(&isAll, "all", "a", false, "print all processes including system processes")
	iosRootCmd.AddCommand(psCmd)
}
