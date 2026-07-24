package sen6x

import (
	"context"
	"log"
	"time"
)

const (
	statePersistenceInterval = 30 * time.Minute
)

func (s *SEN6x) persistStateLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.i2cBusMu.Lock()
			if err := s.saveState(); err != nil {
				log.Printf("%s: failed to save state: %v", Name, err)
			}
			s.i2cBusMu.Unlock()
		}
	}
}

// saveState persists the sensor's state. The I2C bus lock must be held when this is called.
func (s *SEN6x) saveState() error {
	// TODO(mtraver) Only do this for SEN65, SEN66, SEN68, and SEN69C.
	state, err := s.dev.GetVOCAlgorithmState()
	if err != nil {
		return err
	}

	return s.vocAlgStateStore.Save(state[:])
}
