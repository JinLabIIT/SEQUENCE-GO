package quantum

import (
	"src/github.com/pkg/profile"
	"testing"
)

func Test_randGraph(t *testing.T) {
	//RandGraph(2, "../../tools/1.json", false)
	defer profile.Start().Stop()
	type args struct {
		threadNum int
		path      string
		optimized bool
	}

	tests := []struct {
		name string
		args args
	}{
		{"2 threads seed 1", args{2, "../../tools/1.json", false}},
		{"2 threads optimized seed 1", args{2, "../../tools/1.json", true}},
		{"4 threads  seed 1", args{4, "../../tools/1.json", false}},
		{"4 threads optimized seed 1", args{4, "../../tools/1.json", true}},
		{"8 threads seed 1", args{8, "../../tools/1.json", false}},
		{"8 threads optimized seed 1", args{8, "../../tools/1.json", true}},
		{"2 threads seed 2", args{2, "../../tools/2.json", false}},
		{"2 threads optimized seed 2", args{2, "../../tools/2.json", true}},
		{"4 threads  seed 2", args{4, "../../tools/2.json", false}},
		{"4 threads optimized seed 2", args{4, "../../tools/2.json", true}},
		{"8 threads seed 2", args{8, "../../tools/2.json", false}},
		{"8 threads optimized seed 2", args{8, "../../tools/2.json", true}},
		{"2 threads seed 3", args{2, "../../tools/3.json", false}},
		{"2 threads optimized seed 3", args{2, "../../tools/3.json", true}},
		{"4 threads seed 3", args{4, "../../tools/3.json", false}},
		{"4 threads optimized seed 3", args{4, "../../tools/3.json", true}},
		{"8 threads seed 3", args{8, "../../tools/3.json", false}},
		{"8 threads optimized seed 3", args{8, "../../tools/3.json", true}},
	}
	tests = tests[0:1]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RandGraph(tt.args.threadNum, tt.args.path, tt.args.optimized)
		})
	}
}
