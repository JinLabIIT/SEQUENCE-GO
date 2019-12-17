package kernel

type Process struct {
	Fnptr   func(message Message)
	Message Message
	Owner   *Timeline
}

func (p *Process) run() {
	p.Fnptr(p.Message)
}
