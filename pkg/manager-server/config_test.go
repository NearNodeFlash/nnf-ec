package server

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	config, err := loadConfig()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v", config)
}
