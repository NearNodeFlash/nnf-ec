//go:build darwin
// +build darwin

package remote

type UDevMonitor struct {
	exit chan struct{}
}

func (*UDevMonitor) Open() error      { return nil }
func (*UDevMonitor) Close() error     { return nil }
func (*UDevMonitor) Run() chan UEvent { return make(chan UEvent, 1) }

type UEvent struct{}

func (*UEvent) IsNvmeEvent() bool { return false }
