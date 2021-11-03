package tagset

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testBuilder() *Builder {
	bldr := newBuilder(newNullFoundry())
	bldr.reset()
	return bldr
}

func TestBuilder_Add_Freeze_Close(t *testing.T) {
	bldr := testBuilder()

	bldr.Add("abc")
	bldr.Add("def")
	bldr.Add("abc")

	tags := bldr.Freeze()
	tags2 := bldr.Freeze()
	bldr.Close()

	require.Equal(t, tags, tags2)
	require.Equal(t, tags.Sorted(), []string{"abc", "def"})
}

func TestBuilder_AddKV_Freeze_Close(t *testing.T) {
	bldr := testBuilder()

	bldr.AddKV("host", "foo")
	bldr.AddKV("host", "bar")
	bldr.AddKV("host", "foo")

	tags := bldr.Freeze()
	bldr.Close()

	require.Equal(t, tags.Sorted(), []string{"host:bar", "host:foo"})
}

func TestBuilder_Contains(t *testing.T) {
	bldr := testBuilder()

	bldr.AddKV("host", "foo")
	bldr.AddKV("host", "bar")
	bldr.AddKV("host", "foo")

	require.True(t, bldr.Contains("host:foo"))
	require.False(t, bldr.Contains("host:bing"))
}
