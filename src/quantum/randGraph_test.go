package quantum

import (
	"golang.org/x/exp/errors/fmt"
	"testing"
)

func Test_randGraph(t *testing.T) {
	type args struct {
		threadNum int
		path      string
		optimized bool
	}

	type test struct {
		name string
		args args
	}
	tests := []test{}

	for n := 48; n <= 80; n += 8 {
		for d := 1.5; d <= 2.5; d += 0.5 {
			for seed := 0; seed <= 2; seed++ {
				filename := fmt.Sprintf("../../tools/%d_%.1f_%d.json", n, d, seed)
				for threadNum := 2; threadNum <= 32; threadNum *= 2 {
					testName := fmt.Sprintf("%s file; %d threads; random schedule", filename, threadNum)
					arg := args{
						threadNum: threadNum,
						path:      filename,
						optimized: false,
					}
					t := test{
						name: testName,
						args: arg,
					}
					tests = append(tests, t)
					testName = fmt.Sprintf("%s file; %d threads; optimized schedule", filename, threadNum)
					arg = args{
						threadNum: threadNum,
						path:      filename,
						optimized: true,
					}
					t = test{testName, arg}
					tests = append(tests, t)
				}
			}
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.name)
			randGraph(tt.args.threadNum, tt.args.path, tt.args.optimized)
		})
	}
}
