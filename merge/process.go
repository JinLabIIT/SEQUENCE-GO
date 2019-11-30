package kernel

type Process struct {
	fnptr   func(message Message)
	message Message
	owner   *Entity
}

func (p *Process) run() {
	p.fnptr(p.message)
}
