package mmqr

import "errors"

var (
	ErrInvalidQRType               = errors.New("mmqr: type must be static or dynamic")
	ErrInvalidMerchantID           = errors.New("mmqr: merchant ID must be 15-digit numeric")
	ErrInvalidTerminalID           = errors.New("mmqr: terminal ID must be numeric")
	ErrInvalidMerchantCategoryCode = errors.New("mmqr: merchant category code must be 4-digit numeric")
	ErrInvalidAmount               = errors.New("mmqr: amount must be digits with an optional decimal point")
	ErrInvalidCRC                  = errors.New("mmqr: invalid CRC")
	ErrMissingCRC                  = errors.New("mmqr: missing CRC")
	ErrPayloadTooLong              = errors.New("mmqr: payload exceeds 512 bytes")
)
