package ios

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := getDevice(udid)
		if err != nil {
			return err
		}

		value, err := device.GetValue("", "ProductVersion")
		if err != nil {
			return fmt.Errorf("get device ProductVersion failed: %v", err)
		}
		log.Info().Str("version", value.(string)).Msg("get device version")

		imageSignatures, errImage := device.Images()

		if listDeveloperDiskImage {
			for i, imgSign := range imageSignatures {
				fmt.Printf("[%d] %s\n", i+1, base64.StdEncoding.EncodeToString(imgSign))
			}
			return nil
		}

		if errImage == nil && len(imageSignatures) > 0 {
			log.Info().Msg("ios developer image is already mounted")
			return nil
		}

		log.Info().Str("dir", developerDiskImageDir).Msg("start to mount ios developer image")

		if !builtin.IsFolderPathExists(developerDiskImageDir) {
			return fmt.Errorf("developer disk image directory not exist: %s", developerDiskImageDir)
		}

		ver := strings.Split(value.(string), ".")
		if len(ver) < 2 {
			return fmt.Errorf("got invalid device ProductVersion: %v", value)
		}
		version := ver[0] + "." + ver[1]

		var dmgPath, signaturePath string
		if builtin.IsFilePathExists(filepath.Join(developerDiskImageDir, "DeveloperDiskImage.dmg")) {
			dmgPath = filepath.Join(developerDiskImageDir, "DeveloperDiskImage.dmg")
			signaturePath = filepath.Join(developerDiskImageDir, "DeveloperDiskImage.dmg.signature")
		} else if builtin.IsFilePathExists(filepath.Join(developerDiskImageDir, version, "DeveloperDiskImage.dmg")) {
			dmgPath = filepath.Join(developerDiskImageDir, version, "DeveloperDiskImage.dmg")
			signaturePath = filepath.Join(developerDiskImageDir, version, "DeveloperDiskImage.dmg.signature")
		} else {
			log.Error().Str("dir", developerDiskImageDir).Msgf(
				"developer disk image %s not found in directory", version)
			return fmt.Errorf("developer disk image %s not found", version)
		}

		if err = device.MountDeveloperDiskImage(dmgPath, signaturePath); err != nil {
			return fmt.Errorf("mount developer disk image %s failed: %s", version, err)
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
