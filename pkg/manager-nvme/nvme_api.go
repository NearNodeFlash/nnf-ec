/*
 * Copyright 2020-2025 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nvme

import (
	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/nvme"
)

type NvmeController interface {
	NewNvmeDeviceController() NvmeDeviceController
}

type NvmeDeviceController interface {
	Initialize() error
	Close() error

	NewNvmeDevice(fabricId, switchId, portId string) (NvmeDeviceApi, error)
}

// NvmeDeviceApi -
type NvmeDeviceApi interface {
	IsDirectDevice() bool

	IdentifyController(controllerId uint16) (*nvme.IdCtrl, error)
	IdentifyNamespace(namespaceId nvme.NamespaceIdentifier) (*nvme.IdNs, error)

	ListSecondary() (*nvme.SecondaryControllerList, error)

	AssignControllerResources(
		controllerId uint16,
		resourceType SecondaryControllerResourceType,
		numResources uint32) error

	OnlineController(controllerId uint16) error

	ListNamespaces(controllerId uint16) ([]nvme.NamespaceIdentifier, error)
	ListAttachedControllers(namespaceId nvme.NamespaceIdentifier) ([]uint16, error)

	CreateNamespace(sizeInSectors uint64, sectorSizeIndex uint8) (nvme.NamespaceIdentifier, nvme.NamespaceGloballyUniqueIdentifier, error)
	DeleteNamespace(namespaceId nvme.NamespaceIdentifier) error

	FormatNamespace(namespaceID nvme.NamespaceIdentifier) error

	AttachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error
	DetachNamespace(namespaceId nvme.NamespaceIdentifier, controllers []uint16) error

	SetNamespaceFeature(namespaceId nvme.NamespaceIdentifier, data []byte) error
	GetNamespaceFeature(namespaceId nvme.NamespaceIdentifier) ([]byte, error)

	// GetSmartLog returns the raw SMART log page data
	GetSmartLog() (*nvme.SmartLog, error)
}

// SecondaryControllersInitFunc -
type SecondaryControllersInitFunc func(count uint8)

// SecondaryControllerHandlerFunc -
type SecondaryControllerHandlerFunc func(controllerId uint16, controllerOnline bool, virtualFunctionNumber uint16, numVQResourcesAssinged, numVIResourcesAssigned uint32) error

// SecondaryControllerResourceType -
type SecondaryControllerResourceType int

const (
	VQResourceType SecondaryControllerResourceType = iota
	VIResourceType
)

// GetWearLevelAsPercentageUsed returns the PercentageUsed field from the SMART log page.
func GetWearLevelAsPercentageUsed(dev NvmeDeviceApi) (uint8, error) {
	smartLog, err := dev.GetSmartLog()
	if err != nil {
		return 0, err
	}
	return smartLog.PercentageUsed, nil
}
