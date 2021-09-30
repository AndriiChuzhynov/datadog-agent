//go:build linux_bpf
// +build linux_bpf

package tracer

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/network"
	"github.com/DataDog/datadog-agent/pkg/network/netlink"
	"github.com/DataDog/datadog-agent/pkg/process/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	ct "github.com/florianl/go-conntrack"
	"github.com/golang/groupcache/lru"
	"golang.org/x/sys/unix"
)

type cachedConntrack struct {
	sync.Mutex
	procRoot         string
	cache            *lru.Cache
	conntrackCreator func(int) (netlink.Conntrack, error)
	closed           bool
}

func newCachedConntrack(procRoot string, conntrackCreator func(int) (netlink.Conntrack, error), size int) *cachedConntrack {
	cache := &cachedConntrack{
		procRoot:         procRoot,
		conntrackCreator: conntrackCreator,
		cache:            lru.New(size),
	}

	cache.cache.OnEvicted = func(_ lru.Key, v interface{}) {
		v.(netlink.Conntrack).Close()
	}

	return cache
}

func (cache *cachedConntrack) Close() error {
	cache.Lock()
	defer cache.Unlock()

	cache.cache.Clear()
	cache.closed = true
	return nil
}

func (cache *cachedConntrack) ExistsInRootNS(c *network.ConnectionStats) (bool, error) {
	return cache.exists(c, 0, 1)
}

func (cache *cachedConntrack) Exists(c *network.ConnectionStats) (bool, error) {
	return cache.exists(c, c.NetNS, int(c.Pid))
}

func (cache *cachedConntrack) exists(c *network.ConnectionStats, netns uint32, pid int) (bool, error) {
	ctrk, err := cache.ensureConntrack(uint64(netns), pid)
	if err != nil {
		return false, err
	}

	if ctrk == nil {
		return false, nil
	}

	var protoNumber uint8 = unix.IPPROTO_UDP
	if c.Type == network.TCP {
		protoNumber = unix.IPPROTO_TCP
	}

	srcBuf := util.IPBufferPool.Get().([]byte)
	dstBuf := util.IPBufferPool.Get().([]byte)
	defer func() {
		util.IPBufferPool.Put(srcBuf)
		util.IPBufferPool.Put(dstBuf)
	}()

	srcAddr, dstAddr := util.NetIPFromAddress(c.Source, srcBuf), util.NetIPFromAddress(c.Dest, dstBuf)
	srcPort, dstPort := c.SPort, c.DPort

	conn := netlink.Con{
		Con: ct.Con{
			Origin: &ct.IPTuple{
				Src: &srcAddr,
				Dst: &dstAddr,
				Proto: &ct.ProtoTuple{
					Number:  &protoNumber,
					SrcPort: &srcPort,
					DstPort: &dstPort,
				},
			},
		},
	}

	ok, err := ctrk.Exists(&conn)
	if err != nil {
		log.Debugf("error while checking conntrack for connection %#v: %s", conn, err)
		cache.removeConntrack(uint64(netns))
		return false, err
	}

	if ok {
		return ok, nil
	}

	conn.Reply = conn.Origin
	conn.Origin = nil
	ok, err = ctrk.Exists(&conn)
	if err != nil {
		log.Debugf("error while checking conntrack for connection %#v: %s", conn, err)
		cache.removeConntrack(uint64(netns))
		return false, err
	}

	return ok, nil
}

func (cache *cachedConntrack) removeConntrack(ino uint64) {
	cache.Lock()
	defer cache.Unlock()

	cache.cache.Remove(ino)
}

func (cache *cachedConntrack) ensureConntrack(ino uint64, pid int) (netlink.Conntrack, error) {
	cache.Lock()
	defer cache.Unlock()

	if cache.closed {
		return nil, fmt.Errorf("cache Close has already been called")
	}

	v, ok := cache.cache.Get(ino)
	if ok {
		return v.(netlink.Conntrack), nil
	}

	ns, err := util.GetNetNamespaceFromPid(cache.procRoot, pid)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		log.Errorf("could not get net ns for pid %d: %s", pid, err)
		return nil, err
	}
	defer ns.Close()

	ctrk, err := cache.conntrackCreator(int(ns))
	if err != nil {
		log.Errorf("could not create conntrack object for net ns %d: %s", ino, err)
		return nil, err
	}

	cache.cache.Add(ino, ctrk)
	return ctrk, nil
}
