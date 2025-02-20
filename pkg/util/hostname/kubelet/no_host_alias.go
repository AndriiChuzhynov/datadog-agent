// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build !kubelet

package hostinfo

import (
	"context"
	"fmt"
)

// GetHostAlias uses the "kubelet" hostname provider to fetch the kubernetes alias
func GetHostAlias(ctx context.Context) (string, error) {
	return "", fmt.Errorf("Kubernetes support not build: couldn't extract a host alias from the kubelet")
}

// GetMetaClusterNameText returns the clusterName text for the agent status output
func GetMetaClusterNameText(ctx context.Context, hostname string) string {
	return ""
}
