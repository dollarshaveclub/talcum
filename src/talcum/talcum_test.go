package talcum_test

import (
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

func TestSelectSmokeStyle(t *testing.T) {
	talcumConfig := &talcum.Config{
		ApplicationName: "test-app",
		ID:              "test-id",
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

	entry, err := selector.Select()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("selected entry: %v", entry)
}
