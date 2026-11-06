package mmqr

import (
	"fmt"
	"strconv"
)

// Parse decodes an MMQR merchant-presented payload and verifies its CRC.
func Parse(raw string) (*Payload, error) {
	if raw == "" {
		return nil, fmt.Errorf("mmqr: payload is empty")
	}
	if len([]byte(raw)) > MaxPayloadBytes {
		return nil, ErrPayloadTooLong
	}
	if len(raw) < 8 || raw[len(raw)-8:len(raw)-4] != crcPlaceholder {
		return nil, ErrMissingCRC
	}
	if !hasValidCRC(raw) {
		return nil, ErrInvalidCRC
	}

	nodes, err := parseNodes(raw)
	if err != nil {
		return nil, err
	}
	return decodePayload(nodes, raw)
}

// parseNodes walks a TLV string from left to right and returns all decoded nodes.
func parseNodes(raw string) ([]node, error) {
	nodes := make([]node, 0)
	for pos := 0; pos < len(raw); {
		n, next, err := parseNode(raw, pos)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
		pos = next
	}
	return nodes, nil
}

// parseNode decodes one TLV field and recursively parses known template fields.
func parseNode(raw string, pos int) (node, int, error) {
	if pos+4 > len(raw) {
		return node{}, pos, fmt.Errorf("mmqr: incomplete field header at byte %d", pos)
	}
	id := raw[pos : pos+2]
	if err := validateID(id); err != nil {
		return node{}, pos, err
	}
	lengthText := raw[pos+2 : pos+4]
	length, err := strconv.Atoi(lengthText)
	if err != nil {
		return node{}, pos, fmt.Errorf("mmqr: invalid length for field %s", id)
	}
	valueStart := pos + 4
	valueEnd, ok := byteIndexAfterRunes(raw, valueStart, length)
	if !ok {
		return node{}, pos, fmt.Errorf("mmqr: field %s length exceeds payload", id)
	}
	value := raw[valueStart:valueEnd]
	n := node{ID: id, Value: value}
	if isTemplateID(id) {
		children, err := parseNodes(value)
		if err != nil {
			return node{}, pos, fmt.Errorf("mmqr: parse template %s: %w", id, err)
		}
		n.Children = children
	}
	return n, valueEnd, nil
}

// byteIndexAfterRunes converts a rune-counted field length into a byte end index.
func byteIndexAfterRunes(raw string, start int, count int) (int, bool) {
	if count == 0 {
		return start, true
	}
	seen := 0
	for offset := range raw[start:] {
		if seen == count {
			return start + offset, true
		}
		seen++
	}
	if seen == count {
		return len(raw), true
	}
	return 0, false
}

// isTemplateID reports whether the field ID contains nested TLV children.
func isTemplateID(id string) bool {
	return id == "26" || id == "62" || id == "64"
}

// decodePayload maps parsed root nodes into the public Payload structure.
func decodePayload(nodes []node, raw string) (*Payload, error) {
	payload := &Payload{Raw: raw}
	for _, n := range nodes {
		switch n.ID {
		case "00":
			payload.PayloadFormat = n.Value
		case "01":
			payload.PointOfInitiation = n.Value
		case "26":
			payload.MerchantAccountInfo = decodeMerchantAccountInfo(n.Children)
		case "52":
			payload.MerchantCategoryCode = n.Value
		case "53":
			payload.TransactionCurrency = n.Value
		case "54":
			payload.TransactionAmount = n.Value
		case "55":
			payload.TipOrConvenience = n.Value
		case "56":
			payload.ConvenienceFixedFee = n.Value
		case "57":
			payload.ConveniencePercent = n.Value
		case "58":
			payload.CountryCode = n.Value
		case "59":
			payload.MerchantNameEN = n.Value
		case "60":
			payload.MerchantCityEN = n.Value
		case "61":
			payload.PostalCode = n.Value
		case "62":
			payload.AdditionalData = decodeAdditionalData(n.Children)
		case "63":
			payload.CRC = n.Value
		case "64":
			payload.Language = decodeLanguageTemplate(n.Children)
		default:
			payload.UnknownFields = append(payload.UnknownFields, toField(n))
		}
	}
	return payload, nil
}

// toField converts an internal parse node into the public Field representation.
func toField(n node) Field {
	children := make([]Field, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, toField(child))
	}
	return Field{ID: n.ID, Value: n.Value, Children: children}
}

// decodeMerchantAccountInfo maps tag 26 child fields into MerchantAccountInfo.
func decodeMerchantAccountInfo(nodes []node) MerchantAccountInfo {
	var out MerchantAccountInfo
	for _, n := range nodes {
		switch n.ID {
		case "00":
			out.GloballyUniqueIdentifier = n.Value
		case "01":
			out.MerchantID = n.Value
		case "02":
			out.TerminalID = n.Value
		}
	}
	return out
}

// decodeAdditionalData maps tag 62 child fields into AdditionalData.
func decodeAdditionalData(nodes []node) AdditionalData {
	var out AdditionalData
	for _, n := range nodes {
		switch n.ID {
		case "01":
			out.BillNumber = n.Value
		case "02":
			out.MobileNumber = n.Value
		case "03":
			out.StoreLabel = n.Value
		case "04":
			out.LoyaltyNumber = n.Value
		case "05":
			out.ReferenceLabel = n.Value
		case "06":
			out.CustomerLabel = n.Value
		case "07":
			out.TerminalLabel = n.Value
		case "08":
			out.PurposeOfTransaction = n.Value
		case "09":
			out.AdditionalConsumerDataRequest = n.Value
		case "10":
			out.MerchantTaxID = n.Value
		case "11":
			out.MerchantChannel = n.Value
		case "50":
			children := n.Children
			if len(children) == 0 && n.Value != "" {
				parsedChildren, err := parseNodes(n.Value)
				if err == nil {
					children = parsedChildren
				}
			}
			for _, child := range children {
				switch child.ID {
				case "01":
					code, err := strconv.Atoi(child.Value)
					if err == nil {
						out.ReceiverType = ReceiverType(code)
					}
				}
			}
		}
	}
	return out
}

// decodeLanguageTemplate maps tag 64 child fields into LanguageTemplate.
func decodeLanguageTemplate(nodes []node) LanguageTemplate {
	var out LanguageTemplate
	for _, n := range nodes {
		switch n.ID {
		case "00":
			out.LanguageCode = n.Value
		case "01":
			out.MerchantNameMM = n.Value
		case "02":
			out.MerchantCityMM = n.Value
		}
	}
	return out
}
