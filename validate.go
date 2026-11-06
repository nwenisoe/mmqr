package mmqr

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rivo/uniseg"
)

var amountPattern = regexp.MustCompile(`^\d+(\.\d{0,2})?$`)

const (
	amountMaxBytes             = 13
	conveniencePercentMaxBytes = 5
)

// ValidateInput checks MMQR merchant-presented business rules before building.
func ValidateInput(in Input) error {
	if in.Type != QRTypeStatic && in.Type != QRTypeDynamic {
		return ErrInvalidQRType
	}
	if len(in.MerchantID) != 15 || !isNumeric(in.MerchantID) {
		return ErrInvalidMerchantID
	}
	if in.TerminalID != "" && !isNumeric(in.TerminalID) {
		return ErrInvalidTerminalID
	}
	if in.TerminalID != "" && len(in.TerminalID) > 25 {
		return fmt.Errorf("mmqr: terminal ID must not exceed 25 digits")
	}
	if len(in.MerchantCategoryCode) != 4 || !isNumeric(in.MerchantCategoryCode) {
		return ErrInvalidMerchantCategoryCode
	}
	if err := validateAmount("amount", in.Amount, in.Type == QRTypeDynamic); err != nil {
		return err
	}
	if err := validateTipAndConvenience(in); err != nil {
		return err
	}
	if err := validateCommonText("merchant English name", in.MerchantNameEN, 25, true); err != nil {
		return err
	}
	if err := validateCommonText("merchant English city", in.MerchantCityEN, 15, true); err != nil {
		return err
	}
	if err := validateCommonText("postal code", in.PostalCode, 10, false); err != nil {
		return err
	}
	if err := validateMyanmarText("merchant Myanmar name", in.MerchantNameMM, 25, true); err != nil {
		return err
	}
	if err := validateMyanmarText("merchant Myanmar city", in.MerchantCityMM, 15, false); err != nil {
		return err
	}
	if err := validateAdditionalData(in.AdditionalData); err != nil {
		return err
	}
	return nil
}

// ValidatePayload checks parseability, CRC, and core MMQR semantic fields.
func ValidatePayload(raw string) error {
	payload, err := Parse(raw)
	if err != nil {
		return err
	}
	if payload.PayloadFormat != PayloadFormatIndicator {
		return fmt.Errorf("mmqr: payload format must be %s", PayloadFormatIndicator)
	}
	if payload.PointOfInitiation != "" && payload.PointOfInitiation != "11" && payload.PointOfInitiation != "12" {
		return fmt.Errorf("mmqr: point of initiation must be 11 or 12")
	}
	if payload.MerchantAccountInfo.GloballyUniqueIdentifier != GUIDMMQR {
		return fmt.Errorf("mmqr: merchant account GUID must be %s", GUIDMMQR)
	}
	if len(payload.MerchantAccountInfo.MerchantID) != 15 || !isNumeric(payload.MerchantAccountInfo.MerchantID) {
		return ErrInvalidMerchantID
	}
	if payload.MerchantAccountInfo.TerminalID != "" &&
		(!isNumeric(payload.MerchantAccountInfo.TerminalID) || len(payload.MerchantAccountInfo.TerminalID) > 25) {
		return ErrInvalidTerminalID
	}
	if len(payload.MerchantCategoryCode) != 4 || !isNumeric(payload.MerchantCategoryCode) {
		return ErrInvalidMerchantCategoryCode
	}
	if payload.TransactionCurrency != CurrencyMMK {
		return fmt.Errorf("mmqr: transaction currency must be %s", CurrencyMMK)
	}
	if err := validateAmount("amount", payload.TransactionAmount, payload.PointOfInitiation == "12"); err != nil {
		return err
	}
	if err := validateParsedTipAndConvenience(payload); err != nil {
		return err
	}
	if payload.ConvenienceFixedFee != "" {
		if err := validateAmount("convenience fixed fee", payload.ConvenienceFixedFee, false); err != nil {
			return err
		}
	}
	if payload.ConveniencePercent != "" {
		if err := validatePercentage("convenience percentage", payload.ConveniencePercent, false); err != nil {
			return err
		}
	}
	if payload.CountryCode != CountryMyanmar {
		return fmt.Errorf("mmqr: country code must be %s", CountryMyanmar)
	}
	if err := validateCommonText("merchant English name", payload.MerchantNameEN, 25, true); err != nil {
		return err
	}
	if err := validateCommonText("merchant English city", payload.MerchantCityEN, 15, true); err != nil {
		return err
	}
	if err := validateCommonText("postal code", payload.PostalCode, 10, false); err != nil {
		return err
	}
	if payload.Language.LanguageCode != LanguageMyanmar {
		return fmt.Errorf("mmqr: language code must be %s", LanguageMyanmar)
	}
	if err := validateMyanmarText("merchant Myanmar name", payload.Language.MerchantNameMM, 25, true); err != nil {
		return err
	}
	if err := validateMyanmarText("merchant Myanmar city", payload.Language.MerchantCityMM, 15, false); err != nil {
		return err
	}
	if err := validateAdditionalData(payload.AdditionalData); err != nil {
		return err
	}
	return nil
}

