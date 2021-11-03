// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.Datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// NOTE: this file is named *_test.go because it is only intended for use in
// tests within this package.

package tagset

import (
	"strings"
	"testing"

	"github.com/twmb/murmur3"
)

func TestNullFoundry(t *testing.T) {
	testFoundry(t, func() Foundry { return newNullFoundry() })
}

// A nullFoundry caches nothing.
//
// This type implements Foundry.
type nullFoundry struct {
	baseFoundry
}

func newNullFoundry() *nullFoundry {
	return &nullFoundry{}
}

// NewTags implements Foundry.NewTags
func (f *nullFoundry) NewTags(tags ...string) *Tags {
	tagsMap := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		tagsMap[t] = struct{}{}
	}
	return f.NewTagsFromMap(tagsMap)
}

// NewUniqueTags implements Foundry.NewUniqueTags
func (f *nullFoundry) NewUniqueTags(tags ...string) *Tags {
	hashes, hash := calcHashes(tags)
	return &Tags{tags, hashes, hash}
}

// NewTagsFromMap implements Foundry.NewTagsFromMap
func (f *nullFoundry) NewTagsFromMap(src map[string]struct{}) *Tags {
	tags := make([]string, 0, len(src))
	for tag := range src {
		tags = append(tags, tag)
	}
	hashes, hash := calcHashes(tags)
	return &Tags{tags, hashes, hash}
}

// NewTag implements Foundry.NewTag
func (f *nullFoundry) NewTag(tag string) *Tags {
	hash := murmur3.StringSum64(tag)
	tags := []string{tag}
	hashes := []uint64{hash}
	return &Tags{tags, hashes, hash}
}

// NewBuilder implements Foundry.NewBuilder
func (f *nullFoundry) NewBuilder(capacity int) *Builder {
	return f.baseFoundry.newBuilder(f, capacity)
}

// ParseDSD implements Foundry.ParseDSD
func (f *nullFoundry) ParseDSD(data []byte) (*Tags, error) {
	tags := strings.Split(string(data), ",")
	return f.NewTags(tags...), nil
}

// Union implements Foundry.Union
func (f *nullFoundry) Union(a, b *Tags) *Tags {
	tags := make(map[string]struct{}, len(a.tags)+len(b.tags))
	for _, t := range a.tags {
		tags[t] = struct{}{}
	}
	for _, t := range b.tags {
		tags[t] = struct{}{}
	}
	slice := make([]string, 0, len(tags))
	for tag := range tags {
		slice = append(slice, tag)
	}
	return f.NewTagsFromMap(tags)
}

// DisjointUnion implements Foundry.DisjoingUnion
func (f *nullFoundry) DisjointUnion(a, b *Tags) *Tags {
	tags := make([]string, len(a.tags)+len(b.tags))
	copy(tags[:len(a.tags)], a.tags)
	copy(tags[len(a.tags):], b.tags)

	hashes := make([]uint64, len(a.hashes)+len(b.hashes))
	copy(hashes[:len(a.hashes)], a.hashes)
	copy(hashes[len(a.hashes):], b.hashes)

	hash := a.hash ^ b.hash

	return &Tags{tags, hashes, hash}
}

// getCachedTags implements Foundry.getCachedTags
func (f *nullFoundry) getCachedTags(cacheID cacheID, hash uint64, miss func() *Tags) *Tags {
	return miss()
}
