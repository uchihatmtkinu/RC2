package shard

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestSortRep(t *testing.T) {
	var a []sortType
	c := 400000
	for i := 0; i < c; i++ {
		var tmp sortType
		tmp.Address = strconv.Itoa(i)
		tmp.Rep = int64(rand.Int() & 100)
		a = append(a, tmp)
	}
	b := a[:]
	SortRep(&b, 0, len(b)-1)
	for i := 0; i < len(b)-1; i++ {
		if b[i].Rep < b[i+1].Rep {
			t.Error(`Sort error`, b[i].Rep, b[i+1].Rep)
		}
	}
}
