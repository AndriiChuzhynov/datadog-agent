// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.Datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tagset

import "sync"

// threadsafeFoundry wraps another foundry and uses a mutex to control
// access.
type threadsafeFoundry struct {
	sync.Mutex
	Foundry
}

// NewThreadsafeFoundry wraps the given foundry with a mutex, ensuring
// thread-safe operation.
func NewThreadsafeFoundry(inner Foundry) Foundry {
	return &threadsafeFoundry{
		Foundry: inner,
	}
}

// NewTags implements Foundry.NewTags
func (f *threadsafeFoundry) NewTags(src ...string) *Tags {
	f.Lock()
	tags := f.Foundry.NewTags(src...)
	f.Unlock()
	return tags
}

// NewUniqueTags implements Foundry.NewUniqueTags
func (f *threadsafeFoundry) NewUniqueTags(src ...string) *Tags {
	f.Lock()
	tags := f.Foundry.NewUniqueTags(src...)
	f.Unlock()
	return tags
}

// NewTagsFromMap implements Foundry.NewTagsFromMap
func (f *threadsafeFoundry) NewTagsFromMap(src map[string]struct{}) *Tags {
	f.Lock()
	tags := f.Foundry.NewTagsFromMap(src)
	f.Unlock()
	return tags
}

// NewTag implements Foundry.NewTag
func (f *threadsafeFoundry) NewTag(tag string) *Tags {
	f.Lock()
	tags := f.Foundry.NewTag(tag)
	f.Unlock()
	return tags
}

// NewBuilder implements Foundry.NewBuilder
func (f *threadsafeFoundry) NewBuilder(capacity int) *Builder {
	f.Lock()
	tags := f.Foundry.NewBuilder(capacity)
	f.Unlock()
	return tags
}

// ParseDSD implements Foundry.ParseDSD
func (f *threadsafeFoundry) ParseDSD(data []byte) (*Tags, error) {
	f.Lock()
	tags, err := f.Foundry.ParseDSD(data)
	f.Unlock()
	return tags, err
}

// Union implements Foundry.Union
func (f *threadsafeFoundry) Union(a, b *Tags) *Tags {
	f.Lock()
	tags := f.Foundry.Union(a, b)
	f.Unlock()
	return tags
}

// DisjointUnion implements Foundry.DisjoingUnion
func (f *threadsafeFoundry) DisjointUnion(a, b *Tags) *Tags {
	f.Lock()
	tags := f.Foundry.DisjointUnion(a, b)
	f.Unlock()
	return tags
}

// getCachedTags implements Foundry.getCachedTags
func (f *threadsafeFoundry) getCachedTags(cacheID cacheID, hash uint64, miss func() *Tags) *Tags {
	f.Lock()
	tags := f.Foundry.getCachedTags(cacheID, hash, miss)
	f.Unlock()
	return tags
}

// builderClosed implements Foundry.builderClosed
func (f *threadsafeFoundry) builderClosed(builder *Builder) {
	f.Lock()
	f.Foundry.builderClosed(builder)
	f.Unlock()
}
