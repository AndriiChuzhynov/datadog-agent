// +build linux_bpf

package http

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unsafe"

	ddebpf "github.com/DataDog/datadog-agent/pkg/ebpf"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	psfilepath "github.com/DataDog/gopsutil/process/filepath"
	"github.com/DataDog/gopsutil/process/so"
)

/*
#include "../ebpf/c/http-types.h"
*/
import "C"

const pathMaxSize = int(C.LIB_PATH_MAX_SIZE)

type libPath = C.lib_path_t

func toLibPath(data []byte) C.lib_path_t {
	return *(*C.lib_path_t)(unsafe.Pointer(&data[0]))
}

func (l *libPath) Bytes() []byte {
	b := *(*[pathMaxSize]byte)(unsafe.Pointer(&l.buf))
	return b[:l.len]
}

// syncInterval controls the frequency at which /proc/<PID>/maps are inspected.
// this is to ensure that we remove/deregister the shared libraries that are no
// longer mapped into memory.
const soSyncInterval = 5 * time.Minute

type soRule struct {
	re           *regexp.Regexp
	registerCB   func(string) error
	unregisterCB func(string) error
}

// soWatcher provides a way to tie callback functions to the lifecycle of shared libraries
type soWatcher struct {
	procRoot   string
	hostMount  string
	all        *regexp.Regexp
	rules      []soRule
	registered map[string]func(string) error
	loadEvents *ddebpf.PerfHandler
}

type seenKey struct {
	pid, path string
}

func newSOWatcher(procRoot string, perfHandler *ddebpf.PerfHandler, rules ...soRule) *soWatcher {
	allFilters := make([]string, len(rules))
	for i, r := range rules {
		allFilters[i] = r.re.String()
	}

	all := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(allFilters, "|")))
	return &soWatcher{
		procRoot:   procRoot,
		hostMount:  os.Getenv("HOST_ROOT"),
		all:        all,
		rules:      rules,
		loadEvents: perfHandler,
	}
}

// Start consuming shared-library events
func (w *soWatcher) Start() {
	seen := make(map[seenKey]struct{})
	sharedLibraries := w.getSharedLibraries()
	w.sync(sharedLibraries)
	go func() {
		ticker := time.NewTicker(soSyncInterval)
		defer ticker.Stop()
		thisPID := os.Getpid()

		for {
			select {
			case <-ticker.C:
				seen = make(map[seenKey]struct{})
				sharedLibraries := w.getSharedLibraries()
				w.sync(sharedLibraries)
			case event, ok := <-w.loadEvents.DataChannel:
				if !ok {
					return
				}

				lib := toLibPath(event.Data)
				// if this shared library was loaded by system-probe we ignore it.
				// this is to avoid a feedback-loop since the shared libraries here monitored
				// end up being opened by system-probe
				if int(lib.pid) == thisPID {
					break
				}

				path := lib.Bytes()
				for _, r := range w.rules {
					if r.re.Match(path) {
						var (
							libPath = string(path)
							pidPath = fmt.Sprintf("%s/%d", w.procRoot, lib.pid)
						)

						// resolving paths is expensive so we cache the libraries we've already seen
						k := seenKey{pidPath, libPath}
						if _, ok := seen[k]; ok {
							break
						}
						seen[k] = struct{}{}

						// resolve namespaced path to host path
						pathResolver := psfilepath.NewResolver(w.procRoot)
						pathResolver.LoadPIDMounts(pidPath)
						if hostPath := pathResolver.Resolve(libPath); hostPath != "" {
							libPath = hostPath
						}

						libPath = w.canonicalizePath(libPath)
						if _, registered := w.registered[libPath]; registered {
							break
						}

						w.register(libPath, r)
						break
					}
				}
			case <-w.loadEvents.LostChannel:
				// Nothing to do in this case
				break
			}
		}
	}()
}

func (w *soWatcher) sync(libraries []so.Library) {
	old := w.registered
	w.registered = make(map[string]func(string) error)

	for _, lib := range libraries {
		path := lib.HostPath
		if callback, ok := old[path]; ok {
			w.registered[path] = callback
			delete(old, path)
			continue
		}

		for _, r := range w.rules {
			if r.re.MatchString(path) {
				w.register(path, r)
				break
			}
		}
	}

	// Now we call the unregister callback for every shared library that is no longer mapped into memory
	for path, unregisterCB := range old {
		if unregisterCB == nil {
			continue
		}

		log.Debugf("unregistering library=%s", path)
		unregisterCB(path)
	}
}

func (w *soWatcher) register(libPath string, r soRule) {
	err := r.registerCB(libPath)
	if err != nil {
		log.Debugf("error registering library=%s: %s", libPath, err)
		r.unregisterCB(libPath)
		w.registered[libPath] = nil
		return
	}

	log.Debugf("registering library=%s", libPath)
	w.registered[libPath] = r.unregisterCB
}

func (w *soWatcher) canonicalizePath(path string) string {
	if w.hostMount != "" {
		path = filepath.Join(w.hostMount, path)
	}

	return followSymlink(path)
}

func (w *soWatcher) getSharedLibraries() []so.Library {
	// libraries will include all host-resolved library paths mapped into memory
	libraries := so.FindProc(w.procRoot, w.all)

	// TODO: should we ensure all entries are unique in the `so` package instead?
	seen := make(map[string]struct{}, len(libraries))
	i := 0
	for _, lib := range libraries {
		originalPath := lib.HostPath
		if _, ok := seen[originalPath]; ok {
			continue
		}
		seen[originalPath] = struct{}{}

		// this ensures that all symlinks are resolved only once
		canonicalPath := w.canonicalizePath(originalPath)
		if canonicalPath != originalPath {
			if _, ok := seen[canonicalPath]; ok {
				continue
			} else {
				seen[canonicalPath] = struct{}{}
				lib.HostPath = canonicalPath
			}
		}

		libraries[i] = lib
		i++
	}

	return libraries[0:i]
}

func followSymlink(path string) string {
	if withoutSymLinks, err := filepath.EvalSymlinks(path); err == nil {
		return withoutSymLinks
	}

	return path
}
