package tools_test

import (
	"testing"
	. "elyby/minecraft-skinsystem/lib/tools"
)

func TestParseUsername(t *testing.T) {
	if ParseUsername("test.png") != "test" {
		t.Error("Function should trim .png at end")
	}

	if ParseUsername("test") != "test" {
		t.Error("Function should return string itself, if it not contains .png at end")
	}
}

func TestBuildKey(t *testing.T) {
	if BuildKey("Test") != "username:test" {
		t.Error("Function shound convert string to lower case and concatenate it with usernmae:")
	}
}
