package network

import (
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
)

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

