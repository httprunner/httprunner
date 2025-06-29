package uixt

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// CacheManager provides a generic cache management interface
type CacheManager[T any] struct {
	cache   sync.Map
	name    string        // cache name for logging
	cleanup func(T) error // cleanup function for cached items
}

// NewCacheManager creates a new cache manager
func NewCacheManager[T any](name string, cleanup func(T) error) *CacheManager[T] {
	return &CacheManager[T]{
		cache:   sync.Map{},
		name:    name,
		cleanup: cleanup,
	}
}

// CachedItem wraps an item with cache metadata
type CachedItem[T any] struct {
	Key      string
	Item     T
	RefCount int32
	Metadata map[string]interface{} // additional metadata
}

// Get retrieves an item from cache
func (cm *CacheManager[T]) Get(key string) (*CachedItem[T], bool) {
	if item, ok := cm.cache.Load(key); ok {
		if cached, ok := item.(*CachedItem[T]); ok {
			cached.RefCount++
			log.Debug().
				Str("cache", cm.name).
				Str("key", key).
				Int32("refCount", cached.RefCount).
				Msg("Retrieved item from cache")
			return cached, true
		}
	}
	return nil, false
}

// Set stores an item in cache
func (cm *CacheManager[T]) Set(key string, item T, metadata map[string]interface{}) *CachedItem[T] {
	cached := &CachedItem[T]{
		Key:      key,
		Item:     item,
		RefCount: 1,
		Metadata: metadata,
	}

	cm.cache.Store(key, cached)
	log.Debug().
		Str("cache", cm.name).
		Str("key", key).
		Msg("Stored item in cache")

	return cached
}

// Release decrements reference count and removes item if count reaches zero
func (cm *CacheManager[T]) Release(key string) error {
	if item, ok := cm.cache.Load(key); ok {
		if cached, ok := item.(*CachedItem[T]); ok {
			cached.RefCount--
			log.Debug().
				Str("cache", cm.name).
				Str("key", key).
				Int32("refCount", cached.RefCount).
				Msg("Released item reference")

			// If no more references, clean up and remove from cache
			if cached.RefCount <= 0 {
				cm.cache.Delete(key)

				// Clean up item if cleanup function is provided
				if cm.cleanup != nil {
					if err := cm.cleanup(cached.Item); err != nil {
						log.Warn().Err(err).
							Str("cache", cm.name).
							Str("key", key).
							Msg("Failed to cleanup cached item")
						return err
					}
				}

				log.Info().
					Str("cache", cm.name).
					Str("key", key).
					Msg("Cleaned up item from cache")
			}
		}
	}
	return nil
}

// Clear removes all items from cache
func (cm *CacheManager[T]) Clear() {
	cm.cache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if cached, ok := value.(*CachedItem[T]); ok {
				// Clean up item if cleanup function is provided
				if cm.cleanup != nil {
					if err := cm.cleanup(cached.Item); err != nil {
						log.Warn().Err(err).
							Str("cache", cm.name).
							Str("key", keyStr).
							Msg("Failed to cleanup cached item")
					}
				}
				log.Debug().
					Str("cache", cm.name).
					Str("key", keyStr).
					Msg("Cleaned up item from cache")
			}
			cm.cache.Delete(key)
		}
		return true
	})
	log.Info().Str("cache", cm.name).Msg("Cleared all cached items")
}

// List returns all cached items
func (cm *CacheManager[T]) List() []CachedItem[T] {
	var items []CachedItem[T]
	cm.cache.Range(func(key, value interface{}) bool {
		if cached, ok := value.(*CachedItem[T]); ok {
			items = append(items, *cached)
		}
		return true
	})
	return items
}

// GetOrCreate gets an existing item or creates a new one using the provided factory function
func (cm *CacheManager[T]) GetOrCreate(key string, factory func() (T, map[string]interface{}, error)) (T, error) {
	// Check cache first
	if cached, ok := cm.Get(key); ok {
		return cached.Item, nil
	}

	// Create new item
	item, metadata, err := factory()
	if err != nil {
		var zero T
		return zero, err
	}

	// Store in cache
	cached := cm.Set(key, item, metadata)
	return cached.Item, nil
}

