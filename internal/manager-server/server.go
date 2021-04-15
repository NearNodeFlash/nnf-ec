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

// Server Controller is responsible for receiving commands from the NNF Manager
// and dictating to the underlying server hardware the necessary actions to
// fullfil the NNF Manager request.
//
// The Server Controller supports two modes of operation.
//   1) Local Server Controller - This is expected to run locally to the NNF
//      manager. The necessary hardware and software components are expected
//      to be readily accessable to the local server controller.
//
//   2) Remote Server Controller - This controller is running apart from the
//      NNF Manager, with access to this controller through some non-local
//      means like TCP/IP, Non-Transparent Bridge, Remote DMA, etc.
//
//
//
// Server Controller API:
//   1) Identify Storage - This is a <Storage Pool, Server> pair, when a
//      Storage Group is created by the NNF Manager, the NVMe Namespaces that
//      make up the Storage Pool are attached to the Server. The Server Controller
//      should be able to validate the existence of the Storage Pool on the
//      particular target server.
//
//   2) Create File System & File Share - This is a quad of <Storage Group,
//      File System, File Share, and Server> that constitues a file system on the
//      server given a list of devices.
//
//   3) Delete File System -
//   4) Delete Storage Group -
//
//   5) Certain File System Actions (Sync, Snapshot, Query, etc.)
//

type ServerControllerApi interface {
	NewServerStoragePool(pid uuid.UUID) *ServerStoragePool

	// These
	GetStatus(*ServerStoragePool) ServerStoragePoolStatus
	CreateFileSystem(*ServerStoragePool, FileSystemApi, string) error
}

func NewServerController(local bool) ServerControllerApi {
	// TODO: The server controller should be constructed based on
	//       where the Server Endpoint resides. Here we just assume
	//       NNF local.
	return &LocalServerController{}
}

// Design:
//
// udev events for namespace add are monitored by the Storage Controller (see storage.go)
// Each time a namespace is added, the Storage Controller recovers the namespace metadata.
// This allows the Server Controller to gather the devices in the storage pool and prepare
// for additional operations.
//
// Each NNF created NS contains immuntable data
//   1) the Storage Pool UUID
//   2) the number of namespaces in the pool
//   3) this namespace index within the pool
//
// A Local Server Controller can be queried by the NNF Manager for the status of the
// Storage Pool - this is essential to bringin the Storage Group from an Offline to an
// Online state.
//
// A Remote Server Controller has a harder time of reporting status to the NNF Manager
// since it doesn't (currently) have a good way to communicate to the NNF controller.
// There are several proposals to be considered
//   1) High-Speed Network - This should be possible since the IP address follow a well
//      defined scheme, so if the NNF controller knows its IP, it knows the IPs of its
//      near compute blades and vice-versa. Another advantage to this is we can present
//      the same channel for customers that need a communication channel between compute
//      and NNF.
//   2) DMA between server and NNF
//   3) Using NVM Express Management Interface Metadata - this uses the managed data
//      on the NVMe drives (either controller or namespace). This is similar to how the
//      namespaces are labeled by the NNF Manager for identification, but would add mutable
//      sections where status + updates fields are written by the remote server.
//
// In any case, the Remote Server Controller API should be agnostic as to the underlying
// communication channel.

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

// RemoteServerController - Stub
type RemoteServerController struct{}

func (*RemoteServerController) GetStatus(*ServerStoragePool) ServerStoragePoolStatus { return "" }
