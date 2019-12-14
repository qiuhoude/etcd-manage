package common

import "testing"

func TestGetRootDir(t *testing.T) {
	dir := GetRootDir()
	t.Log(dir)
}
