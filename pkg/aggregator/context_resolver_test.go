// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build test

package aggregator

import (
	// stdlib

	"fmt"
	"testing"

	// 3p
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/datadog-agent/pkg/aggregator/ckey"
	"github.com/DataDog/datadog-agent/pkg/metrics"
)

func TestGenerateContextKey(t *testing.T) {
	mSample := metrics.MetricSample{
		Name:       "my.metric.name",
		Value:      1,
		Mtype:      metrics.GaugeType,
		Tags:       []string{"foo", "bar"},
		Host:       "metric-hostname",
		SampleRate: 1,
	}

	contextKey := generateContextKey(&mSample)
	assert.Equal(t, ckey.ContextKey(0xd28d2867c6dd822c), contextKey)
}

func TestTrackContext(t *testing.T) {
	mSample1 := metrics.MetricSample{
		Name:       "my.metric.name",
		Value:      1,
		Mtype:      metrics.GaugeType,
		Tags:       []string{"foo", "bar"},
		SampleRate: 1,
	}
	mSample2 := metrics.MetricSample{
		Name:       "my.metric.name",
		Value:      1,
		Mtype:      metrics.GaugeType,
		Tags:       []string{"foo", "bar", "baz"},
		SampleRate: 1,
	}
	mSample3 := metrics.MetricSample{ // same as mSample2, with different Host
		Name:       "my.metric.name",
		Value:      1,
		Mtype:      metrics.GaugeType,
		Tags:       []string{"foo", "bar", "baz"},
		Host:       "metric-hostname",
		SampleRate: 1,
	}
	expectedContext1 := Context{
		Name: mSample1.Name,
		Tags: mSample1.Tags,
	}
	expectedContext2 := Context{
		Name: mSample2.Name,
		Tags: mSample2.Tags,
	}
	expectedContext3 := Context{
		Name: mSample3.Name,
		Tags: mSample3.Tags,
		Host: mSample3.Host,
	}
	contextResolver := newContextResolver()

	// Track the 2 contexts
	contextKey1 := contextResolver.trackContext(&mSample1)
	contextKey2 := contextResolver.trackContext(&mSample2)
	contextKey3 := contextResolver.trackContext(&mSample3)

	// When we look up the 2 keys, they return the correct contexts
	context1 := contextResolver.contextsByKey[contextKey1]
	assert.Equal(t, expectedContext1, *context1)

	context2 := contextResolver.contextsByKey[contextKey2]
	assert.Equal(t, expectedContext2, *context2)

	context3 := contextResolver.contextsByKey[contextKey3]
	assert.Equal(t, expectedContext3, *context3)

	unknownContextKey := ckey.ContextKey(0xffffffffffffffff)
	_, ok := contextResolver.contextsByKey[unknownContextKey]
	assert.False(t, ok)
}

func TestExpireContexts(t *testing.T) {
	mSample1 := metrics.MetricSample{
		Name:       "my.metric.name",
		Value:      1,
		Mtype:      metrics.GaugeType,
		Tags:       []string{"foo", "bar"},
		SampleRate: 1,
	}
	mSample2 := metrics.MetricSample{
		Name:       "my.metric.name",
		Value:      1,
		Mtype:      metrics.GaugeType,
		Tags:       []string{"foo", "bar", "baz"},
		SampleRate: 1,
	}
	contextResolver := newTimestampContextResolver()

	// Track the 2 contexts
	contextKey1 := contextResolver.trackContext(&mSample1, 4)
	contextKey2 := contextResolver.trackContext(&mSample2, 6)

	// With an expireTimestap of 3, both contexts are still valid
	assert.Len(t, contextResolver.expireContexts(3), 0)
	_, ok1 := contextResolver.resolver.contextsByKey[contextKey1]
	_, ok2 := contextResolver.resolver.contextsByKey[contextKey2]
	assert.True(t, ok1)
	assert.True(t, ok2)

	// With an expireTimestap of 5, context 1 is expired
	expiredContextKeys := contextResolver.expireContexts(5)
	if assert.Len(t, expiredContextKeys, 1) {
		assert.Equal(t, contextKey1, expiredContextKeys[0])
	}

	// context 1 is not tracked anymore, but context 2 still is
	_, ok := contextResolver.resolver.contextsByKey[contextKey1]
	assert.False(t, ok)
	_, ok = contextResolver.resolver.contextsByKey[contextKey2]
	assert.True(t, ok)
}

