package coredns

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	json, err := ParseToJsonConfig()
	if err != nil {
		t.Error(err)
		return
	}
	jsonstr := string(json)
	t.Log(jsonstr)
}
