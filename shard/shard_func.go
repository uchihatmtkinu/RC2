package shard

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"strings"

	"fmt"

	"github.com/uchihatmtkinu/RC/gVar"
)

type sortType struct {
	ID      uint32
	Rep     int64
	Address string
}

//Instance is the struct for sharding
type Instance struct {
	rng *rand.Rand
}

//GetRBData get all the data from reputation block for sharding
func GetRBData() {

}

//CompareRep returns whether a has a great reputation than b
func CompareRep(a *sortType, b *sortType) int {
	if a.Rep > b.Rep {
		return 1
	} else if b.Rep > a.Rep {
		return -1
	} else {
		return strings.Compare(a.Address, b.Address)
	}
}

//SortRep sorts all miners based on their reputation
func SortRep(a *[]sortType, l int, r int) error {
	x := (*a)[(l+r)/2]
	i := l
	j := r
	if l >= r {
		return nil
	}
	for i <= j {
		for i < r && CompareRep(&(*a)[i], &x) > 0 {
			i++
		}
		for j > 0 && CompareRep(&x, &(*a)[j]) > 0 {
			j--
		}
		if i <= j {
			y := (*a)[i]
			(*a)[i] = (*a)[j]
			(*a)[j] = y
			i++
			j--
		}
	}
	if i < r {
		SortRep(a, i, r)
	}
	if l < j {
		SortRep(a, l, j)
	}
	return nil
}

//GenerateSeed come out the seed used in random number
func (c *Instance) GenerateSeed(a *[][32]byte) error {
	var tmp []byte
	for i := 0; i < len(*a); i++ {
		tmp = append(tmp, (*a)[i][:]...)
	}
	hash := sha256.Sum256(tmp)
	c.rng = rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(hash[:]))))

	return nil
}

//Sharding do the shards given reputations
func (c *Instance) Sharding(a *[]MemShard, b *[][]int) {
	sortData := make([]sortType, len(*a))
	for i := 0; i < len(*a); i++ {
		(*a)[i].PreShard = (*a)[i].Shard
		sortData[i].Address = (*a)[i].PrivateAddress
		sortData[i].ID = uint32(i)
		sortData[i].Rep = (*a)[i].CalTotalRep()
	}
	SortRep(&sortData, 0, len(*a)-1)
	now := 0
	for i := uint32(0); i < gVar.ShardSize; i++ {
		for j := uint32(0); j < gVar.ShardCnt; j++ {
			(*b)[j][i] = -1
		}
	}
	for i := uint32(0); i < gVar.ShardSize; i++ {
		check := make([]int, gVar.ShardCnt)
		for j := uint32(0); j < gVar.ShardCnt; j++ {
			x := uint32(c.rng.Int()^int((*a)[sortData[now].ID].CalTotalRep())) % gVar.ShardCnt
			if check[x] == 0 {
				check[x] = 1
				(*b)[x][i] = int(sortData[now].ID)
				now++
			} else {
				j--
			}
		}
	}

	//select leader, index of 0 is the leader
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		c.LeaderSort(a, b, i)
	}

	//set shardid, shard, role for all the members
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		for j := uint32(0); j < gVar.ShardSize; j++ {
			(*a)[(*b)[i][j]].InShardId = int(j)
			(*a)[(*b)[i][j]].Shard = int(i)
			//(*a)[(*b)[i][j]].ClearRep()
			if j == 0 {
				//leader
				(*a)[(*b)[i][j]].Role = 0
			} else {
				//miner
				(*a)[(*b)[i][j]].Role = 1
			}
		}
	}
	fmt.Println(b)
}

//LeaderSort give the priority of being leader in this round
func (c *Instance) LeaderSort(a *[]MemShard, b *[][]int, xx uint32) {
	tmp := make([]float32, len((*b)[xx]))
	for i := 0; i < len(tmp); i++ {
		//TODO nagative rep
		if i < len(tmp)/2 {
			tmp[i] = c.rng.Float32() / float32((*a)[(*b)[xx][i]].CalTotalRep()+1)
		} else {
			tmp[i] = float32((*a)[(*b)[xx][0]].CalTotalRep()) / float32((*a)[(*b)[xx][i]].CalTotalRep())
		}
	}
	for i := 0; i < len(tmp); i++ {
		for j := i + 1; j < len(tmp); j++ {
			if tmp[i] > tmp[j] {
				y := tmp[i]
				tmp[i] = tmp[j]
				tmp[j] = y
				x := (*b)[xx][i]
				(*b)[xx][i] = (*b)[xx][j]
				(*b)[xx][j] = x
			}
		}
	}
}

//ReturnRepData gets the rep data of a specific shard
func ReturnRepData(shardID uint32) *[]int64 {
	tmp := make([]int64, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		tmp[i] = GlobalGroupMems[ShardToGlobal[shardID][i]].Rep
	}
	return &tmp
}
