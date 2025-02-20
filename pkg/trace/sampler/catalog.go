// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package sampler

import (
	"container/list"
	"sync"

	coreconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// defaultServiceRateKey specifies the key for the default rate to be used by any service that
// doesn't have a rate specified.
const defaultServiceRateKey = "service:,env:"

// maxCatalogEntries specifies the maximum number of entries allowed in the catalog.
const maxCatalogEntries = 5000

// serviceKeyCatalog reverse-maps service signatures to their generated hashes for
// easy look up.
type serviceKeyCatalog struct {
	mu         sync.Mutex
	items      map[ServiceSignature]*list.Element
	ll         *list.List
	maxEntries int
}

type catalogEntry struct {
	key ServiceSignature
	sig Signature
}

// newServiceLookup returns a new serviceKeyCatalog.
func newServiceLookup() *serviceKeyCatalog {
	entries := maxCatalogEntries
	if v := coreconfig.Datadog.GetInt("apm_config.max_catalog_entries"); v > 0 {
		entries = v
	}
	return &serviceKeyCatalog{
		items:      make(map[ServiceSignature]*list.Element),
		ll:         list.New(),
		maxEntries: entries,
	}
}

func (cat *serviceKeyCatalog) register(svcSig ServiceSignature) Signature {
	cat.mu.Lock()
	defer cat.mu.Unlock()
	if el, ok := cat.items[svcSig]; ok {
		// signature already exists, move to front and return already-computed hash
		cat.ll.MoveToFront(el)
		return el.Value.(catalogEntry).sig
	}
	// new signature, compute new hash
	hash := svcSig.Hash()
	el := cat.ll.PushFront(catalogEntry{key: svcSig, sig: hash})
	cat.items[svcSig] = el
	if cat.ll.Len() > cat.maxEntries {
		// list went beyond maximum allowed entries, removed back of the list
		del := cat.ll.Remove(cat.ll.Back()).(catalogEntry)
		delete(cat.items, del.key)
		log.Warnf("More than %d services in service-rates catalog. Dropping %v.", cat.maxEntries, del.key)
	}
	return hash
}

// ratesByService returns a map of service signatures mapping to the rates identified using
// the signatures.
func (cat *serviceKeyCatalog) ratesByService(localRates, remoteRates map[Signature]float64, defaultRate float64) map[ServiceSignature]float64 {
	rbs := make(map[ServiceSignature]float64, len(localRates)+1)
	cat.mu.Lock()
	defer cat.mu.Unlock()
	for key, el := range cat.items {
		sig := el.Value.(catalogEntry).sig
		// todo:raphael distinguish remote rates from local rates with a separate priority value
		if rate, ok := remoteRates[sig]; ok {
			rbs[key] = rate
		} else if rate, ok := localRates[sig]; ok {
			rbs[key] = rate
		} else {
			cat.ll.Remove(el)
			delete(cat.items, key)
		}
	}
	rbs[ServiceSignature{}] = defaultRate
	return rbs
}
