package kernel

type Event struct {
	time     uint64
	priority int
	process  *Process
	message  *Message
}
