package ios

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
)

var xctestCmd = &cobra.Command{
	Use:   "xctest",
	Short: "run xctest",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_xctest", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		if bundleID == "" {
			return fmt.Errorf("bundleID is required")
		}
		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		err = device.RunXCTest(context.Background(), bundleID, testRunnerBundleID, xctestConfig)
		if err != nil {
			return errors.Wrap(err, "run xctest failed")
		}
		return nil
	},
}

var (
	bundleID           string
	testRunnerBundleID string
	xctestConfig       string
)

func init() {
	xctestCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify ios device's UDID")
	xctestCmd.Flags().StringVarP(&bundleID, "bundleID", "b", "com.gtf.wda.runner.xctrunner", "specify ios bundleID")
	xctestCmd.Flags().StringVarP(&testRunnerBundleID, "testRunnerBundleID", "t", "com.gtf.wda.runner.xctrunner", "specify ios testRunnerBundleID")
	xctestCmd.Flags().StringVarP(&xctestConfig, "xctestConfig", "x", "GtfWdaRunner.xctest", "specify ios xctestConfig")
	CmdIOSRoot.AddCommand(xctestCmd)
}
