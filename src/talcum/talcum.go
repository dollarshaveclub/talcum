package talcum

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// Config contains keys that are used when setting locks in order to
// namespace the current selection process.
type Config struct {
	ApplicationName string
	SelectionID     string
	LockDelay       time.Duration
	DebugMode       bool
}

// MetricsConfig contains options related to Datadog
type MetricsConfig struct {
	StatsdAddr string
	Namespace  string
	Tags       []string
	TagStr     string
}

// SelectorEntry contains the name of each role and the number of
// times it must be selected.
type SelectorEntry struct {
	RoleName string `json:"role_name"`
	Num      int    `json:"num"`
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
	hasher := sha256.New()
	hasher.Write([]byte(entry.RoleName))
	hasher.Write([]byte(strconv.Itoa(num)))

	return fmt.Sprintf("%s/%s/%x/%v",
		s.talcumConfig.ApplicationName,
		s.talcumConfig.SelectionID, hasher.Sum(nil)[:10], num)
}

// SelectRandom returns a random entry, weighing each entry using its
// expected number of occurrences as a weight.
func (s *Selector) SelectRandom() *SelectorEntry {
	max := len(s.selectorConfig.entryLocks())
	r := rand.Intn(max)
	return s.selectorConfig.entryLocks()[r].selectorEntry
}

// Select locks an entry and returns it. Select attempts to lock all
// entries up to the configured number. If all entries have been
// locked, a random one is returned.
func (s *Selector) Select() (*SelectorEntry, error) {
	entryLocks := shuffleEntryLocks(s.selectorConfig.entryLocks())

	for _, entryLock := range entryLocks {
		key := s.lockKey(entryLock.selectorEntry, entryLock.lockValue)

		if s.talcumConfig.DebugMode {
			log.Printf("Attempting to lock key: %s", key)
		}

		locked, err := s.locker.Lock(key)
		if err != nil {
			return nil, err
		}
		if locked {
			return entryLock.selectorEntry, nil
		}

		if s.talcumConfig.DebugMode {
			log.Printf("Could not lock key: %s", key)
		}

		// Optionally sleep as to not hammer the locking
		// backend.
		if s.talcumConfig.LockDelay > 0 {
			if s.talcumConfig.DebugMode {
				log.Printf("Sleeping before attempting to select new key")
			}

			time.Sleep(s.talcumConfig.LockDelay)
		}
	}

	if s.talcumConfig.DebugMode {
		log.Printf("All keys are locked, selecting random key")
	}

	// If we couldn't claim anything, choose an entry randomly.
	return s.SelectRandom(), nil
}
