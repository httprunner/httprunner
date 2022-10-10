package ios

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:              "ps",
	Short:            "show running processes",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		apps, err := device.AppList()
		if err != nil {
			return errors.Wrap(err, "get ios apps failed")
		}

		maxNameLen := 0
		mapper := make(map[string]interface{})
		for _, app := range apps {
			mapper[app.ExecutableName] = app.CFBundleIdentifier
			if len(app.ExecutableName) > maxNameLen {
				maxNameLen = len(app.ExecutableName)
			}
		}

		runningProcesses, err := device.AppRunningProcesses()
		if err != nil {
			return errors.Wrap(err, "get running processes failed")
		}
		for _, p := range runningProcesses {
			if !isAll && !p.IsApplication {
				continue
			}
			bundleID, ok := mapper[p.Name]
			if !ok {
				bundleID = ""
			}

			fmt.Printf("%4d %-"+fmt.Sprintf("%d", maxNameLen)+"s %20s %s\n",
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
