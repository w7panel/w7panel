package cgroups

import "testing"

func TestCurrentStat(t *testing.T) {
	stat, err := CurrentStat()
	if err != nil {
		t.Log(err)
	}
	t.Log(stat)
}