// Size returns the number of items in cache
func (cm *CacheManager[T]) Size() int {
	count := 0
	cm.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Use cache manager for XTDriver caching
var driverCacheManager = NewCacheManager("xt-driver", cleanupXTDriver)

// cleanupXTDriver cleans up XTDriver resources
func cleanupXTDriver(driver *XTDriver) error {
	if driver != nil && driver.IDriver != nil {
		if err := driver.DeleteSession(); err != nil {
			log.Warn().Err(err).Msg("Failed to delete driver session during cleanup")
			return err
		}
	}
	return nil
}

// CachedXTDriver is an alias for CachedItem[*XTDriver] for backward compatibility
type CachedXTDriver = CachedItem[*XTDriver]

// DriverCacheConfig holds configuration for driver creation
type DriverCacheConfig struct {
	Platform   string
	Serial     string
	AIOptions  []option.AIServiceOption
	DeviceOpts *option.DeviceOptions // unified device options
}

// GetOrCreateXTDriver gets an existing driver from cache or creates a new one
func GetOrCreateXTDriver(config DriverCacheConfig) (*XTDriver, error) {
	// Handle empty serial case - try to find existing driver first
	if config.Serial == "" {
		if driver := findCachedDriver(config.Platform); driver != nil {
			return driver, nil
		}
	}

	// Use shared cache manager's GetOrCreate functionality
	return driverCacheManager.GetOrCreate(config.Serial, func() (*XTDriver, map[string]interface{}, error) {
		// Create new driver
		driverExt, err := createXTDriverWithConfig(config)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create XTDriver: %w", err)
		}

		// Get actual serial from the created driver
		actualSerial := driverExt.GetDevice().UUID()

		// Check if a driver with actual serial already exists (for empty serial case)
		if config.Serial == "" && actualSerial != "" {
			if existingCached, ok := driverCacheManager.Get(actualSerial); ok {
				// Clean up the newly created driver since we have a cached one
				if err := driverExt.DeleteSession(); err != nil {
					log.Warn().Err(err).Str("serial", actualSerial).Msg("Failed to delete newly created driver session")
				}
				return existingCached.Item, existingCached.Metadata, nil
			}
		}

		// Create metadata
		metadata := map[string]interface{}{
			"platform": config.Platform,
			"serial":   actualSerial,
		}

		log.Info().
			Str("platform", config.Platform).
			Str("serial", actualSerial).
			Msg("Created and cached new XTDriver")

		return driverExt, metadata, nil
	})
}

// createXTDriverWithConfig creates a new XTDriver based on configuration
func createXTDriverWithConfig(config DriverCacheConfig) (*XTDriver, error) {
	platform := config.Platform
	if platform == "" {
		log.Warn().Msg("platform is not set, using android as default")
		platform = "android"
	}

	// Create device based on platform and configuration
	var device IDevice
	var err error

	// Create device based on platform and configuration
	if config.DeviceOpts != nil {
		// Use specific device options
		switch strings.ToLower(platform) {
		case "android":
			androidOpts := config.DeviceOpts.ToAndroidOptions().Options()
			device, err = NewAndroidDevice(androidOpts...)
		case "ios":
			iosOpts := config.DeviceOpts.ToIOSOptions().Options()
			device, err = NewIOSDevice(iosOpts...)
		case "harmony":
			harmonyOpts := config.DeviceOpts.ToHarmonyOptions().Options()
			device, err = NewHarmonyDevice(harmonyOpts...)
		case "browser":
			browserOpts := config.DeviceOpts.ToBrowserOptions().Options()
			device, err = NewBrowserDevice(browserOpts...)
		default:
			return nil, errors.Wrapf(code.InvalidParamError, "unsupported platform: %s", platform)
		}
	} else {
		// Use default options, let NewXXDevice handle serial (empty or specified)
		switch strings.ToLower(platform) {
		case "android":
			if config.Serial != "" {
				device, err = NewAndroidDevice(option.WithSerialNumber(config.Serial))
			} else {
				device, err = NewAndroidDevice()
			}
		case "ios":
			if config.Serial != "" {
				device, err = NewIOSDevice(option.WithUDID(config.Serial))
			} else {
				device, err = NewIOSDevice()
			}
		case "harmony":
			if config.Serial != "" {
				device, err = NewHarmonyDevice(option.WithConnectKey(config.Serial))
			} else {
				device, err = NewHarmonyDevice()
			}
		case "browser":
			if config.Serial != "" {
				device, err = NewBrowserDevice(option.WithBrowserID(config.Serial))
			} else {
				device, err = NewBrowserDevice()
			}
		default:
			return nil, errors.Wrapf(code.InvalidParamError, "unsupported platform: %s", platform)
		}
	}
	if err != nil {
		return nil, err
	}

	// Create driver
	driver, err := device.NewDriver()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create driver")
	}

	// Create XTDriver with AI options
	aiOpts := config.AIOptions
	if len(aiOpts) == 0 {
		// Default AI options
		aiOpts = []option.AIServiceOption{
			option.WithCVService(option.CVServiceTypeVEDEM),
			option.WithLLMConfig(option.RecommendedConfigurations()["ui_focused"]),
		}
	}

	driverExt, err := NewXTDriver(driver, aiOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create XTDriver")
	}
	return driverExt, nil
}