func TestCountBasedExpireContexts(t *testing.T) {
	mSample1 := metrics.MetricSample{Name: "my.metric.name1"}
	mSample2 := metrics.MetricSample{Name: "my.metric.name2"}
	mSample3 := metrics.MetricSample{Name: "my.metric.name3"}
	contextResolver := newCountBasedContextResolver(2)

	contextKey1 := contextResolver.trackContext(&mSample1)
	contextKey2 := contextResolver.trackContext(&mSample2)
	require.Len(t, contextResolver.expireContexts(), 0)

	contextKey3 := contextResolver.trackContext(&mSample3)
	contextResolver.trackContext(&mSample2)
	require.Len(t, contextResolver.expireContexts(), 0)

	expiredContextKeys := contextResolver.expireContexts()
	require.ElementsMatch(t, expiredContextKeys, []ckey.ContextKey{contextKey1})

	expiredContextKeys = contextResolver.expireContexts()
	require.ElementsMatch(t, expiredContextKeys, []ckey.ContextKey{contextKey2, contextKey3})

	require.Len(t, contextResolver.expireContexts(), 0)
	require.Len(t, contextResolver.resolver.contextsByKey, 0)
}

func TestTagDeduplication(t *testing.T) {
	resolver := newContextResolver()

	ckey := resolver.trackContext(&metrics.MetricSample{
		Name: "foo",
		Tags: []string{"bar", "bar"},
	})

	assert.Equal(t, len(resolver.contextsByKey[ckey].Tags), 1)
	assert.Equal(t, resolver.contextsByKey[ckey].Tags, []string{"bar"})
}

// TODO(remy): dedup this method which has been stolen in ckey pkg
func genTags(count int, div int) ([]string, []string) {
	var tags []string
	uniqMap := make(map[string]struct{})
	for i := 0; i < count; i++ {
		tag := fmt.Sprintf("tag%d:value%d", i/div, i/div)
		tags = append(tags, tag)
		uniqMap[tag] = struct{}{}
	}

	uniq := []string{}
	for tag := range uniqMap {
		uniq = append(uniq, tag)
	}

	return tags, uniq
}

func BenchmarkContextResolverTrackContext(b *testing.B) {
	resetAggregator()
	agg := NewBufferedAggregator(nil, nil, "hostname", 0)
	SetDefaultAggregator(agg)

	// track 1M contexts with 30 tags
	for contextsCount := 1; contextsCount < 2<<20; contextsCount *= 2 {
		tags, _ := genTags(30, 1)
		resolver := newContextResolver()
		b.Run(fmt.Sprintf("with-%d-contexts", contextsCount), func(b *testing.B) {
			b.ReportAllocs()
			j := 0
			for n := 0; n < b.N; n++ {
				resolver.trackContext(&metrics.MetricSample{
					Name: fmt.Sprintf("metric.name%d", j),
					Tags: tags,
				})
				j++
				if j >= contextsCount {
					j = 0
				}
			}
		})
	}
}

func BenchmarkContextResolverGetWorstCase(b *testing.B) {
	resetAggregator()
	agg := NewBufferedAggregator(nil, nil, "hostname", 0)
	SetDefaultAggregator(agg)

	// track 1M contexts with 30 tags
	for contextsCount := 1; contextsCount < 2<<20; contextsCount *= 2 {
		tags, _ := genTags(30, 1)
		resolver := newContextResolver()
		ckeys := make([]ckey.ContextKey, 0)
		for i := 0; i < contextsCount; i++ {
			ckeys = append(ckeys, resolver.trackContext(&metrics.MetricSample{
				Name: fmt.Sprintf("metric.name%d", i),
				Tags: tags,
			}))
		}

		b.Run(fmt.Sprintf("with-%d-contexts", contextsCount), func(b *testing.B) {
			b.ReportAllocs()
			j := 0
			for n := 0; n < b.N; n++ {
				resolver.get(ckeys[j])
				j++
				if j >= contextsCount {
					j = 0
				}
			}
		})
	}
}
