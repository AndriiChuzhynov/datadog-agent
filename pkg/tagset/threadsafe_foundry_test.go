// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.Datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tagset

import (
	"testing"
)

func TestThreadsafeFoundry(t *testing.T) {
	testFoundry(t, func() Foundry { return NewThreadsafeFoundry(newCachingFoundry()) })
	testFoundryCaching(t, func() Foundry { return NewThreadsafeFoundry(newCachingFoundry()) })
}
