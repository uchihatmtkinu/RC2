package main

import (
	"fmt"

	//"golang.org/x/crypto/ed25519"
	//"golang.org/x/crypto/ed25519/cosi"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"bytes"
)

// This example demonstrates how to generate a
// collective signature involving two cosigners,
// and how to check the resulting collective signature.
func main() {

	// Create keypairs for the two cosigners.
	//sbMessage1 := bytes.NewReader([]byte{1})
	s1 := [64]byte{1}
	pubKey1, priKey1, _ := ed25519.GenerateKey(bytes.NewReader(s1[:]))
	//sbMessage2 := bytes.NewReader([]byte{2})
	s2 := [64]byte{2}
	pubKey2, priKey2, _ := ed25519.GenerateKey(bytes.NewReader(s2[:]))
	//s3 := [64]byte{3}
	//pubKey3, _, _ := ed25519.GenerateKey(bytes.NewReader(s3[:]))
	pubKeys := []ed25519.PublicKey{pubKey1, pubKey2}

	// Sign a test message.
	message := []byte("Hello World")
	sig := Sign(message, pubKeys, priKey1, priKey2)

	// Now verify the resulting collective signature.
	// This can be done by anyone any time, not just the leader.\
	currentPolicy := cosi.ThresholdPolicy(1)
	valid := cosi.Verify(pubKeys, currentPolicy, message, sig)
	fmt.Printf("signature valid: %v", valid)

	// Output:
	// signature valid: true
}

// Helper function to implement a bare-bones cosigning process.
// In practice the two cosigners would be on different machines
// ideally managed by independent administrators or key-holders.
func Sign(message []byte, pubKeys []ed25519.PublicKey,
	priKey1, priKey2 ed25519.PrivateKey) []byte {
	var cosimask []byte
	intilizeMaskBit(&cosimask, (3+7)>>3, cosi.Disabled)

	setMaskBit(0, cosi.Enabled, &cosimask)
	//setMaskBit(1, cosi.Enabled, &cosimask)
	//setMaskBit(1, cosi.Enabled, &cosimask)
	for i:= 0; i<3; i++{
		if maskBit(i, &cosimask) == cosi.Enabled{
			fmt.Println(i," is enabled")
		}
	}
	// Each cosigner first needs to produce a per-message commit.
	s1 := [64]byte{1}
	commit1, secret1, _ := cosi.Commit(bytes.NewReader(s1[:]))
	s2 := [64]byte{2}
	commit2, secret2, _ := cosi.Commit(bytes.NewReader(s2[:]))
	commits := []cosi.Commitment{commit1, commit2}

	// The leader then combines these into an aggregate commit.

	cosigners := cosi.NewCosigners(pubKeys, cosimask)
	fmt.Println(cosigners.CountTotal())
	fmt.Println(cosigners.CountEnabled())
	aggregatePublicKey := cosigners.AggregatePublicKey()
	aggregateCommit := cosigners.AggregateCommit(commits)

	// The cosigners now produce their parts of the collective signature.
	sigPart1 := cosi.Cosign(priKey1, secret1, message, aggregatePublicKey, aggregateCommit)
	sigPart2 := cosi.Cosign(priKey2, secret2, message, aggregatePublicKey, aggregateCommit)
	//fmt.Println("no use", sigPart2)
	sigParts := []cosi.SignaturePart{sigPart1, sigPart2}

	// Finally, the leader combines the two signature parts
	// into a final collective signature.
	sig := cosigners.AggregateSignature(aggregateCommit, sigParts)

	return sig
}


//intilizeMaskBit set all the mas to disable
func intilizeMaskBit(mask *[]byte, len int, value cosi.MaskBit){
	var setValue byte
	*mask = make([]byte, len)
	if value == cosi.Disabled {
		setValue = 0xff
	} else {
		setValue = 0x00
	}
	for i := 0; i < len; i++ {
		(*mask)[i] = setValue // all disabled
	}
}

// setMaskBit enable = 0 = false, disable = 1= true
func setMaskBit(signer int, value cosi.MaskBit, mask *[]byte) {
	byt := signer >> 3
	bit := byte(1) << uint(signer&7)

	if value == cosi.Disabled { // disable
		if (*mask)[byt]&bit == 0 { // was enabled
			(*mask)[byt] |= bit // disable it
		}
	} else { // enable
		if (*mask)[byt]&bit != 0 { // was disabled
			(*mask)[byt] &^= bit

		}
	}
}

// maskBit returns a boolean value indicating whether the indicated signer is Enabled or Disabled.
func maskBit(signer int, mask *[]byte) (value cosi.MaskBit) {
	byt := signer >> 3
	bit := byte(1) << uint(signer&7)
	return ((*mask)[byt] & bit) != 0
}

