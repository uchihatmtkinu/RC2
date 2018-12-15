package basic

import (
	"crypto/sha256"
	"math/big"
	"strconv"
)

//FindByte32 gives a [32]byte data with the input int
func FindByte32(a int) [32]byte {
	tmp := strconv.FormatInt(int64(a), 10)
	return sha256.Sum256([]byte(tmp))
}

//FindBigInt gives a bigInt data with the input int
func FindBigInt(a int) big.Int {
	tmp := strconv.FormatInt(int64(a), 10)
	var xxx = new(big.Int)
	xxx.SetString(tmp, 10)
	return *xxx
}
