package rccache

import (
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
)

//New initiate the TxDB struct
func (t *CrossShardDec) New(a *basic.Transaction) {
	newT := *a
	t.Data = &newT
	t.Res = 0
	t.Visible = true
	t.InCheck = make([]int, gVar.ShardCnt)
	tmp := make([]bool, gVar.ShardCnt)
	t.InCheckSum = 0
	for i := uint32(0); i < a.TxoutCnt; i++ {
		xx := a.Out[i].ShardIndex()
		//fmt.Println("Out ", i, " ", xx)
		tmp[xx] = true
		t.Value += a.Out[i].Value
		t.InCheck[xx] = -1
	}
	for i := uint32(0); i < a.TxinCnt; i++ {
		xx := a.In[i].ShardIndex()
		//fmt.Println("In ", i, " ", xx)
		tmp[xx] = true
		t.InCheck[xx] = 3
	}

	t.ShardRelated = make([]uint32, 0, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if tmp[i] {
			t.ShardRelated = append(t.ShardRelated, i)
		}
		if t.InCheck[i] == 3 {
			t.InCheckSum++
		}
	}
	t.Total = t.InCheckSum
}

//Update from the transaction
func (t *CrossShardDec) Update(a *basic.Transaction) {
	newT := *a
	t.Data = &newT
	tmp := make([]int, gVar.ShardCnt)
	t.InCheckSum = 0

	for i := uint32(0); i < a.TxoutCnt; i++ {
		tmp[a.Out[i].ShardIndex()] = 1
		t.Value += a.Out[i].Value
	}
	for i := uint32(0); i < a.TxinCnt; i++ {
		xx := a.In[i].ShardIndex()
		tmp[xx] = 2
		if t.InCheck[xx] == 0 {
			t.InCheck[xx] = 3
		}
	}

	t.ShardRelated = make([]uint32, 0, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if tmp[i] == 0 {
			continue
		}
		t.ShardRelated = append(t.ShardRelated, i)
		if tmp[i] == 1 {
			t.InCheck[i] = -1
		} else {
			if t.InCheck[i] > 1 {
				if t.InCheck[i] == 2 {
					t.Res = -1
				}
				t.InCheckSum++
			}
		}
	}
	t.Total = t.InCheckSum
	if t.InCheckSum == 0 {
		t.Res = 1
	}
}

//NewFromOther initiate the TxDB struct by cross-shard data
func (t *CrossShardDec) NewFromOther(index uint32, res bool) {
	t.Data = nil
	t.Visible = true
	t.Res = 0
	t.InCheck = make([]int, gVar.ShardCnt)

	t.Total = int(gVar.ShardCnt - 1)
	if res {
		t.InCheck[index] = 1
		t.InCheckSum = int(gVar.ShardCnt - 1)
	} else {
		t.Res = -1
		t.InCheck[index] = 2
		t.InCheckSum = int(gVar.ShardCnt)
	}
}

//UpdateFromOther initiate the TxDB struct by cross-shard data
func (t *CrossShardDec) UpdateFromOther(index uint32, res bool) {

	if res {
		if t.InCheck[index] == 3 || (t.InCheck[index] == 0 && t.Data == nil) {
			t.InCheck[index] = 1
			t.InCheckSum--
			if t.InCheckSum == 0 {
				t.Res = 1
			}
			t.Total--
		}
	} else {
		if t.InCheck[index] == 3 || (t.InCheck[index] == 0 && t.Data == nil) {
			t.InCheck[index] = 2
			t.Total--
		}
		t.Res = -1
	}

}