// validateCommonText checks ASCII/common-character MMQR fields with byte limits.
func validateCommonText(label string, value string, byteLimit int, required bool) error {
	if value == "" {
		if required {
			return fmt.Errorf("mmqr: %s is required", label)
		}
		return nil
	}
	if required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("mmqr: %s is required", label)
	}
	if len([]byte(value)) > byteLimit {
		return fmt.Errorf("mmqr: %s must not exceed %d bytes", label, byteLimit)
	}
	if !isCommonCharacterSet(value) {
		return fmt.Errorf("mmqr: %s must use the common character set", label)
	}
	return nil
}

// validateMyanmarText checks Myanmar-language text fields using grapheme limits.
func validateMyanmarText(label string, value string, limit int, required bool) error {
	if strings.TrimSpace(value) == "" {
		if required {
			return fmt.Errorf("mmqr: %s is required", label)
		}
		return nil
	}
	if graphemeCount(value) > limit {
		return fmt.Errorf("mmqr: %s must not exceed %d characters", label, limit)
	}
	return nil
}

// validateAdditionalData validates supported tag 62 fields and local extensions.
func validateAdditionalData(in AdditionalData) error {
	fields := []struct {
		name  string
		value string
		limit int
	}{
		{"bill number", in.BillNumber, 25},
		{"mobile number", in.MobileNumber, 25},
		{"store label", in.StoreLabel, 25},
		{"loyalty number", in.LoyaltyNumber, 25},
		{"reference label", in.ReferenceLabel, 25},
		{"customer label", in.CustomerLabel, 25},
		{"terminal label", in.TerminalLabel, 25},
		{"purpose of transaction", in.PurposeOfTransaction, 25},
		{"merchant tax ID", in.MerchantTaxID, 20},
	}
	for _, field := range fields {
		if err := validateCommonText(field.name, field.value, field.limit, false); err != nil {
			return err
		}
	}
	if in.AdditionalConsumerDataRequest != "" {
		if len([]byte(in.AdditionalConsumerDataRequest)) > 3 {
			return fmt.Errorf("mmqr: additional consumer data request must not exceed 3 bytes")
		}
		seen := map[rune]bool{}
		for _, r := range in.AdditionalConsumerDataRequest {
			if r != 'A' && r != 'M' && r != 'E' {
				return fmt.Errorf("mmqr: additional consumer data request supports only A, M, and E")
			}
			if seen[r] {
				return fmt.Errorf("mmqr: duplicate additional consumer data request indicator %q", r)
			}
			seen[r] = true
		}
	}
	if in.MerchantChannel != "" && (len(in.MerchantChannel) != 3 || !isNumeric(in.MerchantChannel)) {
		return fmt.Errorf("mmqr: merchant channel must be 3 digits")
	}
	if in.MerchantChannel != "" && !validMerchantChannel(in.MerchantChannel) {
		return fmt.Errorf("mmqr: merchant channel contains unsupported code")
	}
	if in.ReceiverType != 0 && in.ReceiverType != 1 && in.ReceiverType != 2 && in.ReceiverType != 3 {
		return fmt.Errorf("mmqr: receiver type must be 1, 2, or 3")
	}
	return nil
}

