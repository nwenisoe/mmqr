package mmqr

import "fmt"

// Build creates a complete MMQR merchant-presented payload, including CRC.
func Build(in Input) (string, error) {
	if err := ValidateInput(in); err != nil {
		return "", err
	}

	pointOfInitiation := "11"
	if in.Type == QRTypeDynamic {
		pointOfInitiation = "12"
	}
	nodes := []node{
		{ID: "00", Value: PayloadFormatIndicator},
		{ID: "01", Value: pointOfInitiation},
		buildMerchantAccountInfo(in),
		{ID: "52", Value: in.MerchantCategoryCode},
		{ID: "53", Value: CurrencyMMK},
	}
	appendIfPresent := func(id string, value string) {
		if value != "" {
			nodes = append(nodes, node{ID: id, Value: value})
		}
	}

	appendIfPresent("54", in.Amount)
	appendIfPresent("55", string(in.TipOrConvenience))
	appendIfPresent("56", in.ConvenienceFixedFee)
	appendIfPresent("57", in.ConveniencePercent)

	nodes = append(nodes,
		node{ID: "58", Value: CountryMyanmar},
		node{ID: "59", Value: in.MerchantNameEN},
		node{ID: "60", Value: in.MerchantCityEN},
	)
	appendIfPresent("61", in.PostalCode)
	if additional, ok := buildAdditionalData(in.AdditionalData); ok {
		nodes = append(nodes, additional)
	}
	nodes = append(nodes, buildLanguageTemplate(in))

	base, err := encodeNodes(nodes)
	if err != nil {
		return "", err
	}
	payload := base + crcPlaceholder + computeCRC(base+crcPlaceholder)
	if len([]byte(payload)) > MaxPayloadBytes {
		return "", ErrPayloadTooLong
	}
	return payload, nil
}

// buildMerchantAccountInfo assembles tag 26 with the MMQR routing fields.
func buildMerchantAccountInfo(in Input) node {
	terminalID := in.TerminalID
	if terminalID == "" {
		terminalID = DefaultTerminalID
	}

	return node{
		ID: "26",
		Children: []node{
			{ID: "00", Value: GUIDMMQR},
			{ID: "01", Value: in.MerchantID},
			{ID: "02", Value: terminalID},
		},
	}
}

// buildAdditionalData assembles tag 62 and omits it entirely when no fields are present.
func buildAdditionalData(in AdditionalData) (node, bool) {
	children := make([]node, 0, 12)
	add := func(id string, value string) {
		if value != "" {
			children = append(children, node{ID: id, Value: value})
		}
	}

	add("01", in.BillNumber)
	add("02", in.MobileNumber)
	add("03", in.StoreLabel)
	add("04", in.LoyaltyNumber)
	add("05", in.ReferenceLabel)
	add("06", in.CustomerLabel)
	add("07", in.TerminalLabel)
	add("08", in.PurposeOfTransaction)
	add("09", in.AdditionalConsumerDataRequest)
	add("10", in.MerchantTaxID)
	add("11", in.MerchantChannel)
	if in.ReceiverType != 0 {
		children = append(children, node{
			ID: "50",
			Children: []node{
				{ID: "01", Value: fmt.Sprintf("%d", in.ReceiverType)},
			},
		})
	}

	if len(children) == 0 {
		return node{}, false
	}
	return node{ID: "62", Children: children}, true
}

// buildLanguageTemplate assembles tag 64 for Myanmar-language merchant fields.
func buildLanguageTemplate(in Input) node {
	children := []node{
		{ID: "00", Value: LanguageMyanmar},
		{ID: "01", Value: in.MerchantNameMM},
	}
	if in.MerchantCityMM != "" {
		children = append(children, node{ID: "02", Value: in.MerchantCityMM})
	}
	return node{ID: "64", Children: children}
}
