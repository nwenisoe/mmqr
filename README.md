# MMQR

A Go library for building, parsing, and validating static and dynamic merchant-presented Myanmar QR (MMQR) payment payloads.

## Build a payload

```go
package main

import (
	"fmt"
	"log"

	"mmqr"
)

func main() {
	payload, err := mmqr.Build(mmqr.Input{
		Type:                 mmqr.QRTypeStatic,
		MerchantID:           "123456000000001",
		MerchantCategoryCode: "5411",
		MerchantNameEN:       "SAMPLE STORE",
		MerchantCityEN:       "YANGON",
		MerchantNameMM:       "နမူနာ စတိုး",
		MerchantCityMM:       "ရန်ကုန်",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(payload)
}
```

For a dynamic QR, set `Type` to `mmqr.QRTypeDynamic` and provide an `Amount` string such as `"1500.50"`. An omitted terminal ID defaults to `000000`.

`Build` validates the input and returns the payload string with its CRC. Use `ValidateInput` to check input separately.

## Parse and validate

Use `Parse` to decode the payload into a typed `Payload` and verify its CRC. Use `ValidatePayload` to check MMQR field rules as well.

```go
parsed, err := mmqr.Parse(payload)
if err != nil {
	log.Fatal(err)
}

fmt.Println(parsed.MerchantAccountInfo.MerchantID)
```

## Supported fields

- Merchant account information (tag `26`)
- Transaction amount, tip, and convenience fee fields
- Additional data such as bill and reference labels (tag `62`)
- Myanmar-language merchant name and city (tag `64`)

The builder uses MMK (`104`) and Myanmar (`MM`). TLV lengths count UTF-8 bytes, and payloads are limited to 512 bytes.

## Tests

```sh
go test ./...
```
