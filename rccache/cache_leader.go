package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC2/basic"
	"github.com/uchihatmtkinu/RC2/gVar"
	"github.com/uchihatmtkinu/RC2/shard"
)

//MakeTXList is to create TxList given transaction
func (d *DbRef) MakeTXList(b *basic.Transaction) error {
	/*err := d.AddCache(b.Hash)
	if err != nil {
		return err
	}*/
	//tmpPre, okw := d.WaitHashCache[basic.HashCut(b.Hash)]
	/*if okw {
		for i := 0; i < len(tmpPre.DataTB); i++ {
			tmpPre.StatTB[i].Valid[tmpPre.IDTB[i]] = 1
			tmpPre.StatTB[i].Stat--
			tmpPre.DataTB[i].TxArray[tmpPre.IDTB[i]] = *b
			SendingChan(&tmpPre.StatTB[i].Channel)
		}
		for i := 0; i < len(tmpPre.DataTDS); i++ {
			tmpPre.StatTDS[i].Valid[tmpPre.IDTDS[i]] = 1
			tmpPre.StatTDS[i].Stat--
			tmpPre.DataTDS[i].TxArray[tmpPre.IDTDS[i]] = b.Hash
			SendingChan(&tmpPre.StatTDS[i].Channel)
		}
		for i := 0; i < len(tmpPre.DataTL); i++ {
			tmpPre.StatTL[i].Valid[tmpPre.IDTL[i]] = 1
			tmpPre.StatTL[i].Stat--
			tmpPre.DataTL[i].TxArray[tmpPre.IDTL[i]] = b.Hash
			SendingChan(&tmpPre.StatTL[i].Channel)
		}
		//fmt.Println(time.Now(), "Sent Chan signal", base58.Encode(b.Hash[:]))
	}*/
	//delete(d.WaitHashCache, basic.HashCut(b.Hash))
	/*tmp, ok := d.TXCache[b.Hash]
	if !ok {
		tmp = new(CrossShardDec)
		tmp.New(b)
	} else {
		tmp.Update(b)
	}
	if tmp.InCheck[d.ShardNum] == 0 {
		if ok {
			delete(d.HashCache, basic.HashCut(b.Hash))
			delete(d.TXCache, b.Hash)
		}
		return fmt.Errorf("Not related TX")
	}
	if tmp.Res == 1 {
		d.Ready = append(d.Ready, *tmp.Data)
	}
	//d.DB.AddTx(b)

	d.TXCache[b.Hash] = tmp*/
	if d.Now == nil {
		d.NewTxList()
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.Now.TLS[i].AddTx(b)
	}
	/*if tmp.InCheck[d.ShardNum] != -1 {
		//fmt.Println("Band:", d.BandCnt)
		if gVar.BandDiverse {
			d.BandCnt += uint32(shard.GlobalGroupMems[d.ID].Bandwidth)
			//fmt.Println("Band New:", d.BandCnt)
			if d.BandCnt >= gVar.MaxBand {
				//fmt.Println("Add new tx")
				for i := uint32(0); i < gVar.ShardCnt; i++ {
					if tmp.InCheck[i] != 0 {
						d.Now.TLS[i].AddTx(b)
					}
				}
				d.BandCnt -= gVar.MaxBand
			}
		} else {
			for i := uint32(0); i < gVar.ShardCnt; i++ {
				if tmp.InCheck[i] != 0 {
					d.Now.TLS[i].AddTx(b)
				}
			}
		}
	}*/
	return nil
}

//BuildTDS is to build all txDecSet
//Must after SignTXL
func (d *DbRef) BuildTDS() {
	d.Now.TLS[d.ShardNum].Sign(&d.prk)
	d.TxCnt += d.Now.TLS[d.ShardNum].TxCnt
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if d.ShardNum != i {
			d.Now.TLS[i].HashID = d.Now.TLS[i].Hash()
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if i == d.ShardNum {
			d.Now.TDS[i].Set(&d.Now.TLS[i], d.ShardNum, 1)
		} else {
			d.Now.TDS[i].Set(&d.Now.TLS[i], d.ShardNum, 0)
		}
	}

}

//SignTDS is to sign all txDecSet
func (d *DbRef) SignTDS(x *TLGroup) {
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		x.TDS[i].Sign(&d.prk)
	}

}

//NewTxList initialize the txList
//Must after BuildTDS
func (d *DbRef) NewTxList() error {
	if d.Now != nil {
		d.TLIndex[d.Now.TLS[d.ShardNum].Hash()] = d.Now
		d.TLRound++
	}
	d.BandCnt = 0
	d.Now = new(TLGroup)
	for i := 0; i < len(d.Ready); i++ {
		d.Now.TLS[d.ShardNum].AddTx(&d.Ready[i])
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.Now.TLS[i].ID = d.ID
		d.Now.TLS[i].Round = d.TLRound
	}
	return nil
}

