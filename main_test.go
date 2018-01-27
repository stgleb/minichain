package minichain

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	InitLogger(0)
	result := m.Run()
	os.Exit(result)
}
