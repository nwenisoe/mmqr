package mmqr

// ReceiverType identifies the wallet category encoded in the QR payload.
type ReceiverType int16

// QRType describes whether the merchant QR is reusable or transaction-specific.
type QRType string

const (
	QRTypeStatic  QRType = "static"
	QRTypeDynamic QRType = "dynamic"
)

const (
	PayloadFormatIndicator = "01"
	GUIDMMQR               = "MM.COM.MMQR"
	CurrencyMMK            = "104"
	CountryMyanmar         = "MM"
	LanguageMyanmar        = "my"
	DefaultTerminalID      = "000000"
	MaxPayloadBytes        = 512
)

// Input contains the MMQR merchant-presented payload fields supported by this
// package. Currency and country are fixed to MMK and Myanmar by the builder.
type Input struct {
	Type                 QRType
	MerchantID           string
	TerminalID           string
	MerchantCategoryCode string
	Amount               string
	TipOrConvenience     TipOrConvenienceIndicator
	ConvenienceFixedFee  string
	ConveniencePercent   string

	MerchantNameEN string
	MerchantCityEN string
	PostalCode     string

	MerchantNameMM string
	MerchantCityMM string

	AdditionalData AdditionalData
}

// TipOrConvenienceIndicator maps to root tag 55.
type TipOrConvenienceIndicator string

const (
	TipOrConveniencePrompt TipOrConvenienceIndicator = "01"
	TipOrConvenienceFixed  TipOrConvenienceIndicator = "02"
	TipOrConveniencePct    TipOrConvenienceIndicator = "03"
)

// AdditionalData maps to the MMQR additional data field template, tag 62.
type AdditionalData struct {
	BillNumber                    string
	MobileNumber                  string
	StoreLabel                    string
	LoyaltyNumber                 string
	ReferenceLabel                string
	CustomerLabel                 string
	TerminalLabel                 string
	PurposeOfTransaction          string
	AdditionalConsumerDataRequest string
	MerchantTaxID                 string
	MerchantChannel               string
	ReceiverType                  ReceiverType
}

// Payload is a parsed MMQR merchant-presented payload.
type Payload struct {
	Raw                  string
	PayloadFormat        string
	PointOfInitiation    string
	MerchantAccountInfo  MerchantAccountInfo
	MerchantCategoryCode string
	TransactionCurrency  string
	TransactionAmount    string
	TipOrConvenience     string
	ConvenienceFixedFee  string
	ConveniencePercent   string
	CountryCode          string
	MerchantNameEN       string
	MerchantCityEN       string
	PostalCode           string
	AdditionalData       AdditionalData
	Language             LanguageTemplate
	CRC                  string
	UnknownFields        []Field
}

// Field is a parsed TLV field that is not currently mapped into a typed field.
type Field struct {
	ID       string
	Value    string
	Children []Field
}

// MerchantAccountInfo maps to the MMQR merchant account template, tag 26.
type MerchantAccountInfo struct {
	GloballyUniqueIdentifier string
	MerchantID               string
	TerminalID               string
}

// LanguageTemplate maps to the merchant information language template, tag 64.
type LanguageTemplate struct {
	LanguageCode   string
	MerchantNameMM string
	MerchantCityMM string
}

type node struct {
	ID       string
	Value    string
	Children []node
}

func (n node) isTemplate() bool {
	return len(n.Children) > 0
}
