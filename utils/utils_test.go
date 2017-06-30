package utils

import (
	"testing"
	"time"
)

func TestParseUsername(t *testing.T) {
	if ParseUsername("test.png") != "test" {
		t.Error("Function should trim .png at end")
	}

	if ParseUsername("test") != "test" {
		t.Error("Function should return string itself, if it not contains .png at end")
	}
}

func TestBuildNonElyTexturesHash(t *testing.T) {
	timeNow = func() time.Time {
		return time.Date(2017, time.November, 30, 16, 15, 34, 0, time.UTC)
	}

	if BuildNonElyTexturesHash("username") != "686d788a5353cb636e8fdff727634d88" {
		t.Error("Function should return fixed hash by username-time pair")
	}

	if BuildNonElyTexturesHash("another-username") != "fb876f761683a10accdb17d403cef64c" {
		t.Error("Function should return fixed hash by username-time pair")
	}

	timeNow = func() time.Time {
		return time.Date(2017, time.November, 30, 16, 20, 12, 0, time.UTC)
	}

	if BuildNonElyTexturesHash("username") != "686d788a5353cb636e8fdff727634d88" {
		t.Error("Function should do not change it's value if hour the same")
	}

	if BuildNonElyTexturesHash("another-username") != "fb876f761683a10accdb17d403cef64c" {
		t.Error("Function should return fixed hash by username-time pair")
	}

	timeNow = func() time.Time {
		return time.Date(2017, time.November, 30, 17, 1, 3, 0, time.UTC)
	}

	if BuildNonElyTexturesHash("username") != "42277892fd24bc0ed86285b3bb8b8fad" {
		t.Error("Function should change it's value if hour changed")
	}
}

func TestBuildElyUrl(t *testing.T) {
	if BuildElyUrl("/route") != "http://ely.by/route" {
		t.Error("Function should add prefix to the provided relative url.")
	}

	if BuildElyUrl("http://ely.by/test/route") != "http://ely.by/test/route" {
		t.Error("Function should do not add prefix to the provided prefixed url.")
	}
}