//GenerateTxBlock makes the TxBlock
func (d *DbRef) GenerateTxBlock(good int) error {
	height := d.TxB.Height
	if good > 0 {
		if good == 1 {
			d.TxB = new(basic.TxBlock)
			d.TxB.MakeTxBlock(d.ID, &d.Ready, d.DB.LastTB, &d.prk, height+1, 0, nil, 0)
		}
		d.TxB.Kind = 0
		/*for i := 0; i < len(d.Ready); i++ {
			d.ClearCache(d.Ready[i].Hash)
		}*/
		//d.Ready = nil
		*(d.TBCache) = append(*(d.TBCache), d.TxB.HashID)

		d.DB.AddBlock(d.TxB)
		//d.DB.UpdateUTXO(d.TxB, d.ShardNum)
	} else {
		d.TxB = new(basic.TxBlock)
		d.TxB.MakeTxBlock(d.ID, &d.Ready, d.DB.LastTB, &d.prk, height+1, 3, nil, 0)
	}
	//d.DB.ShowAccount()

	return nil
}

//GenerateFinalBlock generate final block
func (d *DbRef) GenerateFinalBlock() error {
	tmp := d.DB.MakeFinalTx(d.ShardNum)
	height := d.TxB.Height
	d.TxB = new(basic.TxBlock)
	d.TxB.MakeTxBlock(d.ID, tmp, d.DB.LastFB[d.ShardNum], &d.prk, height+1, 1, &d.FB[d.ShardNum].HashID, d.ShardNum)
	//*(d.TBCache) = append(*(d.TBCache), d.TxB.HashID)
	d.FB[d.ShardNum] = d.TxB
	err := d.DB.AddFinalBlock(d.TxB)
	if err != nil {
		fmt.Println(err)
	}
	xxx := d.TxB.Serial()
	//fmt.Println("FB length: ", len(xxx))
	xxxtmp := new(basic.TxBlock)
	err = xxxtmp.Decode(&xxx, 1)
	//fmt.Println(err)
	return nil
}

//GenerateStartBlock generate Start block
func (d *DbRef) GenerateStartBlock() error {
	height := d.FB[d.ShardNum].Height
	tmp := make([][32]byte, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmp[i] = d.FB[i].HashID
	}
	d.TxB = new(basic.TxBlock)
	d.TxB.MakeStartBlock(d.ID, &tmp, d.DB.LastFB[d.ShardNum], &d.prk, height+1)
	d.DB.AddStartBlock(d.TxB)
	return nil
}

//UpdateTXCache is to pick the transactions into ready slice given txdecision
func (d *DbRef) UpdateTXCache(a *basic.TxDecision, index *uint32) error {
	if a.Single == 1 {
		return fmt.Errorf("TxDecision parameter error")
	}
	tmp, ok := d.TLIndex[a.HashID]
	if !ok {
		index = nil
		return fmt.Errorf("TxDecision Hash error, wrong or time out")
	}
	*index = tmp.TLS[d.ShardNum].Round - d.PrevHeight
	var x, y uint32 = 0, 0
	tmpTD := make([]basic.TxDecision, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {

		tmpTD[i].Set(a.ID, i, 1)
		tmpTD[i].HashID = tmp.TLS[i].HashID
		tmpTD[i].Sig = nil
		tmpTD[i].Sig = append(tmpTD[i].Sig, a.Sig[i])
	}
	//fmt.Println("Leader ", d.ID, " process TxDecision: ")
	for i := uint32(0); i < tmp.TLS[d.ShardNum].TxCnt; i++ {
		/*tmpTx, ok := d.TXCache[tmp.TLS[d.ShardNum].TxArray[i]]
		if !ok {
			fmt.Println("Not related tx?")
		}*/
		//tmpTx.Print()
		for j := uint32(0); j < gVar.ShardCnt; j++ {
			tmpTD[j].Add((a.Decision[x] >> y) & 1)
		}
		if y < 7 {
			y++
		} else {
			x++
			y = 0
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		//fmt.Println("Decision: ", tmpTD[i].Decision, "Sig: ", tmpTD[i].Sig[0])
		if !tmpTD[i].Verify(&shard.GlobalGroupMems[a.ID].RealAccount.Puk, 0) {
			return fmt.Errorf("Signature not match %d", i)
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmp.TDS[i].Add(&tmpTD[i])
	}
	return nil
}

//ProcessTDS deal with the TDS
func (d *DbRef) ProcessTDS(b *basic.TxDecSet, res *[gVar.ShardSize]int64) {
	if b.ShardIndex == d.ShardNum {
		//tmp1, _ := d.TLIndex[b.HashID]
		//tmpIndex := tmp - uint32(d.StartIndex)
		//tmpTL := (*d.TLSCache[tmpIndex])[d.ShardNum]
		//b.TxCnt = tmp1.TLS[d.ShardNum].TxCnt
		//b.TxArray = tmp1.TLS[d.ShardNum].TxArray
		x := shard.ReturnRepData(0)
		for i := 0; i < len(*x); i++ {
			fmt.Print((*x)[i], " ")
		}
		fmt.Println()
	}
}

//Release delete the first element of the cache
func (d *DbRef) Release(x *TLGroup) {
	//d.TLCache = d.TLCache[1:]
	hash := x.TLS[d.ShardNum].HashID
	delete(d.TLIndex, hash)
}
