package ios

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "tunnel start",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_ios_tunnel", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()
		ctx := context.TODO()
		err = uixt.StartTunnel(ctx, os.TempDir(), ios.HttpApiPort(), true)
		if err != nil {
			log.Error().Err(err).Msg("failed to start tunnel")
		}
		<-ctx.Done()
		return err
	},
}

func init() {
	CmdIOSRoot.AddCommand(tunnelCmd)
}
