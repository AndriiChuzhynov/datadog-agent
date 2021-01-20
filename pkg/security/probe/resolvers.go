// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build linux

package probe

import (
	"context"
	"os"
	"sort"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/DataDog/gopsutil/process"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Resolvers holds the list of the event attribute resolvers
type Resolvers struct {
	probe             *Probe
	DentryResolver    *DentryResolver
	MountResolver     *MountResolver
	ContainerResolver *ContainerResolver
	TimeResolver      *TimeResolver
	ProcessResolver   *ProcessResolver
	UserGroupResolver *UserGroupResolver
}

// NewResolvers creates a new instance of Resolvers
func NewResolvers(probe *Probe, client *statsd.Client) (*Resolvers, error) {
	dentryResolver, err := NewDentryResolver(probe)
	if err != nil {
		return nil, err
	}

	timeResolver, err := NewTimeResolver()
	if err != nil {
		return nil, err
	}

	userGroupResolver, err := NewUserGroupResolver()
	if err != nil {
		return nil, err
	}

	resolvers := &Resolvers{
		probe:             probe,
		DentryResolver:    dentryResolver,
		MountResolver:     NewMountResolver(probe),
		TimeResolver:      timeResolver,
		ContainerResolver: &ContainerResolver{},
		UserGroupResolver: userGroupResolver,
	}

	processResolver, err := NewProcessResolver(probe, resolvers, client, ProcessResolverOpts{})
	if err != nil {
		return nil, err
	}

	resolvers.ProcessResolver = processResolver

	return resolvers, nil
}

// Start the resolvers
func (r *Resolvers) Start(ctx context.Context) error {
	if err := r.ProcessResolver.Start(ctx); err != nil {
		return err
	}

	if err := r.MountResolver.Start(); err != nil {
		return err
	}

	return r.DentryResolver.Start()
}

// Snapshot collects data on the current state of the system to populate user space and kernel space caches.
func (r *Resolvers) Snapshot() error {
	if err := retry.Do(r.snapshot, retry.Delay(0), retry.Attempts(5)); err != nil {
		return errors.Wrap(err, "unable to snapshot processes")
	}

	return nil
}

// snapshot internal version of Snapshot. Calls the relevant resolvers to sync their caches.
func (r *Resolvers) snapshot() error {
	// List all processes, to trigger the process and mount snapshots
	all, err := process.AllProcesses()
	if err != nil {
		return err
	}

	var processes []*process.FilledProcess
	for _, proc := range all {
		// If Exe is not set, the process is a short lived process and its /proc entry has already expired, move on.
		if len(proc.Exe) == 0 {
			continue
		}

		processes = append(processes, proc)
	}

	// make to insert them in the creation time order
	sort.Slice(processes, func(i, j int) bool {
		procA := processes[i]
		procB := processes[j]

		if procA.CreateTime == procB.CreateTime {
			return processes[i].Pid < processes[j].Pid
		}

		return procA.CreateTime < procB.CreateTime
	})

	cacheModified := false

	for _, proc := range processes {
		// Start with the mount resolver because the process resolver might need it to resolve paths
		if err := r.MountResolver.SyncCache(uint32(proc.Pid)); err != nil {
			if !os.IsNotExist(err) {
				log.Debug(errors.Wrapf(err, "snapshot failed for %d: couldn't sync mount points", proc.Pid))
			}
		}

		// Sync the process cache
		cacheModified = r.ProcessResolver.SyncCache(proc)
	}

	// There is a possible race condition when a process starts right after we called process.AllProcesses
	// and before we inserted the cache entry of its parent. Call Snapshot again until we do not modify the
	// process cache anymore
	if cacheModified {
		return errors.New("cache modified")
	}

	return nil
}
