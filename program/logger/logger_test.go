package logger

import (
	"testing"
)

func TestInitLogger(t *testing.T) {
	log, err := InitLogger("log", false)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("hihi")
}
