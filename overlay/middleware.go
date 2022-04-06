package overlay

type NetworkWillStart struct {
	Func     func(Network) bool
	Priority int32
}

type NetworkStarted struct {
	Func     func(Network) bool
	Priority int32
}

type NetworkWillStop struct {
	Func     func(Network) bool
	Priority int32
}

type NetworkStopped struct {
	Func     func(Network) bool
	Priority int32
}
