package nnf

import (
	nvme "stash.us.cray.com/rabsw/nnf-ec/internal/nvme-namespace-manager"
)

type AllocationPolicy interface {
	Initialize(capacityBytes uint64) error
	CheckCapacity() bool
	Allocate() ([]ProvidingVolume, error)
}

type AllocationPolicyType string

const (
	SpareAllocationPolicyType        AllocationPolicyType = "spare"
	GlobalAllocationPolicyType                            = "global"
	SwitchLocalAllocationPolicyType                       = "switch-local"
	ComputeLocalAllocationPolicyType                      = "compute-local"
)

type AllocationStandardType string

const (
	StrictAllocationStandardType  AllocationStandardType = "strict"
	RelaxedAllocationStandardType                        = "relaxed"
)

const (
	DefaultAlloctionPolicy   = SpareAllocationPolicyType
	DefaultAlloctionStandard = StrictAllocationStandardType
)

type AllocationPolicyOem struct {
	Policy           string
	Standard         string
	ServerEndpointId string // This is a hint as to the server the allocation policy is designed for
}

// NewAllocationPolicy - Allocates a new Allocation Policy with the desired parameters.
// The provided config is the defaults as defined by the NNF Config (see config.yaml);
// Knowledgable users have the option to specify overrides if the default configuration
// is not as desired.
func NewAllocationPolicy(config AllocationConfig, overrides map[string]interface{}) AllocationPolicy {

	// TODO: Decode overrides

	policy := DefaultAlloctionPolicy
	standard := DefaultAlloctionStandard

	switch policy {
	default:
		return &SpareAllocationPolicy{standard: standard}
	}

}

/* ------------------------------ Spare Allocation Policy --------------------- */

type SpareAllocationPolicy struct {
	standard       AllocationStandardType
	storage        []*nvme.Storage
	capacityBytes  uint64
	allocatedBytes uint64
}

func (p *SpareAllocationPolicy) Initialize(capacityBytes uint64) error {
	storage := nvme.GetStorage()

	// TODO: Find the 16 least consumed storage devices
	
	p.storage = storage
	p.capacityBytes = capacityBytes

	return nil
}

func (p *SpareAllocationPolicy) CheckCapacity() bool {
	if p.capacityBytes == 0 {
		return false
	}

	var availableBytes = uint64(0)
	for _, s := range p.storage {
		availableBytes += s.UnallocatedBytes()
	}

	if availableBytes < p.capacityBytes {
		return false
	}

	if p.standard == RelaxedAllocationStandardType &&
		p.capacityBytes%uint64(len(p.storage)) != 0 {
		return false
	}

	return true
}

func (p *SpareAllocationPolicy) Allocate() ([]ProvidingVolume, error) {

	perStorageCapacityBytes := p.capacityBytes / uint64(len(p.storage))

	volumes := []ProvidingVolume{}
	for _, storage := range p.storage {

		v, err := nvme.CreateVolume(storage, perStorageCapacityBytes)

		if err != nil {
			//TODO: Rollyback i.e. defer policy.Deallocte()
			panic("Not Yet Implemented")
		}

		volumes = append(volumes, ProvidingVolume{volume: v})
	}

	return volumes, nil
}
