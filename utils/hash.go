package utils

import (
	"errors"
)

const MOD int64 = 1e9 + 7
const BASE int64 = 67

var revmap = [128]int64{
	0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
	1,2,3,4,5,6,7,8,9,10,0,0,0,0,0,0,0,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,
	33,34,35,36,0,0,0,0,0,0,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,
	0,0,0,0,0,
}


func HashUID(s *string) (int64, error) {
	var h int64
	h = 0
	for _, c := range *s {
		ind := int(c)

		if ind >= 128 {
			return 0, errors.New("Invalid Character")
		}
		id := revmap[ind]

		h = (h * BASE + id) % MOD
	}

	return h, nil
}
