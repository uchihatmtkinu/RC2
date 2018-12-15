package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/shard"
)

//PreTxList is to process the TxListX
func (d *DbRef) PreTxList(b *basic.TxList, s *PreStat) error {
	if s == nil {
		s = new(PreStat)
		s.Stat = -2
	}

	if s.Stat == -2 {
		if d.Leader != b.ID {
			return fmt.Errorf("PreTxList: Txlist from a miner")
		}
		if int(b.TxCnt) != len(b.TxArrayX) {
			return fmt.Errorf("PreTxList: Number of Tx wrong")
		}
		s.Stat = -1
	}
	/*
		if s.Stat == -1 {
			s.Stat = int(b.TxCnt)
			s.Valid = make([]int, b.TxCnt)
			b.TxArray = make([][32]byte, b.TxCnt)
			for i := uint32(0); i < b.TxCnt; i++ {
				xxx, ok := d.HashCache[b.TxArrayX[i]]
				if !ok {
					tmpWait, okW := d.WaitHashCache[b.TxArrayX[i]]
					if okW {
						tmpWait.DataTL = append(tmpWait.DataTL, b)
						tmpWait.StatTL = append(tmpWait.StatTL, s)
						tmpWait.IDTL = append(tmpWait.IDTL, int(i))
					} else {
						tmpWait = WaitProcess{nil, nil, nil, nil, nil, nil, nil, nil, nil}
						tmpWait.DataTL = append(tmpWait.DataTL, b)
						tmpWait.StatTL = append(tmpWait.StatTL, s)
						tmpWait.IDTL = append(tmpWait.IDTL, int(i))
					}
					d.WaitHashCache[b.TxArrayX[i]] = tmpWait
				} else {
					s.Stat--
					if xxx == nil {
						fmt.Println("xxx[0] is null")
					} else {
						if len(xxx) < 1 {
							fmt.Println("xxx length not enough")
						}
						_, tmpOK := d.TXCache[xxx[0]]
						if !tmpOK {
							fmt.Println("TXcache not ok! hash:", xxx[0])
						} else if d.TXCache[xxx[0]] == nil {
							fmt.Println("TxCache data is null")
						} else if d.TXCache[xxx[0]].Data == nil {
							fmt.Println("Tx Data is null")
						}
					}
					b.TxArray[i] = d.TXCache[xxx[0]].Data.Hash
					s.Valid[i] = 1
				}
			}
		}
		if s.Stat > 0 {
			s.Channel = make(chan bool, s.Stat)
		}*/
	s.Stat = 0
	if s.Stat == 0 {
		if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk) {
			return fmt.Errorf("PreTxList: Signature not match")
		}
	}
	return nil
}

//PreTxDecision is verify the txdecisi
func (d *DbRef) PreTxDecision(b *basic.TxDecision, hash [32]byte) error {
	if int(b.TxCnt) > len(b.Decision)*8 {
		return fmt.Errorf("PreTxDecision: Decision lengh not enough")
	}
	if b.Single == 0 {
		if b.Target != d.ShardNum {
			return fmt.Errorf("PreTxDecision: TxDecision should be the intra-one")
		}
		if shard.GlobalGroupMems[b.ID].Shard != int(d.ShardNum) {
			return fmt.Errorf("PreTxDecision: Not the same shard")
		}
		if len(b.Sig) != int(gVar.ShardCnt) {
			return fmt.Errorf("PreTxDecision: Signature not enough")
		}
		if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk, d.ShardNum) {
			return fmt.Errorf("PreTxDecision: Signature not match")
		}
	} else {
		b.HashID = hash
		if b.Target != d.ShardNum {
			return fmt.Errorf("PreTxDecision: Not the target shard")
		}
		if len(b.Sig) != 1 {
			return fmt.Errorf("PreTxDecision: Signature not enough")
		}
		if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk, 0) {
			return fmt.Errorf("PreTxDecision: Signature not match")
		}
	}
	return nil
}

