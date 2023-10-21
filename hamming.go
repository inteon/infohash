package infohash

func calculateHammingCode(fields []uint32) []uint32 {
	log2NumberOfFieldsPlusOne := log2OfXPlusOne(uint32(len(fields)))
	parityCodes := make([]uint32, log2NumberOfFieldsPlusOne)

	for id, field := range fields {
		for i := range parityCodes {
			if (id+1)&(1<<i) != 0 {
				parityCodes[i] ^= field
			}
		}
	}

	return parityCodes
}

func findErrorLocation(fields []uint32, hammingCode []uint32) (bool, uint32) {
	expectedCode := calculateHammingCode(fields)
	var errorLocation uint32

	for i := range hammingCode {
		if hammingCode[i] != expectedCode[i] {
			errorLocation |= 1 << i
		}
	}

	if errorLocation == 0 || errorLocation > uint32(len(fields)) {
		return false, 0
	}

	return true, errorLocation - 1
}

func log2OfXPlusOne(x uint32) uint32 {
	var r uint32
	for x > 0 {
		x >>= 1
		r += 1
	}
	return r
}
