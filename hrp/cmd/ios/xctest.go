package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
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

		log.Info().Str("bundleID", bundleID).Msg("run xctest")
		err = device.RunXCTest(bundleID, testRunnerBundleID, xctestConfig)
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
	xctestCmd.Flags().StringVarP(&udid, "udid", "u", "", "filter by device's udid")
	xctestCmd.Flags().StringVarP(&bundleID, "bundleID", "b", "", "specify ios bundleID")
	xctestCmd.Flags().StringVarP(&testRunnerBundleID, "testRunnerBundleID", "t", "", "specify ios testRunnerBundleID")
	xctestCmd.Flags().StringVarP(&xctestConfig, "xctestConfig", "x", "", "specify ios xctestConfig")
	iosRootCmd.AddCommand(xctestCmd)
}
