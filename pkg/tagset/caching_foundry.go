// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.Datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tagset

import (
	"strings"

	"github.com/twmb/murmur3"
)

// A cachingFoundry caches tagsets with no eviction policy.
//
// This type implements Foundry.
type cachingFoundry struct {
	baseFoundry
	caches [numCacheIDs]map[uint64]*Tags
}

func newCachingFoundry() *cachingFoundry {
	var caches [numCacheIDs]map[uint64]*Tags
	for i := range caches {
		caches[i] = make(map[uint64]*Tags)
	}
	return &cachingFoundry{
		caches: caches,
	}
}

// NewTags implements Foundry.NewTags
func (f *cachingFoundry) NewTags(tags ...string) *Tags {
	tagsMap := make(map[uint64]string, len(tags))
	hash := uint64(0)
	for _, t := range tags {
		h := murmur3.StringSum64(t)
		_, seen := tagsMap[h]
		if seen {
			continue
		}
		tagsMap[h] = t
		hash ^= h
	}

	return f.getCachedTags(byTagsetHashCache, hash, func() *Tags {
		// write hashes and rewrite tags based on the map
		hashes := make([]uint64, len(tagsMap))
		tags = tags[:len(tagsMap)]
		i := 0
		for h, t := range tagsMap {
			tags[i] = t
			hashes[i] = h
			i++
		}

		return &Tags{tags, hashes, hash}
	})
}

// NewUniqueTags implements Foundry.NewUniqueTags
func (f *cachingFoundry) NewUniqueTags(tags ...string) *Tags {
	hashes, hash := calcHashes(tags)
	return f.getCachedTags(byTagsetHashCache, hash, func() *Tags {
		return &Tags{tags, hashes, hash}
	})
}

// NewTagsFromMap implements Foundry.NewTagsFromMap
func (f *cachingFoundry) NewTagsFromMap(src map[string]struct{}) *Tags {
	tags := make([]string, 0, len(src))
	for tag := range src {
		tags = append(tags, tag)
	}
	hashes, hash := calcHashes(tags)
	return f.getCachedTags(byTagsetHashCache, hash, func() *Tags {
		return &Tags{tags, hashes, hash}
	})
}

// NewTag implements Foundry.NewTag
func (f *cachingFoundry) NewTag(tag string) *Tags {
	hash := murmur3.StringSum64(tag)
	return f.getCachedTags(byTagsetHashCache, hash, func() *Tags {
		return &Tags{[]string{tag}, []uint64{hash}, hash}
	})
}

// NewBuilder implements Foundry.NewBuilder
func (f *cachingFoundry) NewBuilder(capacity int) *Builder {
	return f.baseFoundry.newBuilder(f, capacity)
}

// ParseDSD implements Foundry.ParseDSD
func (f *cachingFoundry) ParseDSD(data []byte) (*Tags, error) {
	// TODO: GO FASTER
	return f.getCachedTags(byDSDHashCache, murmur3.Sum64(data), func() *Tags {
		tags := strings.Split(string(data), ",")
		return f.NewTags(tags...)
	}), nil
}

// Union implements Foundry.Union
func (f *cachingFoundry) Union(a, b *Tags) *Tags {
	tags := make(map[string]struct{}, len(a.tags)+len(b.tags))
	for _, t := range a.tags {
		tags[t] = struct{}{}
	}
	for _, t := range b.tags {
		tags[t] = struct{}{}
	}
	return f.NewTagsFromMap(tags)
}

// DisjointUnion implements Foundry.DisjoingUnion
func (f *cachingFoundry) DisjointUnion(a, b *Tags) *Tags {
	hash := a.hash ^ b.hash
	return f.getCachedTags(byTagsetHashCache, hash, func() *Tags {

		tags := make([]string, len(a.tags)+len(b.tags))
		copy(tags[:len(a.tags)], a.tags)
		copy(tags[len(a.tags):], b.tags)

		hashes := make([]uint64, len(a.hashes)+len(b.hashes))
		copy(hashes[:len(a.hashes)], a.hashes)
		copy(hashes[len(a.hashes):], b.hashes)
		return &Tags{tags, hashes, hash}
	})
}

// getCachedTags implements Foundry.getCachedTags
func (f *cachingFoundry) getCachedTags(cacheID cacheID, hash uint64, miss func() *Tags) *Tags {
	cache := f.caches[cacheID]
	v, ok := cache[hash]
	if !ok {
		v = miss()
		cache[hash] = v
	}
	return v
}
