// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build linux

package probe

import (
	"crypto/md5"
	"hash"
	"io"

	"github.com/DataDog/datadog-agent/pkg/security/log"
	manager "github.com/DataDog/ebpf-manager"
)

const errorSentinel uint64 = ^uint64(0)

// ConstantFetcher represents a source of constants that can be used to fill up
// eBPF relocations
type ConstantFetcher interface {
	AppendSizeofRequest(id, typeName, headerName string)
	AppendOffsetofRequest(id, typeName, fieldName, headerName string)
	FinishAndGetResults() (map[string]uint64, error)
}

// ComposeConstantFetcher represents a composition of child constant fetchers
// It allows the usage of fallbacks if the main source is not working
type ComposeConstantFetcher struct {
	hasher   hash.Hash
	fetchers []ConstantFetcher
	requests []*composeRequest
}

// ComposeConstantFetchers creates a new ComposeConstantFetcher based on the
// passed fetchers
func ComposeConstantFetchers(fetchers []ConstantFetcher) *ComposeConstantFetcher {
	return &ComposeConstantFetcher{
		hasher:   md5.New(),
		fetchers: fetchers,
	}
}

func (f *ComposeConstantFetcher) appendRequest(req *composeRequest) {
	f.requests = append(f.requests, req)
	io.WriteString(f.hasher, req.id)
	io.WriteString(f.hasher, req.typeName)
	io.WriteString(f.hasher, req.fieldName)
	io.WriteString(f.hasher, req.headerName)
}

// AppendSizeofRequest appends a sizeof request
func (f *ComposeConstantFetcher) AppendSizeofRequest(id, typeName, headerName string) {
	f.appendRequest(&composeRequest{
		id:         id,
		sizeof:     true,
		typeName:   typeName,
		fieldName:  "",
		headerName: headerName,
		value:      errorSentinel,
	})
}

// AppendOffsetofRequest appends an offset request
func (f *ComposeConstantFetcher) AppendOffsetofRequest(id, typeName, fieldName, headerName string) {
	f.appendRequest(&composeRequest{
		id:         id,
		sizeof:     false,
		typeName:   typeName,
		fieldName:  fieldName,
		headerName: headerName,
		value:      errorSentinel,
	})
}

func (f *ComposeConstantFetcher) getHash() []byte {
	return f.hasher.Sum(nil)
}

// FinishAndGetResults does the actual fetching and returns the results
func (f *ComposeConstantFetcher) FinishAndGetResults() (map[string]uint64, error) {
	currentHash := f.getHash()
	if constantsCache.isMatching(currentHash) {
		log.Warnf("cached hit")
		return constantsCache.constants, nil
	}

	for _, fetcher := range f.fetchers {
		for _, req := range f.requests {
			if req.value == errorSentinel {
				if req.sizeof {
					fetcher.AppendSizeofRequest(req.id, req.typeName, req.headerName)
				} else {
					fetcher.AppendOffsetofRequest(req.id, req.typeName, req.fieldName, req.headerName)
				}
			}
		}

		res, err := fetcher.FinishAndGetResults()
		if err != nil {
			log.Errorf("failed to run constant fetcher: %v", err)
		}

		for _, req := range f.requests {
			if req.value == errorSentinel {
				if newValue, present := res[req.id]; present {
					req.value = newValue
				}
			}
		}
	}

	finalRes := make(map[string]uint64)
	for _, req := range f.requests {
		finalRes[req.id] = req.value
	}

	constantsCache = &cachedConstants{
		constants: finalRes,
		hash:      currentHash,
	}
	return finalRes, nil
}

type composeRequest struct {
	id                  string
	sizeof              bool
	typeName, fieldName string
	headerName          string
	value               uint64
}

// FallbackConstantFetcher is a constant fetcher that uses the old fallback
// heuristics to fetch constants
type FallbackConstantFetcher struct {
	probe *Probe
	res   map[string]uint64
}

// NewFallbackConstantFetcher returns a new FallbackConstantFetcher
func NewFallbackConstantFetcher(probe *Probe) *FallbackConstantFetcher {
	return &FallbackConstantFetcher{
		probe: probe,
		res:   make(map[string]uint64),
	}
}

func (f *FallbackConstantFetcher) appendRequest(id string) {
	var value = errorSentinel
	switch id {
	case "sizeof_inode":
		value = getSizeOfStructInode(f.probe)
	case "sb_magic_offset":
		value = getSuperBlockMagicOffset(f.probe)
	case "tty_offset":
		value = getSignalTTYOffset(f.probe)
	case "tty_name_offset":
		value = getTTYNameOffset(f.probe)
	}
	f.res[id] = value
}

// AppendSizeofRequest appends a sizeof request
func (f *FallbackConstantFetcher) AppendSizeofRequest(id, typeName, headerName string) {
	f.appendRequest(id)
}

// AppendOffsetofRequest appends an offset request
func (f *FallbackConstantFetcher) AppendOffsetofRequest(id, typeName, fieldName, headerName string) {
	f.appendRequest(id)
}

// FinishAndGetResults returns the results
func (f *FallbackConstantFetcher) FinishAndGetResults() (map[string]uint64, error) {
	return f.res, nil
}

func createConstantEditors(probe *Probe, constants map[string]uint64) []manager.ConstantEditor {
	res := make([]manager.ConstantEditor, 0, len(constants))
	log.Warnf("kernel version: %v", probe.kernelVersion)
	for name, value := range constants {
		if value == errorSentinel {
			log.Warnf("failed to fetch constant for %s", name)
			value = 0
		}

		log.Warnf("constant %s = %v", name, value)

		res = append(res, manager.ConstantEditor{
			Name:  name,
			Value: value,
		})
	}
	return res
}

var constantsCache *cachedConstants

type cachedConstants struct {
	constants map[string]uint64
	hash      []byte
}

func (cc *cachedConstants) isMatching(hash []byte) bool {
	if cc == nil {
		return false
	}

	if len(hash) != len(cc.hash) {
		return false
	}

	for i, v := range cc.hash {
		if v != hash[i] {
			return false
		}
	}
	return true
}
