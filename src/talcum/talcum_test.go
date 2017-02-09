package talcum_test

import (
	"strconv"
	"testing"

	"github.com/dollarshaveclub/talcum/src/talcum"
)

type mockLocker struct {
	lockedKeys map[string]bool
}

func newMockLocker() *mockLocker {
	return &mockLocker{
		lockedKeys: make(map[string]bool),
	}
}

func (m *mockLocker) Lock(key string) (bool, error) {
	if _, ok := m.lockedKeys[key]; ok {
		return false, nil
	}
	m.lockedKeys[key] = true
	return true, nil
}

func TestSelectSmoke(t *testing.T) {
	talcumConfig := &talcum.Config{
		ApplicationName: "test-app",
		SelectionID:     "test-id",
	}
	selectorConfig := []*talcum.SelectorEntry{
		{
			Value: "1",
			Num:   1,
		},
		{
			Value: "2",
			Num:   2,
		},
	}
	locker := newMockLocker()

	selector := talcum.NewSelector(talcumConfig, selectorConfig, locker)

	for n := 0; n < 100; n++ {
		_, err := selector.Select()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSelectChoosesAllEntries(t *testing.T) {
	talcumConfig := &talcum.Config{
		ApplicationName: "test-app",
		SelectionID:     "test-id",
	}
	var selectorConfig []*talcum.SelectorEntry

	for n := 1; n <= 10; n++ {
		selectorConfig = append(selectorConfig, &talcum.SelectorEntry{
			Value: strconv.Itoa(n),
			Num:   n,
		})
	}

	for j := 0; j < 10; j++ {
		locker := newMockLocker()
		selector := talcum.NewSelector(talcumConfig, selectorConfig, locker)
		seen := make(map[string]int)

		for i := 0; i < 55; i++ {
			entry, err := selector.Select()
			if err != nil {
				t.Fatal(err)
			}
			seen[entry.Value]++
		}

		for _, entry := range selectorConfig {
			numSeen := seen[entry.Value]
			if numSeen != entry.Num {
				t.Fatalf("expected: %d, seen: %d", entry.Num, numSeen)
			}
		}
	}
}

func TestSelectChoosesAllEntriesAtLeastMin(t *testing.T) {
	talcumConfig := &talcum.Config{
		ApplicationName: "test-app",
		SelectionID:     "test-id",
	}
	var selectorConfig []*talcum.SelectorEntry

	for n := 1; n <= 10; n++ {
		selectorConfig = append(selectorConfig, &talcum.SelectorEntry{
			Value: strconv.Itoa(n),
			Num:   n,
		})
	}

	for j := 0; j < 10; j++ {
		locker := newMockLocker()
		selector := talcum.NewSelector(talcumConfig, selectorConfig, locker)
		seen := make(map[string]int)

		for i := 0; i < 100; i++ {
			entry, err := selector.Select()
			if err != nil {
				t.Fatal(err)
			}
			seen[entry.Value]++
		}

		for _, entry := range selectorConfig {
			numSeen := seen[entry.Value]
			if numSeen < entry.Num {
				t.Fatalf("expected at least: %v", entry.Num)
			}
		}
	}
}
