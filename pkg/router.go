package nnf

// Router contains all the Redfish / Swordfish API calls that are hosted by
// the NNF module. All handler calls are of the form RedfishV1{endpoint}, where
// and endpoint is unique to the RF/SF caller.
//
// Router calls must reflect the same function name as the RF/SF API, as the
// element controller will perform a 1:1 function call based on the RF/SF
// caller's name.
type router struct {
	service DefaultApiServicer
}

/*
func NewDefaultRouter(s DefaultApiServicer) *router {
	return &router{service: s}
}

func (rtr *router) RedfishV1FabricsGet(_ string) (string, error) {
	if err := r.service.RedfishV1FabricsGet(fabrics); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(fabrics)
}

// RedfishV1FabricsFabricIdGet -
func (rtr *router) RedfishV1FabricsFabricIdGet(r string) (string, error) {
	params := utils.Params(r)
	id := params["fabricId"]

	fabric := sf.FabricV120Fabric{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s", id),
		OdataType: "#Fabric.v1_2_0.Fabric",
		Id:        id,
		Name:      "Fabric",
	}

	if err := rtr.service.RedfishV1FabricsFabricIdGet(id, fabric); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(fabric)
}

func (rtr *router) RedfishV1FabricsFabricIdSwitchesGet(r string) (string, error) {
	params := utils.Params(r)
	id := params["fabricId"]

	switches := sf.SwitchCollectionSwitchCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches", id),
		OdataType: "#SwitchCollection.v1_0_0.SwitchCollection",
		Name:      "Switch Collection",
	}

	if err := rtr.service.RedfishV1FabricsFabricIdSwitchesGet(id, switches); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(fabric)
}

func (rtr *router) RedfishV1FabricsFabricIdSwitchesSwitchIdGet(r string) (string, error) {
	params := utils.Params(req)
	fabricId := params["fabricId"]

}

// RedfishV1StorageGet -
func (*handler) RedfishV1StorageGet(r string) (string, error) {
	storage := sf.StorageCollectionStorageCollection{
		OdataContext: "/redfish/v1/$metadata#StorageCollection.StorageCollection",
		OdataType:    "#StorageCollection.v1_0_0.StorageCollection",
		OdataId:      "/redfish/v1/Storage",
		Description:  "",
		Name:         "Storage Collection",
	}

	if err := r.service.RedfishV1StorageGet(storage); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(storage)
}

// RedfishV1StorageStorageIdGet -
func (*handler) RedfishV1StorageStorageIdGet(r string) (string, error) {
	params := utlis.Params(r)
	id := params["storageId"]

	storage := sf.StorageV190Storage{
		OdataContext:       "/redfish/v1/",
		OdataType:          "#Storage.1.9.0.Storage",
		OdataId:            "/redfish/v1/Storage/",
		StorageControllers: nil,
	}

	return encodeJsonResponse(storage)
}

// RedfishV1StorageStorageIdStoragePoolsGet -
func (*handler) RedfishV1StorageStorageIdStoragePoolsGet(r string) (string, error) {

}


*/
