package hrp

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// StepRendezvous implements IStep interface.
type StepRendezvous struct {
	step *TStep
}

func (s *StepRendezvous) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return s.step.Rendezvous.Name
}

func (s *StepRendezvous) Type() StepType {
	return stepTypeRendezvous
}

func (s *StepRendezvous) Struct() *TStep {
	return s.step
}

func (s *StepRendezvous) Run(r *SessionRunner) (*StepResult, error) {
	rendezvous := s.step.Rendezvous
	log.Info().
		Str("name", rendezvous.Name).
		Float32("percent", rendezvous.Percent).
		Int64("number", rendezvous.Number).
		Int64("timeout", rendezvous.Timeout).
		Msg("rendezvous")

	stepResult := &StepResult{
		Name:     rendezvous.Name,
		StepType: stepTypeRendezvous,
		Success:  true,
	}

	// pass current rendezvous if already released, activate rendezvous sequentially after spawn done
	if rendezvous.isReleased() || !isPreRendezvousAllReleased(rendezvous, r.caseRunner.testCase.ToTCase()) || !rendezvous.isSpawnDone() {
		return stepResult, nil
	}

	// activate the rendezvous only once during each cycle
	rendezvous.once.Do(func() {
		close(rendezvous.activateChan)
	})

	// check current cnt using double check lock before updating to avoid negative WaitGroup counter
	if atomic.LoadInt64(&rendezvous.cnt) < rendezvous.Number {
		rendezvous.lock.Lock()
		if atomic.LoadInt64(&rendezvous.cnt) < rendezvous.Number {
			atomic.AddInt64(&rendezvous.cnt, 1)
			rendezvous.wg.Done()
			rendezvous.timerResetChan <- struct{}{}
		}
		rendezvous.lock.Unlock()
	}

	// block until current rendezvous released
	<-rendezvous.releaseChan
	return stepResult, nil
}

func isPreRendezvousAllReleased(rendezvous *Rendezvous, testCase *TCase) bool {
	for _, step := range testCase.TestSteps {
		preRendezvous := step.Rendezvous
		if preRendezvous == nil {
			continue
		}
		// meet current rendezvous, all previous rendezvous released, return true
		if preRendezvous == rendezvous {
			return true
		}
		if !preRendezvous.isReleased() {
			return false
		}
	}
	return true
}

// WithUserNumber sets the user number needed to release the current rendezvous
func (s *StepRendezvous) WithUserNumber(number int64) *StepRendezvous {
	s.step.Rendezvous.Number = number
	return s
}

// WithUserPercent sets the user percent needed to release the current rendezvous
func (s *StepRendezvous) WithUserPercent(percent float32) *StepRendezvous {
	s.step.Rendezvous.Percent = percent
	return s
}

// WithTimeout sets the timeout of duration between each user arriving at the current rendezvous
func (s *StepRendezvous) WithTimeout(timeout int64) *StepRendezvous {
	s.step.Rendezvous.Timeout = timeout
	return s
}

const (
	defaultRendezvousTimeout int64   = 5000
	defaultRendezvousPercent float32 = 1.0
)

type Rendezvous struct {
	Name           string  `json:"name" yaml:"name"`                           // required
	Percent        float32 `json:"percent,omitempty" yaml:"percent,omitempty"` // default to 1(100%)
	Number         int64   `json:"number,omitempty" yaml:"number,omitempty"`
	Timeout        int64   `json:"timeout,omitempty" yaml:"timeout,omitempty"` // milliseconds
	cnt            int64
	releasedFlag   uint32
	spawnDoneFlag  uint32
	wg             sync.WaitGroup
	timerResetChan chan struct{}
	activateChan   chan struct{}
	releaseChan    chan struct{}
	once           *sync.Once
	lock           sync.Mutex
}

func (r *Rendezvous) reset() {
	r.cnt = 0
	r.releasedFlag = 0
	r.wg.Add(int(r.Number))
	// timerResetChan channel will not be closed, thus init only once
	if r.timerResetChan == nil {
		r.timerResetChan = make(chan struct{})
	}
	r.activateChan = make(chan struct{})
	r.releaseChan = make(chan struct{})
	r.once = new(sync.Once)
}

