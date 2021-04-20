package server

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/uuid"

	"stash.us.cray.com/rabsw/nnf-ec/internal/common"
	"stash.us.cray.com/rabsw/nnf-ec/internal/logging"
)

type DefaultServerControllerProvider struct {}

func (DefaultServerControllerProvider) NewServerController(opts ServerControllerOptions) ServerControllerApi {
	if opts.Local {
	return &LocalServerController{}
	}
	return &RemoteServerController{}
}

type LocalServerController struct {
	pools []ServerStoragePool
}

func (c *LocalServerController) NewServerStoragePool(pid uuid.UUID) *ServerStoragePool {
	c.pools = append(c.pools, ServerStoragePool{id: pid, ctrl: c})
	return &c.pools[len(c.pools)-1]
}

func (c *LocalServerController) GetStatus(pool *ServerStoragePool) ServerStoragePoolStatus {

	// We really shouldn't need to refresh on every GetStatus() call if we're correctly
	// tracking udev add/remove events. There should be a single refresh on launch (or
	// possibily a udev-info call to pull in the initial hardware?)
	c.refresh()

	// There should always be 1 or more namespces, so zero namespaces mean we are still starting.
	// Once we've recovered the expected number of namespaces (nsExpected > 0) we continue to
	// return a starting status until all namespaces are available.

	// TODO: We should check for an error status somewhere... as this stands we are not
	// ever going to return an error if refresh() fails.
	if (len(pool.ns) == 0) ||
		((pool.nsExpected > 0) && (len(pool.ns) < pool.nsExpected)) {
		return ServerStoragePoolStarting
	}

	if pool.nsExpected == len(pool.ns) {
		return ServerStoragePoolReady
	}

	return ServerStoragePoolError
}

func (c *LocalServerController) CreateFileSystem(pool *ServerStoragePool, fs FileSystemApi, mp string) error {
	opts := FileSystemCreateOptions{
		"mountpoint": mp,
	}

	return fs.Create(pool.Devices(), opts)
}

func (c *LocalServerController) refresh() error {
	nss, err := c.namespaces()
	if err != nil {
		return err
	}

	for _, ns := range nss {
		sns, err := c.newStorageNamespace(ns)
		if err != nil {
			return err
		}

		if sns == nil {
			continue
		}

		pool := c.findPool(sns.poolId)
		if pool == nil {
			pool = c.NewServerStoragePool(sns.poolId)

			pool.nsExpected = sns.poolTotal

			pool.ns = append(pool.ns, *sns)
			continue
		}

		// We've identified a pool for this particular namespace
		// Add the namespace to the pool if it's not present.
		pool.UpsertStorageNamespace(sns)
	}

	return nil
}

func (c *LocalServerController) findPool(pid uuid.UUID) *ServerStoragePool {
	for idx, p := range c.pools {
		if p.id == pid {
			return &c.pools[idx]
		}
	}

	return nil
}

func (c *LocalServerController) namespaces() ([]string, error) {
	nss, err := c.run("ls -A1 /dev/nvme[0-1023]n[1-128]")

	// The above will return an err if zero namespaces exist. In
	// this case, ENOENT is returned and we should instead return
	// a zero length array.
	if exit, ok := err.(*exec.ExitError); ok {
		if syscall.Errno(exit.ExitCode()) == syscall.ENOENT {
			return make([]string, 0), nil
		}
	}

	return strings.Fields(string(nss)), err
}

func (c *LocalServerController) newStorageNamespace(path string) (*StorageNamespace, error) {
	// First we need to identify the namespace for the provided path.
	nsidstr, err := c.run(fmt.Sprintf(`nvme id-ns %s | awk 'match($0, "NVME Identify Namespace ([0-9])+:$", m) {printf m[1]}'`, path))
	if err != nil {
		return nil, err
	}

	nsid, err := strconv.Atoi(string(nsidstr))
	if err != nil {
		return nil, err
	}

	data, err := c.run(fmt.Sprintf("nvme get-feature %s --namespace-id=%d --feature-id=0x7F --raw-binary", path, nsid))
	if err != nil {
		return nil, err
	}

	meta, err := common.DecodeNamespaceMetadata(data[6:])
	if err != nil {
		return nil, err
	}

	return &StorageNamespace{
		path:      path,
		nsid:      nsid,
		poolId:    meta.Id,
		poolIdx:   int(meta.Index),
		poolTotal: int(meta.Count),
	}, nil
}

func (c *LocalServerController) run(cmd string) ([]byte, error) {
	return logging.Cli.Trace(cmd, func(cmd string) ([]byte, error) {
		return exec.Command("bash", "-c", cmd).Output()
	})
}