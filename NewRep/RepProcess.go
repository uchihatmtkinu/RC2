package newrep

import (
	"math"

	"github.com/uchihatmtkinu/RC/gVar"
)

//RepMsgGen builds up a RegMsg data
func RepMsgGen(id uint32) {

}

//RepBlockMake generates a new reputation block
func RepBlockMake(data []RepMsg, input []RepSecMsg, prevHash [32]byte) *RepBlock {
	tmp := new(RepBlock)
	tmp.PrevHash = prevHash
	tmp.Round = data[0].Round
	var repvote [gVar.ShardSize]float64
	weight := make([]float64, gVar.ShardSize)
	tmp.TBHash = nil
	TBCache := make([][32]byte, 0, len(data[0].TBHash))
	var mapCache map[[32]byte]uint32
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i].TBHash); j++ {
			tmpHash := data[i].TBHash[j]
			test, ok := mapCache[tmpHash]
			if !ok {
				mapCache[tmpHash] = 1
				TBCache = append(TBCache, tmpHash)
			} else {
				mapCache[tmpHash] = test + 1
			}
		}
	}
	for i := 0; i < len(TBCache); i++ {
		if mapCache[TBCache[i]] > gVar.ShardSize/2 {
			tmp.TBHash = append(tmp.TBHash, TBCache[i])
		}
	}
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i].Vote); j++ {
			repvote[data[i].Vote[j].ID] += float64(data[i].Vote[j].Rep)
		}
	}
	for i := 0; i < len(repvote); i++ {
		repvote[i] = repvote[i] / float64(len(data))
	}
	totalweight := float64(0)
	for i := 0; i < len(data); i++ {
		tmp := float64(0)
		for j := 0; j < len(repvote); j++ {
			tmp = tmp + (repvote[j]-float64(data[i].Vote[j].Rep))*(repvote[j]-float64(data[i].Vote[j].Rep))
		}
		tmp = math.Sqrt(tmp / float64(len(repvote)))
		weight[data[i].ID] = tmp
		totalweight = totalweight + weight[i]
	}
	totalweight2 := float64(0)
	for i := 0; i < len(data); i++ {
		weight[data[i].ID] = totalweight - weight[data[i].ID]
		totalweight2 = totalweight2 + weight[data[i].ID]
	}
	for i := 0; i < len(data); i++ {
		weight[data[i].ID] = weight[data[i].ID] / totalweight2
	}
	tmp.Rep = make([]NewRep, gVar.ShardSize)
	for i := 0; i < len(repvote); i++ {
		tmp.Rep[i].ID = uint32(i)
		tmp.Rep[i].Rep = 0
		tmpRep := float64(0)
		for j := 0; j < len(data); j++ {
			tmpRep += weight[data[j].ID] * float64(data[j].Vote[i].Rep)
		}
		tmp.Rep[i].Rep = uint64(tmpRep)
	}
	tmp.Hash = tmp.GetHash()
	return tmp
}
