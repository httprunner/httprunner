package uixt

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

var driverCache sync.Map // key is serial, value is *XTDriver

// setupXTDriver initializes an XTDriver based on the platform and serial.
func setupXTDriver(_ context.Context, args map[string]any) (*XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)
	if platform == "" {
		log.Warn().Msg("platform is not set, using android as default")
		platform = "android"
	}

	// Check if driver exists in cache
	cacheKey := serial
	if cachedDriver, ok := driverCache.Load(cacheKey); ok {
		if driverExt, ok := cachedDriver.(*XTDriver); ok {
			log.Info().Str("platform", platform).Str("serial", serial).Msg("Using cached driver")
			return driverExt, nil
		}
	}

	driverExt, err := NewXTDriverWithDefault(platform, serial)
	if err != nil {
		return nil, err
	}
	// store driver in cache
	driverCache.Store(cacheKey, driverExt)
	return driverExt, nil
}
