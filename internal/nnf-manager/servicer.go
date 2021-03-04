package nnf

import (
	"fmt"
	"net/http"

	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"

	. "stash.us.cray.com/rabsw/nnf-ec/internal/common"
)

// DefaultApiService -
type DefaultApiService struct{}

// NewDefaultApiService -
func NewDefaultApiService() Api {
	return &DefaultApiService{}
}

// RedfishV1StorageServicesGet -
func (*DefaultApiService) RedfishV1StorageServicesGet(w http.ResponseWriter, r *http.Request) {

	model := sf.StorageServiceCollectionStorageServiceCollection{
		OdataId:   "/redfish/v1/StorageServices",
		OdataType: "#StorageServiceCollection.v1_0_0.StorageServiceCollection",
		Name:      "Storage Service Collection",
	}

	err := Get(&model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	model := sf.StorageServiceV150StorageService{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s", storageServiceId),
		OdataType: "#StorageService.v1_5_0.StorageService",
		Id:        storageServiceId,
		Name:      "Storage Service",
	}

	err := StorageServiceIdGet(storageServiceId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStoragePoolsGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStoragePoolsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	model := sf.StoragePoolCollectionStoragePoolCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/StoragePools", storageServiceId),
		OdataType: "#StoragePoolCollection.v1_0_0.StoragePoolCollection",
		Name:      "Storage Pool Collection",
	}

	err := StorageServiceIdStoragePoolsGet(storageServiceId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStoragePoolsPost -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStoragePoolsPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	var model sf.StoragePoolV150StoragePool

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(model, err, w)
		return
	}

	if err := StorageServiceIdStoragePoolPost(storageServiceId, &model); err != nil {
		EncodeResponse(model, err, w)
		return
	}

	model.OdataId = fmt.Sprintf("/redfish/v1/StorageServices/%s/StoragePools/%s", storageServiceId, model.Id)
	model.OdataType = "#StoragePool.v1_5_0.StoragePool"
	model.Name = "Storage Pool"

	EncodeResponse(model, nil, w)
}

// RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storagePoolId := params["StoragePoolId"]

	model := sf.StoragePoolV150StoragePool{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/StoragePools/%s", storageServiceId, storagePoolId),
		OdataType: "#StoragePool.v1_5_0.StoragePool",
		Name:      "Storage Pool",
	}

	err := StorageServiceIdStoragePoolIdGet(storageServiceId, storagePoolId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdDelete -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdDelete(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storagePoolId := params["StoragePoolId"]

	err := StorageServiceIdStoragePoolIdDelete(storageServiceId, storagePoolId)

	EncodeResponse(nil, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdAllocatedVolumesGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdAllocatedVolumesGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storagePoolId := params["StoragePoolId"]

	model := sf.VolumeCollectionVolumeCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/StoragePools/%s/AllocateVolumes", storageServiceId, storagePoolId),
		OdataType: "#VolumeCollection.v1_0_0.VolumeCollection",
		Name:      "Allocated Volume Collection",
	}

	err := StorageServiceIdStoragePoolIdAlloctedVolumesGet(storageServiceId, storagePoolId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdProvidingVolumesGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStoragePoolsStoragePoolIdProvidingVolumesGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storagePoolId := params["StoragePoolId"]

	model := sf.VolumeCollectionVolumeCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/StoragePools/%s/ProvidingVolumes", storageServiceId, storagePoolId),
		OdataType: "#VolumeCollection.v1_0_0.VolumeCollection",
		Name:      "Providing Volume Collection",
	}

	err := StorageServiceIdStoragePoolIdProvidingVolumesGet(storageServiceId, storagePoolId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdVolumesGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdVolumesGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	model := sf.VolumeCollectionVolumeCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/Volumes", storageServiceId),
		OdataType: "#VolumeCollection.v1_0_0.VolumeCollection",
		Name:      "Volume Collection",
	}

	err := StorageServiceIdVolumesGet(storageServiceId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdVolumesVolumeIdGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdVolumesVolumeIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	volumeId := params["VolumeId"]

	model := sf.VolumeV161Volume{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/Volumes/%s", storageServiceId, volumeId),
		OdataType: "#Volume.v1_6_1.Volume",
		Name:      "Volume",
	}

	err := StorageServiceIdVolumeIdGet(storageServiceId, volumeId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdVolumesVolumeIdProvidingPoolsGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdVolumesVolumeIdProvidingPoolsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	volumeId := params["VolumeId"]

	model := sf.StoragePoolCollectionStoragePoolCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/Volumes/%s/ProvidingPools", storageServiceId, volumeId),
		OdataType: "#StoragePoolCollection.v1_0_0.StoragePoolCollection",
		Name:      "Storage Pool Collection",
	}

	err := StorageServiceIdVolumeIdProvidingPoolsGet(storageServiceId, volumeId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStorageGroupsGet
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStorageGroupsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	model := sf.StorageGroupCollectionStorageGroupCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/StorageGroups", storageServiceId),
		OdataType: "#StorageGroupCollection.v1_0_0.StorageGroupCollection",
		Name:      "Storage Group Collection",
	}

	err := StorageServiceIdStorageGroupsGet(storageServiceId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStorageGroupsPost -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStorageGroupsPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	var model sf.StorageGroupV150StorageGroup

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(model, err, w)
		return
	}

	if err := StorageServiceIdStorageGroupPost(storageServiceId, &model); err != nil {
		EncodeResponse(model, err, w)
		return
	}

	model.OdataId = fmt.Sprintf("/redfish/v1/StorageServices/%s/StorageGroup/%s", storageServiceId, model.Id)
	model.OdataType = "#StorageGroup.v1_5_0.StorageGroup"
	model.Name = "Storage Group"

	EncodeResponse(model, nil, w)
}

// RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdGet
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storageGroupId := params["StorageGroupId"]

	model := sf.StorageGroupV150StorageGroup{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/StorageGroup/%s", storageServiceId, storageGroupId),
		OdataType: "#StorageGroup.v1_5_0.StorageGroup",
		Name:      "Storage Group",
	}

	err := StorageServiceIdStorageGroupIdGet(storageServiceId, storageGroupId, &model)

	EncodeResponse(model, err, w)
}

//RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdDelete(w http.ResponseWriter, r *http.Request)
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdDelete(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storageGroupId := params["StorageGroupId"]

	err := StorageServiceIdStorageGroupIdDelete(storageServiceId, storageGroupId)

	EncodeResponse(nil, err, w)
}

// RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdActionsStorageGroupExposeVolumesPost -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdActionsStorageGroupExposeVolumesPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storageGroupId := params["StorageGroupId"]

	var model sf.StorageGroupV150ExposeVolumes

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(model, err, w)
		return
	}

	err := StorageServiceIdStorageGroupIdExposeVolumesPost(storageServiceId, storageGroupId, &model)

	EncodeResponse(model, err, w)
}

