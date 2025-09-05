package main

import "testing"

func TestActionEmbedding(t *testing.T) {
	action := Action{
		Human: Human{
			Health: 100,
		},
	}
	if action.Health != 100 {
		t.Errorf("Health expected to be %d, but got %d", 100, action.Health)
	}
	action.kill()
	if action.Health != 0 {
		t.Errorf("Health expected to be %d, but got %d", 0, action.Health)
	}
}
