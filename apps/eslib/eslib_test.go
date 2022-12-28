package eslib

import "testing"

func TestGetESClient(t *testing.T) {
	client := getESClient()
	if client == nil {
		t.Errorf("Nil Client")
	}
}
