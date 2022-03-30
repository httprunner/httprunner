package hrp

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type Transaction struct {
	Name string          `json:"name" yaml:"name"`
	Type transactionType `json:"type" yaml:"type"`
}

type transactionType string

const (
	transactionStart transactionType = "start"
	transactionEnd   transactionType = "end"
)

// StepTransaction implements IStep interface.
type StepTransaction struct {
	step *TStep
}

func (s *StepTransaction) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("transaction %s %s", s.step.Transaction.Name, s.step.Transaction.Type)
}

func (s *StepTransaction) Type() StepType {
	return stepTypeTransaction
}

func (s *StepTransaction) Struct() *TStep {
	return s.step
}

func (s *StepTransaction) Run(r *SessionRunner) (*StepResult, error) {
	transaction := s.step.Transaction
	log.Info().
		Str("name", transaction.Name).
		Str("type", string(transaction.Type)).
		Msg("transaction")

	stepResult := &StepResult{
		Name:        transaction.Name,
		StepType:    stepTypeTransaction,
		Success:     true,
		Elapsed:     0,
		ContentSize: 0, // TODO: record transaction total response length
	}

	// create transaction if not exists
	if _, ok := r.transactions[transaction.Name]; !ok {
		r.transactions[transaction.Name] = make(map[transactionType]time.Time)
	}

	// record transaction start time, override if already exists
	if transaction.Type == transactionStart {
		r.transactions[transaction.Name][transactionStart] = time.Now()
	}
	// record transaction end time, override if already exists
	if transaction.Type == transactionEnd {
		r.transactions[transaction.Name][transactionEnd] = time.Now()

		// if transaction start time not exists, use testcase start time instead
		if _, ok := r.transactions[transaction.Name][transactionStart]; !ok {
			r.transactions[transaction.Name][transactionStart] = r.startTime
		}

		// calculate transaction duration
		duration := r.transactions[transaction.Name][transactionEnd].Sub(
			r.transactions[transaction.Name][transactionStart])
		stepResult.Elapsed = duration.Milliseconds()
		log.Info().Str("name", transaction.Name).Dur("elapsed", duration).Msg("transaction")
	}

	return stepResult, nil
}