func (r *Rendezvous) isSpawnDone() bool {
	return atomic.LoadUint32(&r.spawnDoneFlag) == 1
}

func (r *Rendezvous) setSpawnDone() {
	atomic.StoreUint32(&r.spawnDoneFlag, 1)
}

func (r *Rendezvous) isReleased() bool {
	return atomic.LoadUint32(&r.releasedFlag) == 1
}

func (r *Rendezvous) setReleased() {
	atomic.StoreUint32(&r.releasedFlag, 1)
}

func initRendezvous(testcase *TestCase, total int64) []*Rendezvous {
	var rendezvousList []*Rendezvous
	for _, s := range testcase.TestSteps {
		step := s.Struct()
		if step.Rendezvous == nil {
			continue
		}
		rendezvous := step.Rendezvous

		// either number or percent should be correctly put, otherwise set to default (total)
		if rendezvous.Number == 0 && rendezvous.Percent > 0 && rendezvous.Percent <= defaultRendezvousPercent {
			rendezvous.Number = int64(rendezvous.Percent * float32(total))
		} else if rendezvous.Number > 0 && rendezvous.Number <= total && rendezvous.Percent == 0 {
			rendezvous.Percent = float32(rendezvous.Number) / float32(total)
		} else {
			log.Warn().
				Str("name", rendezvous.Name).
				Int64("default number", total).
				Float32("default percent", defaultRendezvousPercent).
				Msg("rendezvous parameter not defined or error, set to default value")
			rendezvous.Number = total
			rendezvous.Percent = defaultRendezvousPercent
		}

		if rendezvous.Timeout <= 0 {
			rendezvous.Timeout = defaultRendezvousTimeout
		}

		rendezvous.reset()
		rendezvousList = append(rendezvousList, rendezvous)
	}
	return rendezvousList
}

func (r *Rendezvous) updateRendezvousNumber(number int64) {
	atomic.StoreInt64(&r.Number, int64(float32(number)*r.Percent))
}

func waitRendezvous(rendezvousList []*Rendezvous, b *HRPBoomer) {
	if rendezvousList != nil {
		lastRendezvous := rendezvousList[len(rendezvousList)-1]
		for _, rendezvous := range rendezvousList {
			go waitSingleRendezvous(rendezvous, rendezvousList, lastRendezvous, b)
		}
	}
}

func waitSingleRendezvous(rendezvous *Rendezvous, rendezvousList []*Rendezvous, lastRendezvous *Rendezvous, b *HRPBoomer) {
	for {
		// cycle start: block current checking until current rendezvous activated
		<-rendezvous.activateChan
		stop := make(chan struct{})
		timeout := time.Duration(rendezvous.Timeout) * time.Millisecond
		timer := time.NewTimer(timeout)
		go func() {
			defer close(stop)
			rendezvous.wg.Wait()
		}()
		for !rendezvous.isReleased() {
			select {
			case <-rendezvous.timerResetChan:
				timer.Reset(timeout)
			case <-stop:
				rendezvous.setReleased()
				close(rendezvous.releaseChan)
				log.Info().
					Str("name", rendezvous.Name).
					Float32("percent", rendezvous.Percent).
					Int64("number", rendezvous.Number).
					Int64("timeout(ms)", rendezvous.Timeout).
					Int64("cnt", rendezvous.cnt).
					Str("reason", "rendezvous release condition satisfied").
					Msg("rendezvous released")
			case <-timer.C:
				rendezvous.setReleased()
				close(rendezvous.releaseChan)
				log.Info().
					Str("name", rendezvous.Name).
					Float32("percent", rendezvous.Percent).
					Int64("number", rendezvous.Number).
					Int64("timeout(ms)", rendezvous.Timeout).
					Int64("cnt", rendezvous.cnt).
					Str("reason", "time's up").
					Msg("rendezvous released")
			}
		}
		// cycle end: reset all previous rendezvous after last rendezvous released
		// otherwise, block current checker until the last rendezvous end
		if rendezvous == lastRendezvous {
			for _, r := range rendezvousList {
				r.reset()
				// dynamic adjustment based on the number of concurrent users
				r.updateRendezvousNumber(int64(b.GetSpawnCount()))
			}
		} else {
			<-lastRendezvous.releaseChan
		}
	}
}
