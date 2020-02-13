package quantum

import "testing"

func Test_randGraph(t *testing.T) {
	type args struct {
		threadNum int
		lookAhead uint64
		path      string
	}

	tests := []struct {
		name string
		args args
	}{
		{"test1", args{2, uint64(5e7), "../../tools/1.json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			randGraph(tt.args.threadNum, tt.args.lookAhead, tt.args.path)
		})
	}
}
