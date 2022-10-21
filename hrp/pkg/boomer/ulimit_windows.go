//go:build windows

package boomer

import (
	"github.com/rs/zerolog/log"
)

// set resource limit
func SetUlimit(limit uint64) {
	log.Warn().Msg("windows does not support setting ulimit")
}
