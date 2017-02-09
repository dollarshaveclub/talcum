package talcum

import (
	"fmt"
	"math/rand"
)

type Config struct {
	ApplicationName string
	ID              string
}

type Locker interface {
	Lock(key string) (bool, error)
}

type SelectorEntry struct {
	Value string `json:"value"`
	Num   int    `json:"num"`
}

type SelectorConfig []*SelectorEntry

func (s SelectorConfig) copyExcludingEntry(entry *SelectorEntry) SelectorConfig {
	var config SelectorConfig

	for _, e := range s {
		if entry != e {
			config = append(config, e)
		}
	}

	return config
}

type selectorRandomRange struct {
	entry      *SelectorEntry
	rangeStart float64
	rangeEnd   float64
}

func selectorRangeFromConfig(config SelectorConfig) []*selectorRandomRange {
	var totalNum int
	for _, entry := range config {
		totalNum += entry.Num
	}

	var rangeMarker float64
	var ranges []*selectorRandomRange
	for _, entry := range config {
		entryProportion := float64(entry.Num) / float64(totalNum)
		rangeEnd := rangeMarker + entryProportion
		ranges = append(ranges, &selectorRandomRange{
			entry:      entry,
			rangeStart: rangeMarker,
			rangeEnd:   rangeEnd,
		})
		rangeMarker = rangeEnd
	}

	// Ensure there aren't issues with floating point math.
	lastEntry := ranges[len(ranges)-1]
	lastEntry.rangeEnd = 1

	return ranges
}

func locateEntryByFloat(randomRange []*selectorRandomRange, f float64) *SelectorEntry {
	if f > 1 || f < 0 {
		panic("illegal float value")
	}

	for _, entry := range randomRange {
		if entry.rangeStart <= f && entry.rangeEnd < f {
			return entry.entry
		}
	}

	// We should never reach this.
	return randomRange[len(randomRange)-1].entry
}

type Selector struct {
	talcumConfig   *Config
	selectorConfig SelectorConfig
	locker         Locker
}

func NewSelector(config *Config, selectorConfig SelectorConfig, locker Locker) *Selector {
	return &Selector{
		talcumConfig:   config,
		selectorConfig: selectorConfig,
		locker:         locker,
	}
}

func (s *Selector) lockKey(entry *SelectorEntry, num int) string {
	return fmt.Sprintf("%s/%s/%v", s.talcumConfig.ApplicationName, s.talcumConfig.ID, num)
}

func (s *Selector) claimEntry(entry *SelectorEntry) (bool, error) {
	for i := 0; i < entry.Num; i++ {
		key := s.lockKey(entry, i)
		locked, err := s.locker.Lock(key)
		if err != nil {
			return false, err
		}
		if locked {
			return true, nil
		}
	}

	return false, nil
}

func (s *Selector) Select() (*SelectorEntry, error) {
	iterations := len(s.selectorConfig)
	selectorConfig := s.selectorConfig
	for i := 0; i < iterations; i++ {
		ranges := selectorRangeFromConfig(selectorConfig)
		entry := locateEntryByFloat(ranges, rand.Float64())

		claimedEntry, err := s.claimEntry(entry)
		if err != nil {
			return nil, err
		}
		if claimedEntry {
			return entry, nil
		}

		selectorConfig = selectorConfig.copyExcludingEntry(entry)
	}

	// If we couldn't claim anything, choose an entry randomly
	// with weights.
	ranges := selectorRangeFromConfig(s.selectorConfig)
	return locateEntryByFloat(ranges, rand.Float64()), nil
}
