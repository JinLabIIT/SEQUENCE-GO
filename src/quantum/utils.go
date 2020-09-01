package quantum

import (
	rng "github.com/leesper/go_rng"
	"math/cmplx"
	"strconv"
	"strings"
)

//type Basis [][]complex128

// help functions
func exists(slice []*Node, val *Node) bool {
	for _, item := range slice {
		if item == val { // question mark
			return true
		}
	}
	return false
}

func multiply(base []float64, state []complex128) []complex128 { // 2*2 matrix * 2*2 matrix
	a := complex(base[0], 0) * state[0]
	b := complex(base[1], 0) * state[1]
	return []complex128{a, b}
}

func makeArray(length int, value int) []int {
	results := make([]int, length)
	for i := 0; i < length; i++ {
		results[i] = value
	}
	return results
}

func outer(a, b *[]complex128) *[][]complex128 { // assume a and b are m*1 and 1*n matrix
	result := make([][]complex128, len(*a))
	for i, c := range *a {
		for _, d := range *b {
			result[i] = append(result[i], c*d)
		}
	}
	return &result
}

func kron(a, b *[][]complex128) *[][]complex128 { // a->m*n b->i*j
	rowA := len(*a)
	rowB := len(*b)
	colA := len((*a)[0])
	colB := len((*b)[0])
	result := make([][]complex128, rowA*rowB)
	for m := 0; m < rowA; m++ {
		for n := 0; n < colA; n++ {
			for i := 0; i < rowB; i++ {
				for j := 0; j < colB; j++ {
					result[m*rowB+i] = append(result[m*rowB+i], (*a)[m][n]*(*b)[i][j])
				}
			}
		}
	}
	return &result
}

func transpose(basis *[][]complex128) *[][]complex128 {
	m := len(*basis)
	n := len((*basis)[0])
	result := make([][]complex128, n) //n = 1,2
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			result[j] = append(result[j], (*basis)[i][j])
		}
	}
	return &result
}

func conj(basis *[][]complex128) *[][]complex128 {
	result := make([][]complex128, len(*basis))
	for i := 0; i < len(*basis); i++ {
		for j := 0; j < len((*basis)[0]); j++ {
			result[i] = append(result[i], cmplx.Conj((*basis)[i][j]))
		}
	}
	return &result
}

func matMul(a, b *[][]complex128) *[][]complex128 { // Matrix multiplication a->m*n b->n*p
	m := len(*a)
	n := len((*a)[0])
	p := len((*b)[0])
	if n != len(*b) {
		panic("the columns of first matrix must equal to the rows of the second matrix")
	}
	result := make([][]complex128, m) // m = 1,2
	for i := 0; i < m; i++ {
		for j := 0; j < p; j++ {
			val := helpMatMul(a, b, i, j) //a[i][]*b[][j]
			result[i] = append(result[i], val)
		}
	}
	return &result
}

func helpMatMul(a, b *[][]complex128, aIndex int, bIndex int) complex128 { // a[i][] * b[][j]
	var result complex128
	for i := 0; i < len((*a)[0]); i++ {
		result += (*a)[aIndex][i] * (*b)[i][bIndex]
	}
	return result
}

func oneToTwo(a *[]complex128) *[][]complex128 { //one dimension to two dimension array
	result := make([][]complex128, 2)
	for i := 0; i < len(*a); i++ {
		result[i] = []complex128{(*a)[i]}
	}
	return &result
}

func divide(a *[][]complex128, b float64) []complex128 { //1*n matrix divided by float
	if b == 0 {
		panic("can not divided by ZERO")
	}
	result := make([]complex128, 2)
	result[0] = (*a)[0][0] / complex(b, 0)
	result[1] = (*a)[1][0] / complex(b, 0)
	return result
}

func arrayConj(arr *[]complex128) *[]complex128 {
	for i, ele := range *arr {
		(*arr)[i] = cmplx.Conj(ele)
	}
	return arr
}

func choice(array []int, n int, rng *rng.UniformGenerator) []int {

	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = array[int(rng.Int32n(int32(len(array))))]
	}
	return result
}

func sliceToInt(slice []int, index int) uint64 { // convert from binary list to int
	var results string
	for i := range slice {
		results += strconv.Itoa(slice[i])
	}
	val, _ := strconv.ParseUint(results, index, 64)
	return val
}

func toString(a []int) string {
	valuesText := make([]string, len(a))
	for i := 0; i < len(a); i++ {
		valuesText[i] = strconv.Itoa(a[i])
	}
	result := strings.Join(valuesText, " ")
	return result
}
