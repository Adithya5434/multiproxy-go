package utils

import (
	"fmt"
)

func ReadVarInt(data []byte, offset int) (value int, pos int, err error) {
	value = 0
	shift := 0
	pos = offset

	for {
		if pos >= len(data) {
			return 0, pos, fmt.Errorf("varint extends beyond data length")
		}

		b := data[pos]
		value |= int(b&0x7F) << shift
		shift += 7
		pos++

		if (b & 0x80) == 0 {
			break
		}

		if shift > 35 {
			return 0, pos, fmt.Errorf("varint too big")
		}
	}

	return value, pos, nil
}

func WriteVarInt(value int) []byte {
	var result []byte

	for {
		temp := byte(value & 0x7F)
		value >>= 7

		if value != 0 {
			temp |= 0x80
		}

		result = append(result, temp)

		if value == 0 {
			break
		}
	}

	return result
}