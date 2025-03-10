package ios

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
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

		if unmountDeveloperDiskImage {
			err := device.UnmountImage()
			if err != nil {
				return fmt.Errorf("unmount developer disk image failed: %v", err)
			}
			return nil
		}

		images, errImage := device.ListImages()
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
			log.Info().Strs("images", images).Msg("ios developer image is already mounted")
			return nil
		}

		if err = device.AutoMountImage(developerDiskImageDir); err != nil {
			return fmt.Errorf("mount developer disk image failed: %s", err)
		}

		log.Info().Msg("mount developer disk image successfully")
		return nil
	},
}

var (
	developerDiskImageDir     string
	listDeveloperDiskImage    bool
	unmountDeveloperDiskImage bool
)

func init() {
	home, _ := os.UserHomeDir()
	defaultDeveloperDiskImageDir := path.Join(home, ".devimages")

	mountCmd.Flags().BoolVar(&listDeveloperDiskImage, "list", false, "list developer disk images")
	mountCmd.Flags().BoolVar(&unmountDeveloperDiskImage, "reset", false, "unmount developer disk images")
	mountCmd.Flags().StringVarP(&developerDiskImageDir, "dir", "d", defaultDeveloperDiskImageDir, "specify developer disk image directory")
	mountCmd.Flags().StringVarP(&udid, "udid", "u", "", "specify device by udid")
	iosRootCmd.AddCommand(mountCmd)
}
