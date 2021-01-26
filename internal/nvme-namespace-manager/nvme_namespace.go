package nvmenamespace

import (
	"fmt"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"

	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
)

const (
	defaultStoragePoolId = "1"
)

type Storage struct {
}

type StorageController struct {
}

func isStorageId(id string) bool     { _, err := findStorage(id); return err == nil }
func isStoragePoolId(id string) bool { return id == defaultStoragePoolId }

func findStorage(id string) (*Storage, error) {
	return nil, ec.ErrNotFound
}

func findStorageController(storageId string, controllerId string) (*StorageController, error) {
	return nil, ec.ErrNotFound
}

func Get(model *sf.StorageCollectionStorageCollection) error {
	// TODO: Return the list of NVMe devices
	return nil
}

func StorageIdGet(storageId string, model *sf.StorageV190Storage) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	model.Controllers.OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StorageControllers", storageId)

	return nil
}

func StorageIdStoragePoolsGet(storageId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	model.MembersodataCount = 1
	model.Members = make([]sf.OdataV4IdRef, model.MembersodataCount)
	model.Members[0].OdataId = fmt.Sprintf("/redfish/v1/Storage/%s/StoragePools/%s", storageId, defaultStoragePoolId)

	return nil
}

func StorageIdStoragePoolIdGet(storageId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	if !isStorageId(storageId) || !isStoragePoolId(storagePoolId) {
		return ec.ErrNotFound
	}

	model.Capacity = sf.CapacityV100Capacity{
		Data: sf.CapacityV100CapacityInfo{
			AllocatedBytes:   0,
			ConsumedBytes:    0,
			GuaranteedBytes:  0,
			ProvisionedBytes: 0,
		},
	}

	model.RemainingCapacityPercent = 0

	return nil
}

func StorageIdControllersGet(storageId string, model *sf.StorageControllerCollectionStorageControllerCollection) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	// TODO: Query the NVMe storage device for # controllers

	return nil
}

func StorageIdControllerIdGet(storageId, controllerId string, model *sf.StorageControllerV100StorageController) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	_, err := findStorageController(storageId, controllerId)
	if err != nil {
		return err
	}

	// TODO: Link Endpoints
	model.Links.EndpointsodataCount = 1
	model.Links.Endpoints = make([]sf.OdataV4IdRef, model.Links.EndpointsodataCount)

	return nil
}

func StorageIdVolumesGet(storageId string, model *sf.VolumeCollectionVolumeCollection) error {
	if !isStorageId(storageId) {
		return ec.ErrNotFound
	}

	// TODO: Query the volumes on the device

	return nil
}

func StorageIdVolumePost(storageId string, model *sf.VolumeV161Volume) (string, error) {
	return "", nil
}

func StorageIdVolumeIdGet(storageId, volumeId string, model *sf.VolumeV161Volume) error {
	return nil
}

func StorageIdVolumeIdDelete(storageId, volumeId string) error {
	return nil
}
