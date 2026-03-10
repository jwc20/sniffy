package main

import "testing"

func TestMain(t *testing.T) {
	t.Log("I am main")

	if 1 == 1 {
		t.Fail()
	}
}