// 	RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdActionsStorageGroupHideVolumesPost -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdStorageGroupsStorageGroupIdActionsStorageGroupHideVolumesPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	storageGroupId := params["StorageGroupId"]

	var model sf.StorageGroupV150HideVolumes

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(nil, err, w)
		return
	}

	err := StorageServiceIdStorageGroupIdHideVolumesPost(storageServiceId, storageGroupId, &model)

	EncodeResponse(model, err, w)
}

// 	RedfishV1StorageServicesStorageServiceIdEndpointsGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdEndpointsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	model := sf.EndpointCollectionEndpointCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/Endpoints", storageServiceId),
		OdataType: "#EndpointCollection.v1_0_0.EndpointCollection",
		Name:      "Endpoint Collection",
	}

	err := StorageServiceIdEndpointsGet(storageServiceId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdFileSystemsGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	model := sf.FileSystemCollectionFileSystemCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/FileSystems", storageServiceId),
		OdataType: "#FileSystemCollection.v1_0_0.FileSystemCollection",
		Name:      "File System Collection",
	}

	err := StorageServiceIdFileSystemsGet(storageServiceId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdFileSystemsPost -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]

	var model sf.FileSystemV122FileSystem

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(nil, err, w)
		return
	}

	if err := StorageServiceIdFileSystemsPost(storageServiceId, &model); err != nil {
		EncodeResponse(nil, err, w)
		return
	}

	model.OdataId = fmt.Sprintf("/redfish/v1/StorageServices/%s/FileSystems/%s", storageServiceId, model.Id)
	model.OdataType = "#FileSystem.v1_2_2.FileSystem"
	model.Name = "File System"

	EncodeResponse(model, nil, w)
}

// RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemIdGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	fileSystemId := params["FileSystemId"]

	model := sf.FileSystemV122FileSystem{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/FileSystems/%s", storageServiceId, fileSystemId),
		OdataType: "#FileSystem.v1_2_2.FileSystem",
		Name:      "File System",
	}

	err := StorageServiceIdFileSystemIdGet(storageServiceId, fileSystemId, &model)

	EncodeResponse(model, err, w)
}

// 	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemIdDelete -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemIdDelete(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	fileSystemId := params["FileSystemId"]

	err := StorageServiceIdFileSystemIdDelete(storageServiceId, fileSystemId)

	EncodeResponse(nil, err, w)
}

// 	RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	fileSystemId := params["FileSystemId"]

	model := sf.FileShareCollectionFileShareCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/FileSystems/%s/ExportedShares", storageServiceId, fileSystemId),
		OdataType: "#FileShareCollection.v1_0_0.FileShareCollection",
		Name:      "File Share Collection",
	}

	err := StorageServiceIdFileSystemIdExportedSharesGet(storageServiceId, fileSystemId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesPost -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	fileSystemId := params["FileSystemId"]

	var model sf.FileShareV120FileShare

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(model, err, w)
		return
	}

	err := StorageServiceIdFileSystemIdExportedSharesPost(storageServiceId, fileSystemId, &model)

	model.OdataId = fmt.Sprintf("/redfish/v1/StorageServices/%s/FileSystems/%s/ExportedShares/%s", storageServiceId, fileSystemId, model.Id)
	model.OdataType = "#FileShare.v1_2_0.FileShare"
	model.Name = "Exported File Share"

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesExportedFileSharesIdGet -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesExportedFileSharesIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	fileSystemId := params["FileSystemId"]
	exportedShareId := params["ExportedShareId"]

	model := sf.FileShareV120FileShare{
		OdataId:   fmt.Sprintf("/redfish/v1/StorageServices/%s/FileSystems/%s/ExportedShares/%s", storageServiceId, fileSystemId, exportedShareId),
		OdataType: "#FileShare.v1_2_0.FileShare",
		Name:      "Exported File Share",
	}

	err := StorageServiceIdFileSystemIdExportedShareIdGet(storageServiceId, fileSystemId, exportedShareId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesExportedFileSharesIdDelete -
func (*DefaultApiService) RedfishV1StorageServicesStorageServiceIdFileSystemsFileSystemsIdExportedFileSharesExportedFileSharesIdDelete(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	storageServiceId := params["StorageServiceId"]
	fileSystemId := params["FileSystemId"]
	exportedShareId := params["ExportedShareId"]

	err := StorageServiceIdFileSystemIdExportedShareIdDelete(storageServiceId, fileSystemId, exportedShareId)

	EncodeResponse(nil, err, w)
}
