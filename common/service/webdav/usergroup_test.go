// nolint
package webdav

import (
	"testing"
)

func TestInit(t *testing.T) {
	ug := NewUserGroup("1")
	g, err := ug.GetGroupName(13)
	if err != nil {
		t.Error(err)
	}
	u, err := ug.GetUserName(13)
	if err != nil {
		t.Error(err)
	}
	t.Log(g, u)
}
