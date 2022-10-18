package hrp

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

type ThinkTime struct {
	Time float64 `json:"time" yaml:"time"`
}

// StepThinkTime implements IStep interface.
type StepThinkTime struct {
	step *TStep
}

func (s *StepThinkTime) Name() string {
	return s.step.Name
}

func (s *StepThinkTime) Type() StepType {
	return stepTypeThinkTime
}

func (s *StepThinkTime) Struct() *TStep {
	return s.step
}

func (s *StepThinkTime) Run(r *SessionRunner) (*StepResult, error) {
	thinkTime := s.step.ThinkTime
	log.Info().Float64("time", thinkTime.Time).Msg("think time")

	stepResult := &StepResult{
		Name:     s.step.Name,
		StepType: stepTypeThinkTime,
		Success:  true,
	}

	cfg := r.caseRunner.parsedConfig.ThinkTimeSetting
	if cfg == nil {
		cfg = &ThinkTimeConfig{thinkTimeDefault, nil, 0}
	}

	var tt time.Duration
	switch cfg.Strategy {
	case thinkTimeDefault:
		tt = time.Duration(thinkTime.Time*1000) * time.Millisecond
	case thinkTimeRandomPercentage:
		// e.g. {"min_percentage": 0.5, "max_percentage": 1.5}
		m, ok := cfg.Setting.(map[string]float64)
		if !ok {
			tt = time.Duration(thinkTime.Time*1000) * time.Millisecond
			break
		}
		res := builtin.GetRandomNumber(int(thinkTime.Time*m["min_percentage"]*1000), int(thinkTime.Time*m["max_percentage"]*1000))
		tt = time.Duration(res) * time.Millisecond
	case thinkTimeMultiply:
		value, ok := cfg.Setting.(float64) // e.g. 0.5
		if !ok || value <= 0 {
			value = thinkTimeDefaultMultiply
		}
		tt = time.Duration(thinkTime.Time*value*1000) * time.Millisecond
	case thinkTimeIgnore:
		// nothing to do
	}

	// no more than limit
	if cfg.Limit > 0 {
		limit := time.Duration(cfg.Limit*1000) * time.Millisecond
		if limit < tt {
			tt = limit
		}
	}

	time.Sleep(tt)
	return stepResult, nil
}
