package kernel

type Event struct {
	Time     uint64
	Priority uint
	Process  *Process
}
