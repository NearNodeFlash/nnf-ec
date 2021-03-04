package nnf

import (
	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
)

func Get(model *sf.StorageServiceCollectionStorageServiceCollection) error {
	return nil
}

func StorageServiceIdGet(storageServiceId string, model *sf.StorageServiceV150StorageService) error {
	return nil
}

func StorageServiceIdStoragePoolsGet(storageServiceId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	return nil
}

func StorageServiceIdStoragePoolPost(storageServiceId string, model *sf.StoragePoolV150StoragePool) error {
	return nil
}

func StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId string, model *sf.StoragePoolV150StoragePool) error {
	return nil
}

func StorageServiceIdStoragePoolIdDelete(storageServiceId, storagePoolId string) error {
	return nil
}

func StorageServiceIdStoragePoolIdAlloctedVolumesGet(storageServiceId, storagePoolId string, model *sf.VolumeCollectionVolumeCollection) error {
	return nil
}

func StorageServiceIdStoragePoolIdProvidingVolumesGet(storageServiceId, storagePoolId string, model *sf.VolumeCollectionVolumeCollection) error {
	return nil
}

func StorageServiceIdVolumesGet(storageServiceId string, model *sf.VolumeCollectionVolumeCollection) error {
	return nil
}

func StorageServiceIdVolumeIdGet(storageServiceId, volumeId string, model *sf.VolumeV161Volume) error {
	return nil
}

func StorageServiceIdVolumeIdProvidingPoolsGet(storageServiceId, volumeId string, model *sf.StoragePoolCollectionStoragePoolCollection) error {
	return nil
}

func StorageServiceIdStorageGroupsGet(storageServiceId string, model *sf.StorageGroupCollectionStorageGroupCollection) error {
	return nil
}

func StorageServiceIdStorageGroupPost(storageServiceId string, model *sf.StorageGroupV150StorageGroup) error {
	return nil
}

func StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId string, model *sf.StorageGroupV150StorageGroup) error {
	return nil
}

func StorageServiceIdStorageGroupIdDelete(storageServiceId, storageGroupId string) error {
	return nil
}

func StorageServiceIdStorageGroupIdExposeVolumesPost(storageServiceId, storageGroupId string, model *sf.StorageGroupV150ExposeVolumes) error {
	return nil
}

func StorageServiceIdStorageGroupIdHideVolumesPost(storageServiceId, storageGroupId string, model *sf.StorageGroupV150HideVolumes) error {
	return nil
}

func StorageServiceIdEndpointsGet(storageServiceId string, model *sf.EndpointCollectionEndpointCollection) error {
	return nil
}

func StorageServiceIdFileSystemsGet(storageServiceId string, model *sf.FileSystemCollectionFileSystemCollection) error {
	return nil
}

func StorageServiceIdFileSystemsPost(storageServiceId string, model *sf.FileSystemV122FileSystem) error {
	return nil
}

func StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId string, model *sf.FileSystemV122FileSystem) error {
	return nil
}

func StorageServiceIdFileSystemIdDelete(storageServiceId, fileSystemId string) error {
	return nil
}

func StorageServiceIdFileSystemIdExportedSharesGet(storageServiceId, fileSystemId string, model *sf.FileShareCollectionFileShareCollection) error {
	return nil
}

func StorageServiceIdFileSystemIdExportedSharesPost(storageServiceId, fileSystemId string, model *sf.FileShareV120FileShare) error {
	return nil
}

func StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId string, model *sf.FileShareV120FileShare) error {
	return nil
}

func StorageServiceIdFileSystemIdExportedShareIdDelete(storageServiceId, fileSystemId, exportedShareId string) error {
	return nil
}
