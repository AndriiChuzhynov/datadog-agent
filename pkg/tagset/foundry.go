// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.Datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tagset

// The Foundry type is responsible for creating new Tags instances. Its
// interface is simple, but provides an opportunity for optimization and
// deduplication.
//
// A single default foundry is provided for use throughout the agent, with
// package-level functions deferring to that foundry. Additional, specific
// foundries may be created for specific purposes. The default foundry is
// thread-safe, but it may be advantageous to build non-thread-safe foundries
// for specific circumstances. Tags instances created by different foundries
// can be used interchangeably and are entirely thread-safe.
//
// Tags instances returned from different foundries may be used
// interchangeably.  The only disadvantage of using multiple foundries is a
// reduced cache rate due to not sharing caches between those foundries.
type Foundry interface {
	// Tags constructors

	// NewTags creates a new *Tags with the given tags.
	NewTags(src ...string) *Tags

	// NewUniqueTags creates a new *Tags.  This method assumes the tags in the
	// given slice are unique.
	NewUniqueTags(src ...string) *Tags

	// NewTagsFromMap creates a new *Tags based on the keys of the given map.
	NewTagsFromMap(src map[string]struct{}) *Tags

	// NewTag creates a new *Tags with a single tag in it
	NewTag(tag string) *Tags

	// Builder constructors

	// NewBuilder returns a fresh Builder
	NewBuilder(capacity int) *Builder

	// NewSliceBuilder returns a fresh Builder
	// TODO (does this need to be able to re-level a tag?)
	// NewSliceBuilder(levels, capacity int) HashingTagsBuilder

	// Parsing

	// UnmarshalJSON unmarshals a JSON array of strings as tags
	// TODO: lint doesn't like this signature
	// UnmarshalJSON(data []byte) (*Tags, error)

	// UnmarshalYAML unmarshals a YAML array of strings as tags
	// TODO
	// UnmarshalYAML(data []byte) (*Tags, error)

	// ParseDSD parses a comma-separated set of tags, as used in the DogStatsD
	// format.
	ParseDSD(data []byte) (*Tags, error)

	// Combination

	// Union combines two *Tags instances that are not known to be
	// disjoint.  That is, there may exist tags that are in both tagsets.
	Union(a, b *Tags) *Tags

	// DisjointUnion combines two *Tags instances when the caller is
	// certain that no tag is in both tagsets.
	DisjointUnion(a, b *Tags) *Tags

	// getCachedTags returns a Tags instance with the given cache key
	// in the given cache.  If the cache element does not exist, then the
	// miss function is called to generate it.
	getCachedTags(cacheID cacheID, hash uint64, miss func() *Tags) *Tags

	// builderClosed returns the given builder to the foundry for reuse.
	// Builder.Close() calls this.
	builderClosed(builder *Builder)
}

// cacheId is an identifier for a cache keying Tags instances by a particular key.
type cacheID = uint

const (
	// byTasgsetHashCache indexes a cache by the Tags instances' hash
	byTagsetHashCache cacheID = iota
	// byDSDHashCache indexes a cache by the murmur3 hash of the DSD data
	byDSDHashCache cacheID = iota
	numCacheIDs
)
