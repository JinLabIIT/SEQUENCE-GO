package kernel

import (
	"fmt"
	"testing"
)

type Dummy struct {
	name string
}

func (d *Dummy) out(message Message) {
	fmt.Printf("name: %s, Message['info']: %s\n", d.name, message["info"])
}

func TestProcess_run(t *testing.T) {
	type fields struct {
		fnptr   func(message Message)
		message Message
	}
	d1 := Dummy{"alice"}
	d2 := Dummy{"bob"}
	msg1 := Message{"info": "say hello"}
	msg2 := Message{"info": "say world"}

	tests := []struct {
		name   string
		fields fields
	}{
		{"alice say1", fields{d1.out, msg1}},
		{"bob say1", fields{d2.out, msg1}},
		{"alice say2", fields{d1.out, msg2}},
		{"bob say2", fields{d2.out, msg2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Process{
				Fnptr:   tt.fields.fnptr,
				Message: tt.fields.message,
			}
			fmt.Printf("%s ", tt.name)
			p.run()
		})
	}
}
