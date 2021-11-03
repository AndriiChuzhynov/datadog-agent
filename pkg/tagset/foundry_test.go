// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.Datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tagset

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testFoundry(t *testing.T, foundryFactory func() Foundry) {
	t.Run("NewTags", func(t *testing.T) {
		f := foundryFactory()
		tags := f.NewTags("tag1", "tag2", "tag1", "tag2")
		tags.validate(t)
		require.Equal(t, []string{"tag1", "tag2"}, tags.Sorted())
	})

	t.Run("NewTags_NoDupes", func(t *testing.T) {
		f := foundryFactory()
		tags := f.NewTags("tag1", "tag2")
		tags.validate(t)
		require.Equal(t, []string{"tag1", "tag2"}, tags.Sorted())
	})

	t.Run("NewUniqueTags", func(t *testing.T) {
		f := foundryFactory()
		tags := f.NewUniqueTags("tag1", "tag2")
		tags.validate(t)
		require.Equal(t, []string{"tag1", "tag2"}, tags.Sorted())
	})

	t.Run("NewTagsFromMap", func(t *testing.T) {
		f := foundryFactory()
		tags := f.NewTagsFromMap(map[string]struct{}{
			"tag1": {},
			"tag2": {},
		})
		tags.validate(t)
		require.Equal(t, []string{"tag1", "tag2"}, tags.Sorted())
	})

	t.Run("NewTag", func(t *testing.T) {
		f := foundryFactory()
		tags := f.NewTag("foo")
		tags.validate(t)
		require.Equal(t, []string{"foo"}, tags.Sorted())
	})

	t.Run("NewBuilder", func(t *testing.T) {
		f := foundryFactory()
		b := f.NewBuilder(2)
		b.Add("t1")
		b.Add("t2")
		tags := b.Freeze()
		b.Close()
		tags.validate(t)

		require.Equal(t, []string{"t1", "t2"}, tags.Sorted())
	})

	t.Run("ParseDSD", func(t *testing.T) {
		f := foundryFactory()
		tags, err := f.ParseDSD([]byte("tag1,tag2,tag1,tag4"))
		require.NoError(t, err)
		tags.validate(t)

		require.Equal(t, []string{"tag1", "tag2", "tag4"}, tags.Sorted())
	})

	t.Run("Union_Overlapping", func(t *testing.T) {
		f := foundryFactory()
		tags1 := f.NewTags("tag1", "tag2")
		tags2 := f.NewTags("tag2", "tag3")
		tags := f.Union(tags1, tags2)
		tags.validate(t)

		require.Equal(t, []string{"tag1", "tag2", "tag3"}, tags.Sorted())
	})

	t.Run("Union_NonOverlapping", func(t *testing.T) {
		f := foundryFactory()
		tags1 := f.NewTags("tag1", "tag2")
		tags2 := f.NewTags("tag5", "tag6")
		tags := f.Union(tags1, tags2)
		tags.validate(t)

		require.Equal(t, []string{"tag1", "tag2", "tag5", "tag6"}, tags.Sorted())
	})

	t.Run("DisjointUnion", func(t *testing.T) {
		f := foundryFactory()
		tags1 := f.NewTags("tag1", "tag2")
		tags2 := f.NewTags("tag5", "tag6")
		tags := f.DisjointUnion(tags1, tags2)
		tags.validate(t)

		require.Equal(t, []string{"tag1", "tag2", "tag5", "tag6"}, tags.Sorted())
	})
}

func testFoundryCaching(t *testing.T, foundryFactory func() Foundry) {
	t.Run("Caching", func(t *testing.T) {
		t.Run("NewTags", func(t *testing.T) {
			f := foundryFactory()
			tags1 := f.NewTags("tag1", "tag2")
			tags2 := f.NewTags("tag2", "tag1")

			// check for pointer equality
			require.True(t, tags1 == tags2)
		})

		t.Run("NewTagsFromMap", func(t *testing.T) {
			f := foundryFactory()
			tags1 := f.NewTagsFromMap(map[string]struct{}{
				"tag1": {},
				"tag2": {},
			})
			tags2 := f.NewTagsFromMap(map[string]struct{}{
				"tag2": {},
				"tag1": {},
			})

			// check for pointer equality
			require.True(t, tags1 == tags2)
		})

		t.Run("NewTag", func(t *testing.T) {
			f := foundryFactory()
			tag1a := f.NewTag("tag1")
			tag1b := f.NewTag("tag1")
			tag2 := f.NewTag("tag2")

			// check for pointer equality
			require.True(t, tag1a == tag1b)
			require.True(t, tag1a != tag2)
		})

		t.Run("NewBuilder", func(t *testing.T) {
			f := foundryFactory()
			tags1 := f.NewTags("t1", "t2")
			tags1.validate(t)

			b := f.NewBuilder(2)
			b.Add("t1")
			b.Add("t2")
			tags2 := b.Freeze()
			b.Close()
			tags2.validate(t)

			// check for pointer equality
			require.True(t, tags1 == tags2)
		})

		t.Run("ParseDSD", func(t *testing.T) {
			f := foundryFactory()

			tags1, err := f.ParseDSD([]byte("tag1,tag2,tag4"))
			require.NoError(t, err)
			tags1.validate(t)

			tags2, err := f.ParseDSD([]byte("tag1,tag2,tag4"))
			require.NoError(t, err)
			tags2.validate(t)

			// check for pointer equality (ideally because the input
			// data hash matched, but we can't actually tell)
			require.True(t, tags1 == tags2)
		})

		t.Run("ParseDSD_DifferentInput", func(t *testing.T) {
			f := foundryFactory()

			tags1, err := f.ParseDSD([]byte("tag1,tag2,tag4"))
			require.NoError(t, err)
			tags1.validate(t)

			tags2, err := f.ParseDSD([]byte("tag4,tag1,tag2"))
			require.NoError(t, err)
			tags2.validate(t)

			// check for pointer equality (these are still equal
			// because the NewTags operation de-duplicates them)
			require.True(t, tags1 == tags2)
		})

		t.Run("Union", func(t *testing.T) {
			f := foundryFactory()
			tags12 := f.NewTags("tag1", "tag2")
			tags23 := f.NewTags("tag2", "tag3")
			tags13 := f.NewTags("tag1", "tag3")
			allTagsA := f.Union(tags12, tags23)
			allTagsA.validate(t)
			allTagsB := f.Union(tags13, tags23)
			allTagsB.validate(t)

			// check for pointer equality
			require.True(t, allTagsA == allTagsB)
		})

		t.Run("DisjointUnion", func(t *testing.T) {
			f := foundryFactory()
			tags12 := f.NewTags("tag1", "tag2")
			tags34 := f.NewTags("tag3", "tag4")
			allTagsA := f.DisjointUnion(tags12, tags34)
			allTagsA.validate(t)
			allTagsB := f.DisjointUnion(tags34, tags12)
			allTagsB.validate(t)

			// check for pointer equality
			require.True(t, allTagsA == allTagsB)
		})
	})
}
