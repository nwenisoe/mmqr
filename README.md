# MMQR Implementation

This package implements Myanmar QR (MMQR) merchant-presented payload generation, parsing, and validation without depending on a third-party EMV QR library.

It is designed around the Myanmar QR Specification for retail payments in Myanmar, merchant-presented mode.

## Quick Start For Package Users

If you are new to this package, think of `mmqr` as a low-level QR payload toolkit.

- Use `mmqr.Build(...)` when you already have clean merchant data and want an MMQR payload string.
- Use `mmqr.Parse(...)` when you already have an MMQR payload string and want typed Go fields back.
- Use `mmqr.ValidateInput(...)` before build if you want to reject bad input early.
- Use `mmqr.ValidatePayload(...)` after parse if you want semantic MMQR validation in addition to CRC validation.

What this package expects from you:

- A valid merchant ID, terminal ID, MCC, merchant name, city, and Myanmar language fields.
- Clear business rules outside the package for who can create QR, who can scan QR, expiry policy, and payment execution.

What this package returns:

- A raw MMQR payload string from `Build`.
- A typed `Payload` struct from `Parse`.

What this package does not do:

- It does not generate PNG/SVG QR images.
- It does not fetch merchant data from a database.
- It does not perform payment, settlement, or wallet balance updates.
- It does not enforce your product workflow; that belongs in the service layer such as `rest/internal/merchantqr/`.

Typical usage order:

1. Collect merchant/profile data from your own service or database.
2. Call `mmqr.ValidateInput(...)` if you want pre-checks before encoding.
3. Call `mmqr.Build(...)` to generate the payload string.
4. Send that payload string to mobile/web to render as a QR image.
5. Later, call `mmqr.Parse(...)` and `mmqr.ValidatePayload(...)` when a QR string comes back for verification.

## Features

- Builds static and dynamic MMQR payloads.
- Encodes EMV-style TLV fields: `ID + Length + Value`.
- Supports nested MMQR templates:
  - `26` Merchant Account Information
  - `62` Additional Data Field Template
  - `64` Merchant Information Language Template
- Calculates and verifies CRC field `63`.
- Parses MMQR payloads into typed Go structs.
- Validates MMQR-specific business rules.
- Supports Myanmar Unicode merchant information through tag `64`.
- Preserves unknown root TLV fields for auditability and forward compatibility.

## Package Location

```text
mmqr/
```

Import path inside this repository:

```go
import "mmqr"
```

## Supported MMQR Fields

The builder currently supports these root fields:

| Tag | Name | Notes |
| --- | --- | --- |
| `00` | Payload Format Indicator | Always `01` |
| `01` | Point of Initiation Method | `11` static, `12` dynamic |
| `26` | Merchant Account Information | MMQR merchant routing data |
| `52` | Merchant Category Code | 4-digit numeric MCC |
| `53` | Transaction Currency | Always `104` for MMK |
| `54` | Transaction Amount | Required for dynamic QR, optional for static QR |
| `55` | Tip or Convenience Indicator | Emitted only when supplied |
| `56` | Convenience Fee Fixed | Emitted only when `55=02` |
| `57` | Convenience Fee Percentage | Emitted only when `55=03` |
| `58` | Country Code | Always `MM` |
| `59` | Merchant Name | English/common character merchant name |
| `60` | Merchant City | English/common character city |
| `61` | Postal Code | Emitted only when supplied |
| `62` | Additional Data | Emitted only when at least one supported subtag has a value |
| `64` | Language Template | Myanmar Unicode merchant name/city |
| `63` | CRC | Generated automatically |

Optional fields with empty values are omitted instead of being emitted as zero-length TLVs. This keeps static QR payloads closer to the MMQR specification guidance, reduces payload size, and avoids interoperability issues with scanners that reject empty optional fields such as `5400`.

## Merchant Account Information: Tag `26`

Tag `26` is generated with:

| Subtag | Name | Value |
| --- | --- | --- |
| `00` | Globally Unique Identifier | `MM.COM.MMQR` |
| `01` | Merchant ID | 15-digit numeric merchant ID |
| `02` | Terminal ID | Provided terminal ID or default `000000` |

## Additional Data: Tag `62`

The MMQR spec treats tag `62` as optional. In this implementation, the builder emits tag `62` only when at least one supported subtag has a value. Inside `62`, only supplied subtags are encoded; unused values are omitted instead of being written as empty `0100`, `0200`, etc. fields.

