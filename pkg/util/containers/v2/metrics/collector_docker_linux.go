// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021-present Datadog, Inc.

//go:build docker && linux
// +build docker,linux

package metrics

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/docker/docker/api/types"
)

func convertContainerStats(stats *types.Stats) *ContainerStats {
	return &ContainerStats{
		Timestamp: time.Now(),
		CPU:       convertCPUStats(&stats.CPUStats),
		Memory:    convertMemoryStats(&stats.MemoryStats),
		IO:        convertIOStats(&stats.BlkioStats),
		PID:       convertPIDStats(&stats.PidsStats),
	}
}

func convertCPUStats(cpuStats *types.CPUStats) *ContainerCPUStats {
	return &ContainerCPUStats{
		Total:            util.Float64Ptr(float64(cpuStats.CPUUsage.TotalUsage)),
		System:           util.Float64Ptr(float64(cpuStats.CPUUsage.UsageInKernelmode)),
		User:             util.Float64Ptr(float64(cpuStats.CPUUsage.UsageInUsermode)),
		ThrottledPeriods: util.Float64Ptr(float64(cpuStats.ThrottlingData.ThrottledPeriods)),
		ThrottledTime:    util.Float64Ptr(float64(cpuStats.ThrottlingData.ThrottledTime)),
	}
}

func convertMemoryStats(memStats *types.MemoryStats) *ContainerMemStats {
	containerMemStats := &ContainerMemStats{
		UsageTotal: util.Float64Ptr(float64(memStats.Usage)),
		Limit:      util.Float64Ptr(float64(memStats.Limit)),
	}

	log.Infof("XXXXXXXXXXXX %#v\n", memStats.Stats)

	if rss, found := memStats.Stats["rss"]; found {
		containerMemStats.RSS = util.Float64Ptr(float64(rss))
	}

	if cache, found := memStats.Stats["cache"]; found {
		containerMemStats.Cache = util.Float64Ptr(float64(cache))
	}

	return containerMemStats
}

func convertIOStats(ioStats *types.BlkioStats) *ContainerIOStats {
	containerIOStats := ContainerIOStats{
		ReadBytes:       util.Float64Ptr(0),
		WriteBytes:      util.Float64Ptr(0),
		ReadOperations:  util.Float64Ptr(0),
		WriteOperations: util.Float64Ptr(0),
		Devices:         make(map[string]DeviceIOStats),
	}

	procPath := config.Datadog.GetString("container_proc_root")
	deviceMapping, err := getDiskDeviceMapping(procPath)
	if err != nil {
		log.Debugf("Error while getting disk mapping, no disk metric will be present, err: %w", err)
	}

	for _, blkioStatEntry := range ioStats.IoServiceBytesRecursive {
		deviceName, found := deviceMapping[fmt.Sprintf("%d:%d", blkioStatEntry.Major, blkioStatEntry.Minor)]

		var device DeviceIOStats
		if found {
			device = containerIOStats.Devices[deviceName]
		}

		switch blkioStatEntry.Op {
		case "Read":
			device.ReadBytes = util.Float64Ptr(float64(blkioStatEntry.Value))
			*containerIOStats.ReadBytes += *device.ReadBytes
		case "Write":
			device.WriteBytes = util.Float64Ptr(float64(blkioStatEntry.Value))
			*containerIOStats.WriteBytes += *device.WriteBytes
		}

		if found {
			containerIOStats.Devices[deviceName] = device
		}
	}

	for _, blkioStatEntry := range ioStats.IoServicedRecursive {
		deviceName, found := deviceMapping[fmt.Sprintf("%d:%d", blkioStatEntry.Major, blkioStatEntry.Minor)]

		var device DeviceIOStats
		if found {
			device = containerIOStats.Devices[deviceName]
		}

		switch blkioStatEntry.Op {
		case "Read":
			device.ReadOperations = util.Float64Ptr(float64(blkioStatEntry.Value))
			*containerIOStats.ReadOperations += *device.ReadOperations
		case "Write":
			device.WriteOperations = util.Float64Ptr(float64(blkioStatEntry.Value))
			*containerIOStats.WriteOperations += *device.WriteOperations
		}

		if found {
			containerIOStats.Devices[deviceName] = device
		}
	}

	return &containerIOStats
}

func convertPIDStats(pidStats *types.PidsStats) *ContainerPIDStats {
	return &ContainerPIDStats{
		ThreadCount: util.Float64Ptr(float64(pidStats.Current)),
		ThreadLimit: util.Float64Ptr(float64(pidStats.Limit)),
	}
}