// validMerchantChannel reports whether the 3-digit merchant channel code is supported.
func validMerchantChannel(channel string) bool {
	validMedia := strings.ContainsRune("01234567", rune(channel[0]))
	validLocation := strings.ContainsRune("0123", rune(channel[1]))
	validPresence := strings.ContainsRune("0123", rune(channel[2]))
	return validMedia && validLocation && validPresence
}

// validateTipAndConvenience enforces the MMQR relationship between tags 55, 56, and 57.
func validateTipAndConvenience(in Input) error {
	switch in.TipOrConvenience {
	case "":
		if in.ConvenienceFixedFee != "" || in.ConveniencePercent != "" {
			return fmt.Errorf("mmqr: tip or convenience indicator is required when fee fields are set")
		}
	case TipOrConveniencePrompt:
		if in.ConvenienceFixedFee != "" || in.ConveniencePercent != "" {
			return fmt.Errorf("mmqr: prompted tip must not include fee amount or percentage")
		}
	case TipOrConvenienceFixed:
		if err := validateAmount("convenience fixed fee", in.ConvenienceFixedFee, true); err != nil {
			return err
		}
		if in.ConveniencePercent != "" {
			return fmt.Errorf("mmqr: fixed convenience fee must not include percentage")
		}
	case TipOrConveniencePct:
		if err := validatePercentage("convenience percentage", in.ConveniencePercent, true); err != nil {
			return err
		}
		if in.ConvenienceFixedFee != "" {
			return fmt.Errorf("mmqr: percentage convenience fee must not include fixed fee")
		}
	default:
		return fmt.Errorf("mmqr: invalid tip or convenience indicator")
	}
	return nil
}

// validateParsedTipAndConvenience reuses input validation rules for parsed payloads.
func validateParsedTipAndConvenience(payload *Payload) error {
	in := Input{
		TipOrConvenience:    TipOrConvenienceIndicator(payload.TipOrConvenience),
		ConvenienceFixedFee: payload.ConvenienceFixedFee,
		ConveniencePercent:  payload.ConveniencePercent,
	}
	return validateTipAndConvenience(in)
}

// validatePercentage validates percentage-style convenience fee values.
func validatePercentage(label string, percentage string, required bool) error {
	if err := validateAmount(label, percentage, required); err != nil {
		return err
	}
	if percentage != "" && len([]byte(percentage)) > conveniencePercentMaxBytes {
		return fmt.Errorf("mmqr: %s must not exceed %d bytes", label, conveniencePercentMaxBytes)
	}
	return nil
}

// validateAmount validates MMQR decimal amount fields and their size limits.
func validateAmount(label string, amount string, required bool) error {
	if amount == "" {
		if required {
			return fmt.Errorf("mmqr: %s is required", label)
		}
		return nil
	}
	if !amountPattern.MatchString(amount) {
		return ErrInvalidAmount
	}
	normalized := strings.TrimSuffix(amount, ".")
	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil || value <= 0 {
		return fmt.Errorf("mmqr: %s must be greater than zero", label)
	}
	if len([]byte(amount)) > amountMaxBytes {
		return fmt.Errorf("mmqr: %s must not exceed %d bytes", label, amountMaxBytes)
	}
	return nil
}

// isCommonCharacterSet reports whether the text stays within printable ASCII.
func isCommonCharacterSet(s string) bool {
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			return false
		}
	}
	return true
}

// isNumeric reports whether the string contains digits only.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// graphemeCount counts user-perceived characters for Myanmar text validation.
func graphemeCount(value string) int {
	gr := uniseg.NewGraphemes(value)
	count := 0
	for gr.Next() {
		count++
	}
	return count
}
