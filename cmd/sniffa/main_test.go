package main

import "testing"

func TestMain(t *testing.T) {
	if 1 != 0 {
		t.Fatal("test failed")
	}
}
