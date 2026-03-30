package registry

import (
	"context"
	"os"
	"testing"
)

func TestNerdCommit(t *testing.T) {
	os.Setenv("DEBUG", "true")

	dig, err := CommitToContainerD(context.Background(), "ccr.ccs.tencentyun.com/afan-public/nginx:test", "nginx-test")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(dig)
}