//PreTxDecSet is to process the TxListX
func (d *DbRef) PreTxDecSet(b *basic.TxDecSet, s *PreStat) error {
	if s == nil {
		s = new(PreStat)
		s.Stat = -2
	}
	if s.Stat == -2 {
		if shard.GlobalGroupMems[b.ID].Role != 0 {
			fmt.Println("PreTxDecSet: Not a Leader")
			return fmt.Errorf("PreTxDecSet: Not a Leader")
		}
		if int(b.TxCnt) != len(b.TxArrayX) || int(b.MemCnt) != len(b.MemD) {
			fmt.Println("PreTxDecSet: TxDecSet parameter not match")
			return fmt.Errorf("PreTxDecSet: TxDecSet parameter not match")
		}
		s.Stat = -1
	}
	/*
		if s.Stat == -1 {
			s.Stat = int(b.TxCnt)
			s.Valid = make([]int, b.TxCnt)
			b.TxArray = make([][32]byte, b.TxCnt)
			for i := uint32(0); i < b.TxCnt; i++ {
				xxx, ok := d.HashCache[b.TxArrayX[i]]
				if !ok {
					tmpWait, okW := d.WaitHashCache[b.TxArrayX[i]]
					if okW {
						tmpWait.DataTDS = append(tmpWait.DataTDS, b)
						tmpWait.StatTDS = append(tmpWait.StatTDS, s)
						tmpWait.IDTDS = append(tmpWait.IDTDS, int(i))
					} else {
						tmpWait = WaitProcess{nil, nil, nil, nil, nil, nil, nil, nil, nil}
						tmpWait.DataTDS = append(tmpWait.DataTDS, b)
						tmpWait.StatTDS = append(tmpWait.StatTDS, s)
						tmpWait.IDTDS = append(tmpWait.IDTDS, int(i))
					}
					d.WaitHashCache[b.TxArrayX[i]] = tmpWait
				} else {
					s.Stat--
					s.Valid[i] = 1
					if xxx == nil {
						fmt.Println("TDS xxx[0] is null")
					} else {
						if len(xxx) < 1 {
							fmt.Println("TDS xxx length not enough")
						}
						_, tmpOK := d.TXCache[xxx[0]]
						if !tmpOK {
							fmt.Println("TDS TXcache not ok! hash:", base58.Encode(xxx[0][:]))
						} else if d.TXCache[xxx[0]] == nil {
							fmt.Println("TDS TxCache data is null")
						} else if d.TXCache[xxx[0]].Data == nil {
							fmt.Println("TDS Tx Data is null")
						}
					}
					b.TxArray[i] = d.TXCache[xxx[0]].Data.Hash
				}
			}
		}
		if s.Stat > 0 {
			s.Channel = make(chan bool, s.Stat)
		}*/
	s.Stat = 0
	if s.Stat == 0 {
		for i := uint32(0); i < b.MemCnt; i++ {
			err := d.PreTxDecision(&b.MemD[i], b.HashID)
			if err != nil {
				return err
			}
		}
		if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk) {
			return fmt.Errorf("PreTxDecSet: Signature not match")
		}
	}
	return nil
}

//PreTxBlock is to process the TxListX
func (d *DbRef) PreTxBlock(b *basic.TxBlock, s *PreStat) error {
	if s == nil {
		s = new(PreStat)
		s.Stat = -2
	}
	if s.Stat == -2 {
		if shard.GlobalGroupMems[b.ID].Role != 0 {
			return fmt.Errorf("PreTxBlock: Not a Leader")
		}
		if int(b.TxCnt) != len(b.TxArrayX) {
			return fmt.Errorf("PreTxBlock: TxBlock parameter not match")
		}
		s.Stat = -1
	}
	/*if s.Stat == -1 {
		s.Stat = int(b.TxCnt)
		s.Valid = make([]int, b.TxCnt)
		b.TxArray = make([]basic.Transaction, b.TxCnt)
		for i := uint32(0); i < b.TxCnt; i++ {
			xxx, ok := d.HashCache[b.TxArrayX[i]]
			if !ok {
				tmpWait, okW := d.WaitHashCache[b.TxArrayX[i]]
				if okW {
					tmpWait.DataTB = append(tmpWait.DataTB, b)
					tmpWait.StatTB = append(tmpWait.StatTB, s)
					tmpWait.IDTB = append(tmpWait.IDTB, int(i))
				} else {
					tmpWait = WaitProcess{nil, nil, nil, nil, nil, nil, nil, nil, nil}
					tmpWait.DataTB = append(tmpWait.DataTB, b)
					tmpWait.StatTB = append(tmpWait.StatTB, s)
					tmpWait.IDTB = append(tmpWait.IDTB, int(i))
				}
				d.WaitHashCache[b.TxArrayX[i]] = tmpWait
			} else {
				s.Stat--
				s.Valid[i] = 1
				b.TxArray[i] = *d.TXCache[xxx[0]].Data
			}
		}
	}
	if s.Stat > 0 {
		s.Channel = make(chan bool, s.Stat)
	}
	*/
	s.Stat = 0
	if s.Stat == 0 {
		tmp, err := b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk)
		if !tmp {
			return err
		}
	}
	return nil
}
