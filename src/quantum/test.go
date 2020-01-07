// Copyright 2009 The GoMatrix Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//target:gomatrix.googlecode.com/hg/matrix

//Linear algebra.
package quantum

//The MatrixRO interface defines matrix operations that do not change the
//underlying data, such as information requests or the creation of transforms
/*
Read-only matrix types (at the moment, PivotMatrix).
*/

type MatrixRO interface {

	//The number of rows in this matrix.
	Rows() int
	//The number of columns in this matrix.
	Cols() int

	//The number of elements in this matrix.
	NumElements() int
	//The size pair, (Rows(), Cols())
	GetSize() (int, int)

	//The element in the ith row and jth column.
	Get(i, j int) complex128

	//Plus(MatrixRO) (Matrix, error)
	//Minus(MatrixRO) (Matrix, error)
	Times(MatrixRO) Matrix

	DenseMatrix() *DenseMatrix
	//SparseMatrix() *SparseMatrix
}

type Matrix interface {
	MatrixRO

	//Set the element at the ith row and jth column to v.
	Set(i int, j int, v complex128)

	Add(MatrixRO) error
	Subtract(MatrixRO) error
	Scale(complex128)
}

type matrix struct {
	rows int
	cols int
}

func (A *matrix) Nil() bool { return A == nil }

func (A *matrix) Rows() int { return A.rows }

func (A *matrix) Cols() int { return A.cols }

func (A *matrix) NumElements() int { return A.rows * A.cols }

func (A *matrix) GetSize() (rows, cols int) {
	rows = A.rows
	cols = A.cols
	return
}

func Kronecker(A, B MatrixRO) (C *DenseMatrix) {
	ars, acs := A.Rows(), A.Cols()
	brs, bcs := B.Rows(), B.Cols()
	C = Zeros(ars*brs, acs*bcs)
	for i := 0; i < ars; i++ {
		for j := 0; j < acs; j++ {
			Cij := C.GetMatrix(i*brs, j*bcs, brs, bcs)
			Cij.SetMatrix(0, 0, Scaled(B, A.Get(i, j)))
		}
	}
	return
}

func Scaled(A MatrixRO, f complex128) (B *DenseMatrix) {
	B = MakeDenseCopy(A)
	B.Scale(f)
	return
}

func MakeDenseCopy(A MatrixRO) *DenseMatrix {
	B := Zeros(A.Rows(), A.Cols())
	for i := 0; i < B.rows; i++ {
		for j := 0; j < B.cols; j++ {
			B.Set(i, j, A.Get(i, j))
		}
	}
	return B
}

type DenseMatrix struct {
	matrix
	// flattened matrix data. elements[i*step+j] is row i, col j
	elements []complex128
	// actual offset between rows
	step int
}

/*
Returns an array of slices referencing the matrix data. Changes to
the slices effect changes to the matrix.
*/
func (A *DenseMatrix) Arrays() [][]complex128 {
	a := make([][]complex128, A.rows)
	for i := 0; i < A.rows; i++ {
		a[i] = A.elements[i*A.step : i*A.step+A.cols]
	}
	return a
}

/*
Returns the contents of this matrix stored into a flat array (row-major).
*/
func (A *DenseMatrix) Array() []complex128 {
	if A.step == A.rows {
		return A.elements[0 : A.rows*A.cols]
	}
	a := make([]complex128, A.rows*A.cols)
	for i := 0; i < A.rows; i++ {
		for j := 0; j < A.cols; j++ {
			a[i*A.cols+j] = A.elements[i*A.step+j]
		}
	}
	return a
}

/*
Get the element in the ith row and jth column.
*/
func (A *DenseMatrix) Get(i int, j int) (v complex128) {
	v = A.elements[i*A.step : i*A.step+A.cols][j]
	return v
}

/*
Set the element in the ith row and jth column to v.
*/
func (A *DenseMatrix) Set(i int, j int, v complex128) {
	A.elements[i*A.step : i*A.step+A.cols][j] = v
}

func Zeros(rows, cols int) *DenseMatrix {
	A := new(DenseMatrix)
	A.elements = make([]complex128, rows*cols)
	A.rows = rows
	A.cols = cols
	A.step = cols
	return A
}

func (A *DenseMatrix) GetMatrix(i, j, rows, cols int) *DenseMatrix {
	B := new(DenseMatrix)
	B.elements = A.elements[i*A.step+j : i*A.step+j+(rows-1)*A.step+cols]
	B.rows = rows
	B.cols = cols
	B.step = A.step
	return B
}

func (A *DenseMatrix) SetMatrix(i, j int, B *DenseMatrix) {
	for r := 0; r < B.rows; r++ {
		for c := 0; c < B.cols; c++ {
			A.Set(i+r, j+c, B.Get(r, c))
		}
	}
}

func (A *DenseMatrix) Scale(f complex128) {
	for i := 0; i < A.rows; i++ {
		index := i * A.step
		for j := 0; j < A.cols; j++ {
			A.elements[index] *= f
			index++
		}
	}
}

func (A *DenseMatrix) Transpose() *DenseMatrix {
	B := Zeros(A.Cols(), A.Rows())
	for i := 0; i < A.Rows(); i++ {
		for j := 0; j < A.Cols(); j++ {
			B.Set(j, i, A.Get(i, j))
		}
	}
	return B
}

func (A *DenseMatrix) Times(B MatrixRO) *DenseMatrix {
	C := Zeros(A.rows, B.Cols())

	for i := 0; i < A.rows; i++ {
		for j := 0; j < B.Cols(); j++ {
			sum := complex128(0)
			for k := 0; k < A.cols; k++ {
				sum += A.elements[i*A.step+k] * B.Get(k, j)
			}
			C.elements[i*C.step+j] = sum
		}
	}
	return C
}

func MakeDenseMatrix(elements []complex128, rows, cols int) *DenseMatrix {
	A := new(DenseMatrix)
	A.rows = rows
	A.cols = cols
	A.step = cols
	A.elements = elements
	return A
}

func MakeDenseMatrixStacked(data [][]complex128) *DenseMatrix {
	rows := len(data)
	cols := len(data[0])
	elements := make([]complex128, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			elements[i*cols+j] = data[i][j]
		}
	}
	return MakeDenseMatrix(elements, rows, cols)
}