// ReleaseXTDriver decrements reference count and removes from cache when count reaches zero
func ReleaseXTDriver(serial string) error {
	return driverCacheManager.Release(serial)
}

// CleanupAllDrivers cleans up all cached drivers
func CleanupAllDrivers() {
	driverCacheManager.Clear()
}

// ListCachedDrivers returns information about all cached drivers
func ListCachedDrivers() []CachedXTDriver {
	return driverCacheManager.List()
}

// findCachedDriver searches for a cached driver by platform
// If platform is empty, returns any available driver
func findCachedDriver(platform string) *XTDriver {
	cachedItems := driverCacheManager.List()

	for _, cachedItem := range cachedItems {
		cachedPlatform, _ := cachedItem.Metadata["platform"].(string)

		// If platform is specified, match platform; otherwise use any available driver
		if platform == "" || cachedPlatform == platform {
			// Increment reference count by getting from cache
			if refreshedItem, ok := driverCacheManager.Get(cachedItem.Key); ok {
				if platform != "" {
					log.Debug().Str("platform", platform).Str("serial", cachedItem.Key).Msg("Using cached XTDriver by platform")
				} else {
					log.Debug().Str("serial", cachedItem.Key).Msg("Using any available cached XTDriver")
				}
				return refreshedItem.Item
			}
		}
	}
	return nil
}

// setupXTDriver initializes an XTDriver based on the platform and serial.
// This function is kept for backward compatibility with MCP integration
func setupXTDriver(_ context.Context, args map[string]any) (*XTDriver, error) {
	platform, _ := args["platform"].(string)
	serial, _ := args["serial"].(string)

	// Extract AI service options from arguments if provided
	var aiOpts []option.AIServiceOption

	// Check for LLM service type
	if llmService, ok := args["llm_service"].(string); ok && llmService != "" {
		aiOpts = append(aiOpts, option.WithLLMService(option.LLMServiceType(llmService)))
	}

	// Check for CV service type
	if cvService, ok := args["cv_service"].(string); ok && cvService != "" {
		aiOpts = append(aiOpts, option.WithCVService(option.CVServiceType(cvService)))
	}

	config := DriverCacheConfig{
		Platform:  platform,
		Serial:    serial,
		AIOptions: aiOpts,
	}
	return GetOrCreateXTDriver(config)
}

// RegisterXTDriver registers an externally created XTDriver to the unified cache
func RegisterXTDriver(serial string, driver *XTDriver) error {
	if serial == "" {
		return fmt.Errorf("serial cannot be empty")
	}
	if driver == nil {
		return fmt.Errorf("driver cannot be nil")
	}

	// Create metadata
	metadata := map[string]interface{}{
		"platform": "external", // Mark as externally registered
		"serial":   serial,
	}

	// Store in cache using shared cache manager
	driverCacheManager.Set(serial, driver, metadata)

	log.Info().
		Str("serial", serial).
		Msg("Registered external XTDriver to unified cache")

	return nil
}

// getXTDriverFromCache gets XTDriver from cache using device UUID
func getXTDriverFromCache(driver IDriver) *XTDriver {
	// Get device info to find the corresponding XTDriver
	device := driver.GetDevice()
	if device == nil {
		log.Warn().Msg("Cannot get device from driver for MCP hook")
		return nil
	}

	// Get device UUID (serial/udid/connectKey/browserID)
	deviceUUID := device.UUID()
	if deviceUUID == "" {
		log.Warn().Msg("Cannot get device UUID for MCP hook")
		return nil
	}

	// Get XTDriver from cache using device UUID as serial
	cachedDrivers := ListCachedDrivers()
	for _, cached := range cachedDrivers {
		if serial, _ := cached.Metadata["serial"].(string); serial == deviceUUID {
			return cached.Item
		}
	}

	log.Warn().Str("uuid", deviceUUID).
		Msg("Cannot find cached XTDriver for MCP hook")
	return nil
}
