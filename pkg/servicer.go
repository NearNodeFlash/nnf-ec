package nnf

// DefaultApiServicer defines the interface into the NNF Redfish/Swordfish API.
// This is an abstract class  such that we can define different services for
// test and real hardware.
type DefaultApiServicer interface {

	// The Fabrics* set of methods defines and controls the switches, endpoints
	// and the connectivity of the PCIe fabric.
	RedfishV1FabricsGet(string) (string, error)
	RedfishV1FabricsFabricIdGet(string) (string, error)

	/*
		RedfishV1FabricsFabricIdSwitchesGet(string) (string, error)
		RedfishV1FabricsFabricIdSwitchesSwitchIdGet(string) (string, error)

		RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet(string) (string, error)
		RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet(string, string, *sf.PortV130Port)


		RedfishV1FabricsFabricIdSwitchesGet(string, *sf.SwitchCollectionSwitchCollection) error
		RedfishV1FabricsFabricIdSwitchesSwitchIdGet(string, string, *sf.SwitchV140Switch) error
		RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet(string, string, *sf.PortCollectionPortCollection) error
		RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet(string, string, *sf.PortV130Port)

		RedfishV1FabricsFabricIdEndpointsGet(string, *sf.EndpointCollectionEndpointCollection) error
		RedfishV1FabricsFabricIdEndpointsEndpointIdGet(string, string, *sf.EndpointV150Endpoint) error

		RedfishV1FabricsFabricIdConnectionsGet(string, *sf.ConnectionCollectionConnectionCollection) error
		RedfishV1FabricsFabricIdConnectionsPost(string, *sf.ConnectionV100Connection) (string, error)

		// The Storage* set of methods define the storage devices, controllers, and volumes
		// for the NVMe devices
		RedfishV1StorageGet(* sf.StorageCollectionStorageCollection) error
	*/
	/*
		RedfishV1StorageStorageIdGet(string) (string, error)
		RedfishV1StorageStorageIdStoragePoolsGet(string) (string, error)
		RedfishV1StorageStorageIdStoragePoolsStoragePoolIdGet(string) (string, error)
		RedfishV1StorageStorageIdStoragePoolsStoragePoolIdCapacitySourcesGet(string) (string, error)
		RedfishV1StorageStorageIdStoragePoolsStoragePoolIdCapacitySourcesCapacitySourceIdGet(string) (string, error)
	*/

	// The StorageServices set of methods define the storage pool, volumes and capacity sources that
	// represent the Near-Node Flash device.
}
