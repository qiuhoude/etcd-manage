package config

import "testing"

func TestCheckEtcdServerName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"~xx", false},
		{"", false},
		{"_", true},
	}
	for _, v := range tests {
		if checkEtcdServerName(v.name) != v.want {
			t.Fatal("checkEtcdServerName() =>")
		}
	}
}
