package clip

import (
	"testing"
)

func TestUrlChunksPrefix(t *testing.T) {
	if urlPart("/abc/123/%%%", 2) != "123" {
		t.Fatalf("Expected %s, was %s", "123", urlPart("/abc/123/%%%", 2))
	}
}

func TestUrlChunksNoPrefix(t *testing.T) {
	if urlPart("abc/123/%%%", 2) != "%%%" {
		t.Fatalf("Expected %s, was %s", "%%%", urlPart("/abc/123/%%%", 2))
	}
}

func TestUrlChunksTrailing(t *testing.T) {
	if urlPart("/abc/123/%%%/", 3) != "%%%" {
		t.Fatalf("Expected %s, was %s", "%%%", urlPart("/abc/123/%%%", 3))
	}
}

func TestUrlChunksOutOfBounds(t *testing.T) {
	if len(urlPart("/foo/bar/rest", 5)) > 0 {
		t.Fatalf(`Expected "%s", was "%s"`, "", urlPart("/foo/bar/rest", 5))
	}
}

func TestUrlChunksFirstIndex(t *testing.T) {
	if urlPart("foo/bar", 0) != "foo" {
		t.Fatalf(`Expected "%s", was "%s"`, "foo", urlPart("foo/bar", 0))
	}
}

func TestUrlChunksFirstAndOnlyIndex(t *testing.T) {
	if urlPart("foo", 0) != "foo" {
		t.Fatalf(`Expected "%s", was "%s"`, "foo", urlPart("foo/bar", 0))
	}
}
