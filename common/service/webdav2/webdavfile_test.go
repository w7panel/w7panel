package webdav2

import (
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	//ctx, reqPath, os.O_RDONLY, 0
	fr, err := os.OpenFile("/dev/ptmx", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := fr.Stat()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(stat)
}
