package hrp

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
)

type ThinkTime struct {
	Time float64 `json:"time" yaml:"time"`
}

// StepThinkTime implements IStep interface.
type StepThinkTime struct {
	StepConfig
	ThinkTime *ThinkTime `json:"think_time,omitempty" yaml:"think_time,omitempty"`
}

func (s *StepThinkTime) Name() string {
	return s.StepName
}

func (s *StepThinkTime) Type() StepType {
	return StepTypeThinkTime
}

func (s *StepThinkTime) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepThinkTime) Run(r *SessionRunner) (*StepResult, error) {
	thinkTime := s.ThinkTime
	log.Info().Float64("time", thinkTime.Time).Msg("think time")

	stepResult := &StepResult{
		Name:     s.StepName,
		StepType: StepTypeThinkTime,
		Success:  true,
	}

	cfg := r.caseRunner.Config.Get().ThinkTimeSetting
	if cfg == nil {
		cfg = &ThinkTimeConfig{ThinkTimeDefault, nil, 0}
	}

	var tt time.Duration
	switch cfg.Strategy {
	case ThinkTimeDefault:
		tt = time.Duration(thinkTime.Time*1000) * time.Millisecond
	case ThinkTimeRandomPercentage:
		// e.g. {"min_percentage": 0.5, "max_percentage": 1.5}
		m, ok := cfg.Setting.(map[string]float64)
		if !ok {
			tt = time.Duration(thinkTime.Time*1000) * time.Millisecond
			break
		}
		res := builtin.GetRandomNumber(int(thinkTime.Time*m["min_percentage"]*1000), int(thinkTime.Time*m["max_percentage"]*1000))
		tt = time.Duration(res) * time.Millisecond
	case ThinkTimeMultiply:
		value, ok := cfg.Setting.(float64) // e.g. 0.5
		if !ok || value <= 0 {
			value = thinkTimeDefaultMultiply
		}
		tt = time.Duration(thinkTime.Time*value*1000) * time.Millisecond
	case ThinkTimeIgnore:
		// nothing to do
	}

	// no more than limit
	if cfg.Limit > 0 {
		limit := time.Duration(cfg.Limit*1000) * time.Millisecond
		if limit < tt {
			tt = limit
		}
	}

	// Use interruptible sleep that can respond to signals
	log.Debug().Float64("duration_ms", float64(tt.Milliseconds())).Msg("starting think time")

	select {
	case <-time.After(tt):
		// Normal completion
		log.Debug().Float64("duration_ms", float64(tt.Milliseconds())).Msg("think time completed normally")
	case <-r.caseRunner.hrpRunner.interruptSignal:
		// Interrupted by signal
		log.Info().Float64("planned_duration_ms", float64(tt.Milliseconds())).
			Msg("think time interrupted by signal")
		return stepResult, fmt.Errorf("think time interrupted")
	}

	return stepResult, nil
}
