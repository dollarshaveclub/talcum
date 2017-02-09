package talcum

import (
	"fmt"
	"math/rand"
)

// Config contains keys that are used when setting locks in order to
// namespace the current selection process.
type Config struct {
	ApplicationName string
	ID              string
}

// SelectorEntry contains the value of each selectable thing and the
// number of times it must be selected.
type SelectorEntry struct {
	Value string `json:"value"`
	Num   int    `json:"num"`
}

// SelectorConfig all selectable entries.
type SelectorConfig []*SelectorEntry

type entryLock struct {
	selectorEntry *SelectorEntry
	lockValue     int
}

func (s SelectorConfig) entryLocks() []*entryLock {
	var locks []*entryLock
	for _, entry := range s {
		for i := 0; i < entry.Num; i++ {
			locks = append(locks, &entryLock{
				selectorEntry: entry,
				lockValue:     i,
			})
		}
	}
	return locks
}

func shuffleEntryLocks(locks []*entryLock) []*entryLock {
	var shuffledLocks []*entryLock

	// This algorithm could potentially not terminate, but the
	// average running time is O(len(locks)).
	seen := make(map[int]bool)
	for {
		val := rand.Intn(len(locks))
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = true

		shuffledLocks = append(shuffledLocks, locks[val])

		if len(seen) == len(locks) {
			break
		}
	}

	return shuffledLocks
}

// Locker can set a lock for an entry.
type Locker interface {
	Lock(key string) (bool, error)
}

// Selector can select one of the entries it is configured to
// track. Each entry is configured to be used `n` times before it can
// be chosen randomly.
type Selector struct {
	talcumConfig   *Config
	selectorConfig SelectorConfig
	locker         Locker
}

// NewSelector creates a new selector.
func NewSelector(config *Config, selectorConfig SelectorConfig, locker Locker) *Selector {
	return &Selector{
		talcumConfig:   config,
		selectorConfig: selectorConfig,
		locker:         locker,
	}
}

func (s *Selector) lockKey(entry *SelectorEntry, num int) string {
	return fmt.Sprintf("%s/%s/%s/%v",
		s.talcumConfig.ApplicationName,
		s.talcumConfig.ID, entry.Value, num)
}

// SelectRandom returns a random entry.
func (s *Selector) SelectRandom() *SelectorEntry {
	return s.selectorConfig.entryLocks()[rand.Intn(len(s.selectorConfig))].selectorEntry
}

// Select locks an entry and returns it. Select attempts to lock all
// entries up to the configured number. If all entries have been
// locked, a random one is returned.
func (s *Selector) Select() (*SelectorEntry, error) {
	entryLocks := shuffleEntryLocks(s.selectorConfig.entryLocks())

	for _, entryLock := range entryLocks {
		locked, err := s.locker.Lock(s.lockKey(entryLock.selectorEntry, entryLock.lockValue))
		if err != nil {
			return nil, err
		}
		if locked {
			return entryLock.selectorEntry, nil
		}
	}

	// If we couldn't claim anything, choose an entry randomly.
	return s.SelectRandom(), nil
}
