package rccache

import (
	"math/rand"

	"github.com/uchihatmtkinu/RC/gVar"
)

//GetGossipID return the next ID of the gossip
func (a *DbRef) GetGossipID(kind int, tmparr []uint32) uint32 {
	can := make([]uint32, 0, gVar.ShardSize)
	if kind == 0 {
		least := a.GossipPool[0]
		for i := uint32(0); i < gVar.ShardSize; i++ {
			if a.GossipPool[i] < least {
				least = a.GossipPool[i]
			}
		}
		for i := uint32(0); i < gVar.ShardSize; i++ {
			if a.GossipPool[i] == least {
				can = append(can, i)
			}
		}
	} else {
		least := a.GossipPool[tmparr[0]]
		for i := 0; i < len(tmparr); i++ {
			if a.GossipPool[tmparr[i]] < least {
				least = a.GossipPool[tmparr[i]]
			}
		}
		for i := 0; i < len(tmparr); i++ {
			if a.GossipPool[tmparr[i]] == least {
				can = append(can, tmparr[i])
			}
		}
	}
	ran := uint32(rand.Intn(len(can)))
	a.GossipPool[can[ran]]++
	return can[ran]
}