Supported subtags:

| Subtag | Field |
| --- | --- |
| `01` | Bill Number |
| `02` | Mobile Number |
| `03` | Store Label |
| `04` | Loyalty Number |
| `05` | Reference Label |
| `06` | Customer Label |
| `07` | Terminal Label |
| `08` | Purpose of Transaction |
| `09` | Additional Consumer Data Request |
| `10` | Merchant Tax ID |
| `11` | Merchant Channel |
| `50.01` | Receiver Type inside a payment-system-specific template |

This implementation uses `62.50` as a payment-system-specific template. Inside that template:

| Subtag | Meaning |
| --- | --- |
| `01` | Receiver Type |

Receiver type values:

| Value | Receiver type |
| --- | --- |
| `1` | `consumer` |
| `2` | `merchant` |
| `3` | `agent` |

`AdditionalConsumerDataRequest` supports only `A`, `M`, and `E` indicators:

| Indicator | Meaning |
| --- | --- |
| `A` | Address |
| `M` | Mobile number |
| `E` | Email |

## Myanmar Language Template: Tag `64`

Tag `64` is generated with:

| Subtag | Name | Notes |
| --- | --- | --- |
| `00` | Language Preference | Always `my` |
| `01` | Merchant Name | Required Myanmar Unicode name |
| `02` | Merchant City | Optional Myanmar Unicode city |

For the current merchant QR flow, validation treats `64.01` as required because the local MMQR requirement documents mark it mandatory for Myanmar deployment.

## Build A Static QR Payload

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
		TerminalID:           "000000",
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

For a normal static QR, `54` transaction amount is omitted. The customer can enter the payment amount in the mobile app after scanning.

## Build A Dynamic QR Payload With Amount

```go
payload, err := mmqr.Build(mmqr.Input{
	Type:                 mmqr.QRTypeDynamic,
	MerchantID:           "123456000000001",
	TerminalID:           "000000",
	MerchantCategoryCode: "5411",
	TipOrConvenience:     mmqr.TipOrConvenienceFixed,
	ConvenienceFixedFee:  "100",
	Amount:               "1500.50",
	MerchantNameEN:       "SAMPLE STORE",
	MerchantCityEN:       "YANGON",
	PostalCode:           "11181",
	MerchantNameMM:       "နမူနာ စတိုး",
	MerchantCityMM:       "ရန်ကုန်",
	AdditionalData: mmqr.AdditionalData{
		BillNumber:                    "INV-1001",
		ReferenceLabel:                "ORDER-77",
		AdditionalConsumerDataRequest: "ME",
		MerchantChannel:               "400",
		ReceiverType:                  2,
	},
})
```

Dynamic QR generation requires `Amount`; validation rejects a dynamic QR input with an empty amount. Dynamic merchant QR also stores the backend reference in `62.05` so refresh and expiry checks can identify the generated QR. Receiver type is embedded in `62.50.01`.

## Parse And Validate A Payload

```go
parsed, err := mmqr.Parse(payload)
if err != nil {
	return err
}

fmt.Println(parsed.MerchantAccountInfo.MerchantID)
fmt.Println(parsed.TransactionAmount)
fmt.Println(parsed.Language.MerchantNameMM)
```

For semantic validation:

```go
if err := mmqr.ValidatePayload(payload); err != nil {
	return err
}
```

## Merchant QR Product Flow

The mobile app should receive the QR payload string from the backend and render the QR image locally. The backend does not generate PNG, SVG, or printable QR images.

Merchant QR flow helpers are implemented in:

```text
rest/internal/merchantqr/
```

