package tagset

import (
	"testing"
)

// These tests just provide coverage of the default stubs.  Other tests
// perform more thorough validation of the functionality.

func TestNewTags(t *testing.T) {
	tags := NewTags("a", "b", "a")
	tags.validate(t)
}

func TestNewUniqueTags(t *testing.T) {
	tags := NewUniqueTags("a", "b")
	tags.validate(t)
}

func TestNewTagsFromMap(t *testing.T) {
	tags := NewTagsFromMap(map[string]struct{}{"a": {}, "b": {}})
	tags.validate(t)
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder(10)
	b.Add("a")
	b.Add("b")
	tags := b.Freeze()
	b.Close()
	tags.validate(t)
}

func TestUnion(t *testing.T) {
	tags := Union(
		NewTags("a", "b", "c"),
		NewTags("c", "d", "e"),
	)
	tags.validate(t)
}

func TestDisjointUnion(t *testing.T) {
	tags := DisjointUnion(
		NewTags("a", "b", "c"),
		NewTags("d", "e"),
	)
	tags.validate(t)
}
