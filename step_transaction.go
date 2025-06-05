package hrp

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type Transaction struct {
	Name string          `json:"name" yaml:"name"`
	Type TransactionType `json:"type" yaml:"type"`
}

type TransactionType string

const (
	TransactionStart TransactionType = "start"
	TransactionEnd   TransactionType = "end"
)

// StepTransaction implements IStep interface.
type StepTransaction struct {
	StepConfig
	Transaction *Transaction `json:"transaction,omitempty" yaml:"transaction,omitempty"`
}

func (s *StepTransaction) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	return fmt.Sprintf("transaction %s %s", s.Transaction.Name, s.Transaction.Type)
}

func (s *StepTransaction) Type() StepType {
	return StepTypeTransaction
}

func (s *StepTransaction) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepTransaction) Run(r *SessionRunner) (*StepResult, error) {
	transaction := s.Transaction
	log.Info().
		Str("name", transaction.Name).
		Str("type", string(transaction.Type)).
		Msg("transaction")

	stepResult := &StepResult{
		Name:        s.Name(),
		StepType:    s.Type(),
		Success:     true,
		Elapsed:     0,
		ContentSize: 0, // TODO: record transaction total response length
	}

	// create transaction if not exists
	if _, ok := r.transactions[transaction.Name]; !ok {
		r.transactions[transaction.Name] = make(map[TransactionType]time.Time)
	}

	// record transaction start time, override if already exists
	if transaction.Type == TransactionStart {
		r.transactions[transaction.Name][TransactionStart] = time.Now()
	}
	// record transaction end time, override if already exists
	if transaction.Type == TransactionEnd {
		r.transactions[transaction.Name][TransactionEnd] = time.Now()

		// if transaction start time not exists, use testcase start time instead
		if _, ok := r.transactions[transaction.Name][TransactionStart]; !ok {
			r.transactions[transaction.Name][TransactionStart] = r.summary.Time.StartAt
		}

		// calculate transaction duration
		duration := r.transactions[transaction.Name][TransactionEnd].Sub(
			r.transactions[transaction.Name][TransactionStart])
		stepResult.Elapsed = duration.Milliseconds()
		log.Info().Str("name", transaction.Name).Dur("elapsed", duration).Msg("transaction")
	}

	return stepResult, nil
}
