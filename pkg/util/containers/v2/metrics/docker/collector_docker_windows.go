// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021-present Datadog, Inc.

//go:build docker && windows
// +build docker,windows

package docker

import (
	"time"

	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/containers/v2/metrics"
	"github.com/docker/docker/api/types"
)

func convertContainerStats(stats *types.Stats) *metrics.ContainerStats {
	return &metrics.ContainerStats{
		Timestamp: time.Now(),
		CPU:       convertCPUStats(&stats.CPUStats),
		Memory:    convertMemoryStats(&stats.MemoryStats),
		IO:        convertIOStats(&stats.StorageStats),
		PID:       convertPIDStats(stats.NumProcs),
	}
}

func convertCPUStats(cpuStats *types.CPUStats) *metrics.ContainerCPUStats {
	return &metrics.ContainerCPUStats{
		// ContainerCPUStats expects CPU metrics in nanoseconds
		// *On Windows* (only) CPUStats units are 100’s of nanoseconds
		Total:  util.UIntToFloatPtr(100 * cpuStats.CPUUsage.TotalUsage),
		System: util.UIntToFloatPtr(100 * cpuStats.CPUUsage.UsageInKernelmode),
		User:   util.UIntToFloatPtr(100 * cpuStats.CPUUsage.UsageInUsermode),
	}
}

func convertMemoryStats(memStats *types.MemoryStats) *metrics.ContainerMemStats {
	return &metrics.ContainerMemStats{
		UsageTotal:        util.UIntToFloatPtr(memStats.Usage),
		Limit:             util.UIntToFloatPtr(memStats.Limit),
		PrivateWorkingSet: util.UIntToFloatPtr(memStats.PrivateWorkingSet),
		CommitBytes:       util.UIntToFloatPtr(memStats.Commit),
		CommitPeakBytes:   util.UIntToFloatPtr(memStats.CommitPeak),
	}
}

func convertIOStats(storageStats *types.StorageStats) *metrics.ContainerIOStats {
	return &metrics.ContainerIOStats{
		ReadBytes:       util.UIntToFloatPtr(storageStats.ReadSizeBytes),
		WriteBytes:      util.UIntToFloatPtr(storageStats.WriteSizeBytes),
		ReadOperations:  util.UIntToFloatPtr(storageStats.ReadCountNormalized),
		WriteOperations: util.UIntToFloatPtr(storageStats.WriteCountNormalized),
	}
}

func convertPIDStats(numProcs uint32) *metrics.ContainerPIDStats {
	return &metrics.ContainerPIDStats{
		ThreadCount: util.UIntToFloatPtr(numProcs),
	}
}
