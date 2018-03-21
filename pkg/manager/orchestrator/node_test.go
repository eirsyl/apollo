package orchestrator

import (
	"reflect"
	"testing"
)

func TestCMAllSlots(t *testing.T) {
	cm := ClusterNeighbour{
		Slots: []string{"1", "100-105", "[10921->-3099d006bd8b5d986077118b29b82f8a23ce1006]"},
	}
	slots, err := cm.allSlots()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(slots, []int{1, 100, 101, 102, 103, 104, 105}) {
		t.Fatalf("Slots not valid: %v", slots)
	}
}
