package quantum

import "math"

func polarization() map[string]interface{} {
	bases := [][2][2]complex128{{{complex128(1), complex128(0)}, {complex128(0), complex128(1)}}, {{complex(math.Sqrt(0.5), 0), complex(math.Sqrt(0.5), 0)}, {complex(-math.Sqrt(0.5), 0), complex(math.Sqrt(0.5), 0)}}}
	polarization := map[string]interface{}{
		"name":  "polarization",
		"bases": bases,
	}
	return polarization
}

func timeBin() map[string]interface{} {
	bases := [2][2][2]complex128{{{complex128(1), complex128(0)}, {complex128(0), complex128(1)}}, {{complex(math.Sqrt(1/2), 0), complex(math.Sqrt(1/2), 0)}, {complex(math.Sqrt(1/2), 0), complex(-math.Sqrt(1/2), 0)}}}
	timeBin := map[string]interface{}{
		"name":          "timeBin",
		"bases":         bases,
		"binSeparation": 1400, //measured in ps
	}
	return timeBin
}
