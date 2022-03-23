package boomer

import (
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"os"
	"runtime/pprof"
	"time"

	"github.com/rs/zerolog/log"
)

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// genMD5 returns the md5 hash of strings.
func genMD5(slice ...string) string {
	h := md5.New()
	for _, v := range slice {
		io.WriteString(h, v)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// startMemoryProfile starts memory profiling and save the results in file.
func startMemoryProfile(file string, duration time.Duration) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	log.Info().Dur("duration", duration).Msg("Start memory profiling")
	time.AfterFunc(duration, func() {
		err := pprof.WriteHeapProfile(f)
		if err != nil {
			log.Error().Err(err).Msg("failed to write memory profile")
		}
		f.Close()
		log.Info().Dur("duration", duration).Msg("Stop memory profiling")
	})
	return nil
}

// startCPUProfile starts cpu profiling and save the results in file.
func startCPUProfile(file string, duration time.Duration) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	log.Info().Dur("duration", duration).Msg("Start CPU profiling")
	err = pprof.StartCPUProfile(f)
	if err != nil {
		f.Close()
		return err
	}

	time.AfterFunc(duration, func() {
		pprof.StopCPUProfile()
		f.Close()
		log.Info().Dur("duration", duration).Msg("Stop CPU profiling")
	})
	return nil
}
