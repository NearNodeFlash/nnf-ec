package common

type PortType int

const (
	UNKNOWN_PORT_TYPE PortType = iota
	USP_PORT_TYPE
	DSP_PORT_TYPE
)

type PortEventType int

const (
	PORT_EVENT_UNKNOWN PortEventType = iota
	PORT_EVENT_UP
	PORT_EVENT_DOWN
	PORT_EVENT_READY
)

// PortEvent -
type PortEvent struct {
	FabricId  string
	SwitchId  string
	PortId    string
	PortType  PortType
	EventType PortEventType
}

// PortEventHandlerFunc
type PortEventHandlerFunc func(PortEvent, interface{})

// PortEventSubscriber
type PortEventSubscriber struct {
	HandlerFunc PortEventHandlerFunc
	Data        interface{}
}

// PortEventManager -
var PortEventManager portEventManager

type portEventManager struct {
	subscribers []PortEventSubscriber
}

// Subscribe -
func (mgr *portEventManager) Subscribe(s PortEventSubscriber) {
	mgr.subscribers = append(mgr.subscribers, s)
}

// Publish
func (mgr *portEventManager) Publish(event PortEvent) {
	for _, s := range mgr.subscribers {
		s.HandlerFunc(event, s.Data)
	}
}
