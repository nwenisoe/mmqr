package mmqr

import "fmt"

const crcPlaceholder = "6304"

// computeCRC calculates the CRC-16/CCITT checksum for an MMQR payload body.
func computeCRC(payload string) string {
	var crc uint16 = 0xFFFF
	for _, b := range []byte(payload) {
		crc ^= uint16(b) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc)
}

// hasValidCRC reports whether the payload ends with a valid MMQR CRC field.
func hasValidCRC(raw string) bool {
	if len(raw) < 8 {
		return false
	}
	crcIndex := len(raw) - 8
	if raw[crcIndex:crcIndex+4] != crcPlaceholder {
		return false
	}
	return computeCRC(raw[:crcIndex]+crcPlaceholder) == raw[crcIndex+4:]
}
