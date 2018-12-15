package basic

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
)

//Sign returns the signature
func (a *RCSign) Sign(b []byte, prk *ecdsa.PrivateKey) {
	a.R = new(big.Int)
	a.S = new(big.Int)
	a.R, a.S, _ = ecdsa.Sign(rand.Reader, prk, b)
}

//New assign a new value
func (a *RCSign) New(b *RCSign) {
	a.R = new(big.Int)
	a.S = new(big.Int)
	(*a.R) = (*b.R)
	(*a.S) = (*b.S)
}

//Verify return the verification
func (a *RCSign) Verify(b []byte, puk *ecdsa.PublicKey) bool {
	return ecdsa.Verify(puk, b, a.R, a.S)
}

//SignToData encode the signature into []byte
func (a *RCSign) SignToData(b *[]byte) {
	EncodeDoubleBig(b, a.R, a.S)
}

//DataToSign decode the []byte into signature
func (a *RCSign) DataToSign(b *[]byte) error {
	a.R = new(big.Int)
	a.S = new(big.Int)
	err := DecodeDoubleBig(b, a.R, a.S)
	if err != nil {
		return fmt.Errorf("RCSign Signature decode failed %s", err)
	}
	return nil
}
