package server

import (
	"github.com/google/uuid"
)

// Server Storage Pool represents a unique collection of Server Storage Volumes
// that are identified by a shared Storage Pool ID.
type ServerStoragePool struct {
	// ID is the Pool ID identified by recovering the NVMe Namespace Metadata
	// for this particular namespace. The Pool ID is common for all namespaces
	// which are part of that storage pool.
	id uuid.UUID

	ctrl ServerControllerApi

	ns []StorageNamespace

	nsExpected int // Expected number of namespaces within this storage pool
}

// Storage Namespace represents an NVMe Namespace present on the host.
type StorageNamespace struct {
	// Path refers to the system path for this particular NVMe Namespace. On unix
	// variants, the path is of the form `/dev/nvme[CTRL]n[INDEX]` where CTRL is the
	// parent NVMe Controller and INDEX is assigned by the operating system. INDEX
	// does _not_ refer to the namespace ID (NSID)
	path string

	nsid int

	id uuid.UUID

	poolId    uuid.UUID
	poolIdx   int // Index of this namespace wihin the storage pool
	poolTotal int // Total number of namespaces within the storage pool, as reported by this namespace

	debug bool // If this storage namespace is in debug mode
}

func (p *ServerStoragePool) GetStatus() ServerStoragePoolStatus {
	return p.ctrl.GetStatus(p)
}

func (p *ServerStoragePool) CreateFileSystem(fs FileSystemApi, mountpoint string) error {
	return p.ctrl.CreateFileSystem(p, fs, mountpoint)
}

// Returns the list of devices for this pool.
func (p *ServerStoragePool) Devices() []string {
	devices := make([]string, len(p.ns))
	for idx := range p.ns {
		devices[idx] = p.ns[idx].path
	}

	return devices
}

func (p *ServerStoragePool) UpsertStorageNamespace(sns *StorageNamespace) {
	for _, ns := range p.ns {
		// Debug mode uses matching paths to track the namespaces
		// This isn't practical in production because the same namespace
		// and come and go on different paths.
		if sns.debug {

			if ns.path == sns.path {
				return
			}
		} else {
			if ns.id == sns.id {
				return
			}
		}
	}

	p.ns = append(p.ns, *sns)
}

type ServerStoragePoolStatus string

const (
	ServerStoragePoolStarting ServerStoragePoolStatus = "starting"
	ServerStoragePoolReady                            = "ready"
	ServerStoragePoolError                            = "error"
)
