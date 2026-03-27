package registry

import (
	"context"
	"testing"
)

func TestCreateRegistry(t *testing.T) {

	CreateSpegelRegistry(context.Background())
}
