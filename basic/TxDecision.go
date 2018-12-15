package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/gVar"
)

//Set initiates the TxDecision given the TxList and the account
func (a *TxDecision) Set(ID uint32, target uint32, single uint32) error {
	a.TxCnt = 0
	a.ID = ID
	a.Target = target
	a.Single = single
	if single == 0 {
		a.Sig = make([]RCSign, gVar.ShardCnt)
	} else {
		a.Sig = make([]RCSign, 1)
	}
	return nil
}

//Add adds one decision given the result
func (a *TxDecision) Add(x byte) error {
	tmpNum := a.TxCnt % 8
	a.TxCnt++
	if tmpNum == 0 {
		a.Decision = append(a.Decision, byte(0))
	}
	tmp := len(a.Decision) - 1

	a.Decision[tmp] += x << tmpNum

	return nil
}

//Sign signs the TxDecision
func (a *TxDecision) Sign(prk *ecdsa.PrivateKey, x uint32) {
	tmp := make([]byte, 0, 36+len(a.Decision))
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, byteSlice(a.ID)...)
	tmpHash := sha256.Sum256(tmp)
	a.Sig[x].Sign(tmpHash[:], prk)
	//fmt.Println(a.Sig[x], " is the sig of ", tmpHash, "Hash is", base58.Encode(a.HashID[:]))
}

//Verify the signature using public key
func (a *TxDecision) Verify(puk *ecdsa.PublicKey, x uint32) bool {
	tmp := make([]byte, 0, 36+len(a.Decision))
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, byteSlice(a.ID)...)
	tmpHash := sha256.Sum256(tmp)
	//fmt.Println("Verify ", a.Sig[x], " is the sig of ", tmpHash, "Hash is", base58.Encode(a.HashID[:]))
	return a.Sig[x].Verify(tmpHash[:], puk)
}

//Encode encodes the TxDecision into []byte
func (a *TxDecision) Encode(tmp *[]byte) {
	Encode(tmp, a.ID)
	Encode(tmp, a.TxCnt)
	Encode(tmp, a.Target)
	Encode(tmp, &a.Decision)
	Encode(tmp, a.Single)
	if a.Single == 1 {
		Encode(tmp, &a.Sig[0])
	} else {
		Encode(tmp, &a.HashID)
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			Encode(tmp, &a.Sig[i])
		}
	}

}

// Decode decodes the []byte into TxDecision
func (a *TxDecision) Decode(buf *[]byte) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxDecsion ID decode failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecsion TxCnt decode failed: %s", err)
	}
	err = Decode(buf, &a.Target)
	if err != nil {
		return fmt.Errorf("TxDecsion Target decode failed: %s", err)
	}
	err = Decode(buf, &a.Decision)
	if err != nil {
		return fmt.Errorf("TxDecsion Decision decode failed: %s", err)
	}
	err = Decode(buf, &a.Single)
	if err != nil {
		return fmt.Errorf("TxDecision Single decode failed: %s", err)
	}
	if a.Single == 1 {
		a.Sig = make([]RCSign, 1)
		err = Decode(buf, &a.Sig[0])
		if err != nil {
			return fmt.Errorf("TxDecision Sig decode failed: %s", err)
		}
	} else {
		err = Decode(buf, &a.HashID)
		if err != nil {
			return fmt.Errorf("TxDecsion HashID decode failed: %s", err)
		}
		a.Sig = make([]RCSign, gVar.ShardCnt)
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			err = Decode(buf, &a.Sig[i])
			if err != nil {
				return fmt.Errorf("TxDecision Sig decode failed: %s", err)
			}
		}
	}
	return nil
}

//Print is
func (a *TxDecision) Print() {
	fmt.Println("----------------TxDecision------------------")
	fmt.Println("TxDecision: ID: ", a.ID, " TxCnt: ", a.TxCnt, " Target: ", a.Target)
	fmt.Println("Single : ", a.Single, " Hash: ", base58.Encode(a.HashID[:]))
	fmt.Print("Result: ")
	for i := uint32(0); i < a.TxCnt; i++ {
		fmt.Print((a.Decision[i/8] >> (i % 8)) & 1)
	}
	fmt.Println()
}
