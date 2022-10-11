//go:build !windows

package boomer

import (
	"syscall"

	"github.com/rs/zerolog/log"
)

// set resource limit
// ulimit -n 10240
func SetUlimit(limit uint64) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Error().Err(err).Msg("get ulimit failed")
		return
	}
	log.Info().Uint64("limit", rLimit.Cur).Msg("get current ulimit")
	if rLimit.Cur >= limit {
		return
	}

	rLimit.Cur = limit
	rLimit.Max = limit
	log.Info().Uint64("limit", rLimit.Cur).Msg("set current ulimit")
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Error().Err(err).Msg("set ulimit failed")
		return
	}
}
