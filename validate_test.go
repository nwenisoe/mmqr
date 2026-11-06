package mmqr

import (
	"strings"
	"testing"
)

func TestValidateMyanmarText_MMQRNameLimit25(t *testing.T) {
	const limit = 25
	const label = "Myanmar Name"

	tests := []struct {
		name          string
		value         string
		required      bool
		expectedCount int // Provided for test transparency
		expectErr     bool
		errMsg        string
	}{
		{
			name:          "Valid - Common Myanmar Name within 25 limit",
			value:         "အောင်ဆန်းစုကြည်", // 7 graphemes
			required:      true,
			expectedCount: 7,
			expectErr:     false,
		},
		{
			name:          "Valid - Complex Myanmar clusters within limit",
			value:         "မြန်မာနိုင်ငံတော်ဗဟိုဘဏ်", // 12 graphemes
			required:      true,
			expectedCount: 12,
			expectErr:     false,
		},
		{
			name:          "Valid - Exactly 25 graphemes (Boundary condition)",
			value:         "ကကကကကကကကကကကကကကကကကကကကကကကကက", // 25 graphemes
			required:      true,
			expectedCount: 25,
			expectErr:     false,
		},
		{
			name:          "Invalid - Exceeds 25 graphemes by 1",
			value:         "ကကကကကကကကကကကကကကကကကကကကကကကကကက", // 26 graphemes
			required:      true,
			expectedCount: 26,
			expectErr:     true,
			errMsg:        "mmqr: Myanmar Name must not exceed 25 characters",
		},
		{
			name:          "Invalid - Long Myanmar Merchant Name (> 25 graphemes)",
			value:         "ဦးမြန်မာပြည်ထောင်စုသမ္မတမြန်မာနိုင်ငံတော်", // 27 graphemes
			required:      false,
			expectedCount: 27,
			expectErr:     true,
			errMsg:        "mmqr: Myanmar Name must not exceed 25 characters",
		},
		{
			name:          "Valid - Optional empty string",
			value:         "",
			required:      false,
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name:          "Invalid - Required empty string",
			value:         "",
			required:      true,
			expectedCount: 0,
			expectErr:     true,
			errMsg:        "mmqr: Myanmar Name is required",
		},
		{
			name:          "Invalid - Required whitespace string",
			value:         "   ",
			required:      true,
			expectedCount: 0,
			expectErr:     true,
			errMsg:        "mmqr: Myanmar Name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify our grapheme counter math matches expectations
			if actualGraphemes := graphemeCount(tt.value); strings.TrimSpace(tt.value) != "" && actualGraphemes != tt.expectedCount {
				t.Errorf("grapheme count mismatch: calculated %d, expected %d", actualGraphemes, tt.expectedCount)
			}

			err := validateMyanmarText(label, tt.value, limit, tt.required)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