Merchant REST APIs are exposed through Goa at:

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/v1/qr/receive/static` | Generate receiver static `qr_string` |
| `POST` | `/api/v1/qr/receive/dynamic` | Generate dynamic `qr_string` with amount and 15-minute TTL |
| `POST` | `/api/v1/qr/receive/dynamic/refresh` | Refresh an expired dynamic QR into a new `qr_string` |
| `POST` | `/api/v1/qr/receive/verify` | Verify QR validity, expiry, amount, and receiver details |

Mobile sends only the unique customer code and receiver type. The generated REST API currently uses numeric receiver type codes:

| Code | Receiver type |
| --- | --- |
| `1` | `consumer` |
| `2` | `merchant` |
| `3` | `agent` |

The backend resolves MMQR profile fields from `customers.code` and `customer_wallets.type`.

Static request:

```json
{
  "code": "M000000001",
  "receiver_type": 3
}
```

### Static Receive QR API ကို မြန်မာလို Step-by-Step ရှင်းပြချက်

Endpoint:

```http
POST /api/v1/qr/receive/static
```

ဒီ API က receiver ဘက်က "My QR" မှာပြမယ့် static QR payload string ထုတ်ပေးဖို့သုံးပါတယ်။ Static QR ဖြစ်တဲ့အတွက် QR တစ်ခုကို ထပ်ခါထပ်ခါ scan/pay လုပ်နိုင်ပြီး backend မှာ expiry မသတ်မှတ်ပါ။

#### Step 1: Mobile က request ပို့သည်

Mobile app က receiver ကိုရှာဖို့ `code` နဲ့ `receiver_type` ပဲပို့ပါတယ်။

```json
{
  "code": "M000000001",
  "receiver_type": 3
}
```

- `code` က `customers.code` ဖြစ်ပါတယ်။
- `receiver_type` က generated REST API မှာ numeric code ဖြစ်ပါတယ်။ `1 = consumer`, `2 = merchant`, `3 = agent` ဖြစ်ပါတယ်။
- Static QR request မှာ `amount` မပို့ပါ။ Amount ကို customer က scan လုပ်ပြီး payment app ထဲမှာဖြည့်နိုင်ပါတယ်။

#### Step 2: Backend က receiver profile ရှာသည်

Handler က `ProfileStore.GetByCode(code, receiver_type)` ကိုခေါ်ပြီး DB ထဲက receiver profile ကိုရှာပါတယ်။

Lookup logic:

- `customers.code = request.code`
- `customer_wallets.type = receiver_type` နဲ့ကိုက်တဲ့ wallet type
- `customers.mmqr_id` ကို MMQR Merchant ID အဖြစ်သုံးသည်
- Terminal ID ကို `000000` အမြဲသုံးသည်
- `business_name`, မရှိရင် `full_name` ကို receiver name အဖြစ်သုံးသည်
- `business_name_mm` ရှိရင် tag `64.01` Myanmar merchant name အဖြစ်သုံးသည်။ မရှိရင် `business_name`, ပြီးမှ `full_name` ကို fallback သုံးသည်။ English ကို Myanmar ဘာသာပြန်မလုပ်ပါ။
- township data ကနေ English city နဲ့ Myanmar city ကို resolve လုပ်သည်
- MCC ကို `reference_business_categories.mcc_code` ကနေယူပြီး မရှိရင် QR generation ကို fail လုပ်သည်။ receiver type placeholder MCC မသုံးပါ။

#### Step 3: Static MMQR payload တည်ဆောက်သည်

Service က `mmqr.Build` ကို `Type: static` နဲ့ခေါ်ပါတယ်။ အဲဒီအချိန်မှာ root tag `01` ကို `11` အဖြစ်ထည့်ပါတယ်။

Static QR မှာ အရေးကြီးတဲ့ tags:

| Tag | ထည့်တဲ့ value | အဓိပ္ပါယ် |
| --- | --- | --- |
| `00` | `01` | MMQR/EMV payload format |
| `01` | `11` | Static QR ဖြစ်ကြောင်းပြသည် |
| `26.00` | `MM.COM.MMQR` | MMQR payment scheme identifier |
| `26.01` | `customers.mmqr_id` | 15-digit Merchant ID |
| `26.02` | `000000` | Terminal ID |
| `52` | MCC | Merchant category code |
| `53` | `104` | MMK currency |
| `58` | `MM` | Myanmar country code |
| `59` | ASCII-safe name | English/common merchant name |
| `60` | ASCII-safe city | English/common city |
| `61` | `11181` | Current default postal code |
| `62.03` | `customers.code` | Store Label / receiver lookup code |
| `62.08` | `PAY_BY_QRCODE` | Purpose of Transaction |
| `62.50.01` | `1/2/3` | Receiver type: consumer / merchant / agent |
| `64.00` | `my` | Myanmar language code |
| `64.01` | `business_name_mm` / receiver name fallback | Myanmar Unicode merchant name |
| `64.02` | township/city | Myanmar Unicode city |
| `63` | CRC | Payload integrity checksum |

Static QR မှာ `54` Transaction Amount မပါပါ။ Empty amount ကို `5400` အဖြစ်မထည့်ဘဲ omit လုပ်ထားတာက scanner interoperability အတွက်ပိုကောင်းပါတယ်။

#### Step 4: Backend က response ပြန်ပေးသည်

Static QR response မှာ အဓိက `qr_string` ပါပါတယ်။

```json
{
  "qr_string": "00020101021126...",
  "kind": "static",
  "code": "M000000001",
  "receiver_type": 3
}
```

Static QR ဖြစ်လို့:

- `id` မလိုပါ
- `ttl_seconds` မပါပါ
- `expires_at` မပါပါ
- dynamic QR လို backend store ထဲသိမ်းပြီး expiry စောင့်မထားပါ

#### Step 5: Mobile က QR image render လုပ်သည်

Backend က PNG/SVG image မထုတ်ပါ။ Mobile app က response ထဲက `qr_string` ကို QR image အဖြစ် render လုပ်ပြီး receiver ရဲ့ My QR screen မှာပြပါတယ်။

#### Step 6: Customer က scan/pay လုပ်သည်

Customer က QR ကို scan လုပ်တဲ့အခါ scanner/payment app က payload ထဲက tags တွေကိုဖတ်ပြီး receiver ကိုသိနိုင်ပါတယ်။

- `26.01` က receiver MMQR merchant ID
- `26.02` က terminal ID ဖြစ်ပြီး current flow မှာ `000000` အမြဲသုံးသည်
- `62.03` က receiver/customer code
- `62.08` က purpose ဖြစ်ပြီး `PAY_BY_QRCODE` အဖြစ်ထည့်သည်
- `62.50.01` က receiver type ဖြစ်ပြီး `1=consumer`, `2=merchant`, `3=agent` အဖြစ် QR payload ထဲတန်းပါသည်
- `59`, `60`, `64` တွေက display name/city အတွက်သုံးနိုင်တဲ့ merchant info

Static QR မှာ amount မပါလို့ customer/payment app ဘက်မှာ amount input တောင်းရပါတယ်။

#### Step 7: Verify flow မှာ static QR ကို lifetime access အဖြစ်သတ်မှတ်သည်

`POST /api/v1/qr/receive/verify` ကိုခေါ်ရင် backend က QR payload ကို parse လုပ်ပြီး CRC, MMQR GUID, merchant ID, terminal ID စတာတွေစစ်ပါတယ်။ `01=11` ဖြစ်ရင် static QR လို့သတ်မှတ်ပြီး expiry မစစ်ပါ။

အကျဉ်းချုပ်ပြောရရင် static receive QR API က receiver profile ကို DB ကနေ resolve လုပ်ပြီး reusable MMQR payload string တစ်ခုထုတ်ပေးတာပါ။ Amount, expiry, backend reference ID မပါဘဲ QR string ကို mobile ဘက်မှာ image ပြောင်းပြပြီး customer scan/pay လုပ်တဲ့ flow ဖြစ်ပါတယ်။

Dynamic request:

```json
{
  "code": "M000000001",
  "receiver_type": 3,
  "amount": "1500"
}
```

Supported receiver types are `merchant`, `agent`, and `consumer`. Static QR can be created for all three receiver types when the matching wallet row exists. Dynamic QR is currently restricted to `merchant` receivers only.

Current DB mapping:

- `customers.code` is the mobile input and is stored in MMQR tag `62.03` Store Label.
- `customer_wallets.type` is matched from `receiver_type`.
- MMQR merchant ID uses `customers.mmqr_id`.
- Terminal ID always uses `000000`.
- Postal code currently defaults to `11181`.
- Purpose of transaction always uses `PAY_BY_QRCODE` in MMQR tag `62.08`.
- Receiver type is embedded in MMQR tag `62.50.01` using the payment-system-specific template under `62.50`.
- Receiver name uses `business_name`, then `full_name`. It does not fall back to `code`.
- Tag `64.01` uses `business_name_mm`, then falls back to the resolved receiver name. If only an English `business_name` exists, `64.01` keeps that English name; the system does not translate English into Myanmar.
- Verify reads receiver type directly from `62.50.01`. QR payloads without this field are not supported by the current verify flow.
- English MMQR fields `59` and `60` use common-character output. Non-ASCII characters are removed there to stay compatible with the root field character set.
- Tag `64.01` uses `business_name_mm` when available and remains mandatory in the payload.
- Tag `64.02` prefers township `name_mm`, then falls back to the English township name, then `YANGON`.
- MCC is resolved from `customers.business_category_id -> reference_business_categories.mcc_code`. If no category MCC exists, QR generation fails instead of emitting a placeholder MCC.

Implemented business rules from the merchant QR requirements:

- Static merchant QR is created for My QR and has lifetime access.
- Dynamic merchant QR is created from Create with amount and includes tag `54` transaction amount.
- Dynamic QR stores a backend reference in tag `62.05` and expires after 15 minutes.
- Dynamic QR records are retained in Redis for 24 hours so refresh/expired verification can distinguish generated QR strings from unknown payloads.
- Dynamic QR does not currently include tag `62.01` Bill Number; this flow uses `62.05` Reference Label for backend refresh/expiry lookup.
- Expired dynamic QR verification returns an expired result.
- Expired dynamic QR can be refreshed into a new QR string with a new reference and new 15-minute expiry.
- Scanner validation currently allows consumer scanning merchant QR, consumer scanning consumer static QR, and agent scanning agent static QR.
- Payment memo validation requires non-empty text and maximum 200 characters.

The QR package intentionally does not enforce customer role, account selection, PIN, fees, limits, discounts, commission, or receipt creation. Those belong to the payment execution layer after the QR has been scanned and verified.

Example dynamic merchant QR response shape:

```json
{
  "id": "B4Y7K3N2P9Q6R8T1V5W0X2ZA",
  "qr_string": "00020101021226...",
  "kind": "dynamic",
  "ttl_seconds": 900,
  "expires_at": "2026-06-01T10:15:00Z"
}
```

## Validation Rules

`ValidateInput` checks:

- QR type must be `static` or `dynamic`.
- Merchant ID must be 15-digit numeric.
- Terminal ID must be numeric and not exceed 25 digits when provided.
- Merchant category code must be 4-digit numeric.
- Amount is required for dynamic QR. When supplied, it must use MMQR decimal format, such as `98`, `98.`, or `98.73`, must have at most two decimal places, and must be greater than zero.
- Tip/convenience fields must follow the `55`, `56`, `57` relationship.
- English merchant name is required.
- English merchant name must use common ASCII characters and not exceed 25 bytes.
- English merchant city is required.
- English merchant city must use common ASCII characters and not exceed 15 bytes.
- Postal code must not exceed 10 bytes when provided.
- Myanmar merchant name is required.
- Myanmar merchant name must not exceed 25 characters.
- Myanmar merchant city must not exceed 15 characters when provided.
- Additional consumer data request can only contain `A`, `M`, and `E`.
- Merchant channel must be a valid 3-digit MMQR channel code when provided.

`ValidatePayload` additionally checks:

- CRC is present and valid.
- Payload format is `01`.
- Dynamic QR has a non-empty transaction amount.
- Currency is `104`.
- Country is `MM`.
- MMQR GUID is `MM.COM.MMQR`.
- Language code is `my`.

## CRC

The CRC field is tag `63`.

The package calculates CRC using:

- Polynomial: `0x1021`
- Initial value: `0xFFFF`
- Output format: uppercase 4-character hexadecimal string

The CRC is calculated over the payload with the `6304` CRC placeholder appended, then the final CRC value is appended.

## Length Handling

TLV field length is calculated using UTF-8 byte length.

This matters for Myanmar Unicode text because Myanmar characters use multiple bytes. The package also enforces the MMQR payload limit of `512` bytes.

## Run Tests

```powershell
go test ./mmqr ./rest/internal/merchantqr
```

## Current Scope

This package currently focuses on merchant-presented MMQR payload strings and merchant QR lifecycle rules.

Implemented:

- Payload generation
- Spec-friendly omission of empty optional TLV fields
- Payload parsing
- CRC generation and verification
- Core MMQR validation
- Static and dynamic QR support
- Merchant static QR lifecycle
- Merchant dynamic QR 15-minute lifecycle
- Merchant dynamic QR refresh
- Tags `26`, `55`, `56`, `57`, `61`, `62`, and `64`
- Unknown root field preservation

Not implemented:

- Backend QR image rendering, because the backend returns only the QR payload string to mobile.
- Consumer-presented QR, because the referenced MMQR document covers merchant-presented mode only.
- Direct provider settlement API calls, because those require bank, wallet, or switch-specific API contracts.
