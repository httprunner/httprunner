package ios

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_mount", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()

		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		images, errImage := device.ListImage()
		if err != nil {
			return fmt.Errorf("list device images failed: %v", err)
		}
		if listDeveloperDiskImage {
			for i, imgSign := range images {
				fmt.Printf("[%d] %s\n", i+1, imgSign)
			}
			return nil
		}

		if errImage == nil && len(images) > 0 {
			log.Info().Msg("ios developer image is already mounted")
			return nil
		}

		log.Info().Str("dir", developerDiskImageDir).Msg("start to mount ios developer image")

		if !builtin.IsFolderPathExists(developerDiskImageDir) {
			return fmt.Errorf("developer disk image directory not exist: %s", developerDiskImageDir)
		}

		if err = device.MountImage(developerDiskImageDir); err != nil {
			return fmt.Errorf("mount developer disk image failed: %s", err)
		}

		log.Info().Msg("mount developer disk image successfully")
		return nil
	},
}

const defaultDeveloperDiskImageDir = "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/DeviceSupport/"

var (
	developerDiskImageDir  string
	listDeveloperDiskImage bool
)

func init() {
	mountCmd.Flags().BoolVar(&listDeveloperDiskImage, "list", false, "list developer disk images")
	mountCmd.Flags().StringVarP(&developerDiskImageDir, "dir", "d", defaultDeveloperDiskImageDir, "specify DeveloperDiskImage directory")
	mountCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	iosRootCmd.AddCommand(mountCmd)
}
