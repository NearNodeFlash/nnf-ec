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

type DefaultServerControllerProvider struct{}

func (DefaultServerControllerProvider) NewServerController(opts ServerControllerOptions) ServerControllerApi {
	if opts.Local {
		return &LocalServerController{}
	}
	return &RemoteServerController{opts.Address}
}

type LocalServerController struct {
	storage []Storage
}

func (c *LocalServerController) NewStorage(pid uuid.UUID) *Storage {
	c.storage = append(c.storage, Storage{Id: pid, ctrl: c})
	return &c.storage[len(c.storage)-1]
}

func (c *LocalServerController) GetStatus(s *Storage) StorageStatus {

	// We really shouldn't need to refresh on every GetStatus() call if we're correctly
	// tracking udev add/remove events. There should be a single refresh on launch (or
	// possibily a udev-info call to pull in the initial hardware?)
	c.Discover(nil)

	// There should always be 1 or more namespces, so zero namespaces mean we are still starting.
	// Once we've recovered the expected number of namespaces (nsExpected > 0) we continue to
	// return a starting status until all namespaces are available.

	// TODO: We should check for an error status somewhere... as this stands we are not
	// ever going to return an error if refresh() fails.
	if (len(s.ns) == 0) ||
		((s.nsExpected > 0) && (len(s.ns) < s.nsExpected)) {
		return StorageStatus_Starting
	}

	if s.nsExpected == len(s.ns) {
		return StorageStatus_Ready
	}

	return StorageStatus_Error
}

func (c *LocalServerController) CreateFileSystem(s *Storage, fs FileSystemApi, mp string) error {
	s.fileSystem = fs

	opts := FileSystemCreateOptions{
		"mountpoint": mp,
	}

	return fs.Create(s.Devices(), opts)
}

func (c *LocalServerController) DeleteFileSystem(s *Storage) error {
	return s.fileSystem.Delete()
}

func (c *LocalServerController) Discover(newStorageFunc func(*Storage)) error {
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

		s := c.findStorage(sns.poolId)
		if s == nil {
			s = c.NewStorage(sns.poolId)

			s.nsExpected = sns.poolTotal

			s.ns = append(s.ns, *sns)

			if newStorageFunc != nil {
				newStorageFunc(s)
			}

			continue
		}

		// We've identified a pool for this particular namespace
		// Add the namespace to the pool if it's not present.
		s.UpsertStorageNamespace(sns)
	}

	return nil
}

func (c *LocalServerController) findStorage(pid uuid.UUID) *Storage {
	for idx, p := range c.storage {
		if p.Id == pid {
			return &c.storage[idx]
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
