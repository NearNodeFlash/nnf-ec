package nnf

import (
	"net/http"
)

type Api interface {
	RedfishV1StorageServicesGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdGet(w http.ResponseWriter, r *http.Request)

	RedfishV1StorageServicesStorageServiceIdStoragePoolsGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStoragePoolsPost(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdDelete(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdAllocatedVolumesGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdProvidingVolumesGet(w http.ResponseWriter, r *http.Request)

	RedfishV1StorageServicesStorageServiceIdVolumesGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdVolumesVolumeIdGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdVolumesVolumeIdProvidingPoolsGet(w http.ResponseWriter, r *http.Request)

	RedfishV1StorageServicesStorageServiceIdStorageGroupsGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStorageGroupsPost(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdDelete(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdActionsStorageGroupExposeVolumesPost(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdActionsStorageGroupHideVolumesPost(w http.ResponseWriter, r *http.Request)

	RedfishV1StorageServicesStorageServiceIdEndpointsGet(w http.ResponseWriter, r *http.Request)

	RedfishV1StorageServicesStorageServiceIdFileSystemsGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdFileSystemsPost(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemIdGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemIdDelete(w http.ResponseWriter, r *http.Request)

	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesPost(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesExportedFileSharesIdGet(w http.ResponseWriter, r *http.Request)
	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesExportedFileSharesIdDelete(w http.ResponseWriter, r *http.Request)
}
