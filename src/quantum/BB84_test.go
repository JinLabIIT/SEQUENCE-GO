package quantum

import (
	"fmt"
	"src/github.com/pkg/profile"
	"testing"
)

/*func Test_matMul(t *testing.T) {
	fmt.Println("matrix multiple test")
	a := Basis{{1,2,3},{4,5,6},{7,8,9}}
	b := Basis{{4,5,6},{7,8,9},{10,11,12}}
	c := matMul(&a,&b)
	fmt.Println(*c)
	fmt.Println("test2")
	a = Basis{{1,2,3},{4,5,6}}
	c = matMul(&a,&b)
	fmt.Println(*c)
}*/

/*func Test_outer(t *testing.T) {
	fmt.Println("Outer test")
	a := []complex128{7,8,9}
	b := []complex128{2,5}
	c := outer(a,b)
	fmt.Println(*c)
	b = []complex128{complex(2,3),complex(5,5)}
	c = outer(b,b)
	fmt.Println(*c)
}*/

/*func Test_Kron(t *testing.T) {
	fmt.Println("Kron Test")
	a := Basis{{1,2},{3,4}}
	b := Basis{{0,5},{6,7}}
	c := kron(&a,&b)
	fmt.Println(*c)

	a = Basis{{1,-4,7},{-2,3,3}}
	b = Basis{{8,-9,-6,5},{1,-3,-4,7},{2,8,-8,-3},{1,2,-5,-1}}
	c = kron(&a,&b)
	fmt.Println(*c)
}*/

/*func Test_Transpose(t *testing.T){
	fmt.Println("Transpose Test")
	a := Basis{{1,2},{3,4},{5,6}}
	fmt.Println(*a.transpose())
}
*/

/*func Test_conj(t *testing.T){
	fmt.Println("Conjuate Test")
	a := Basis{{complex(2,3),complex(4,5)},{complex(6,7),complex(8,9)},{complex(1,9),complex(2,8)}}
	fmt.Println(*a.conj())
}
*/
/*func Test_Divided(t *testing.T){
	fmt.Println(math.Sqrt(0.5))
	fmt.Println(-complex(math.Sqrt(0.5),0))
	//panic("stop here")
}*/

func Test(t *testing.T) {
	fmt.Println("start test")
	//test()
	fmt.Println("test finished")
}

func Test2(t *testing.T) {
	fmt.Println("start test2")
	//test2()
	fmt.Println("test2 finished")
}

func Test3(t *testing.T) {
	defer profile.Start().Stop()
	Main(64, 4, 50000000)
}

/*func Test123(t *testing.T){
	go testxxx()
	go testxxx()
	go testxxx()
	testxxx()
	println("Done")
}*/
