package mmqr

import (
	"fmt"
	"strings"
)

// encodeNodes serializes a list of nodes into TLV text in declaration order.
func encodeNodes(nodes []node) (string, error) {
	var b strings.Builder
	for _, n := range nodes {
		encoded, err := encodeNode(n)
		if err != nil {
			return "", err
		}
		b.WriteString(encoded)
	}
	return b.String(), nil
}

// encodeNode serializes a single node, encoding child templates recursively.
func encodeNode(n node) (string, error) {
	value := n.Value
	if n.isTemplate() {
		encodedChildren, err := encodeNodes(n.Children)
		if err != nil {
			return "", err
		}
		value = encodedChildren
	}

	if err := validateID(n.ID); err != nil {
		return "", err
	}
	valueLen := len([]rune(value))
	if valueLen > 99 {
		return "", fmt.Errorf("mmqr: field %s value exceeds 99 characters", n.ID)
	}
	return n.ID + fmt.Sprintf("%02d", valueLen) + value, nil
}

// validateID checks that a TLV field identifier uses exactly two numeric digits.
func validateID(id string) error {
	if len(id) != 2 || !isNumeric(id) {
		return fmt.Errorf("mmqr: field ID must be two digits: %q", id)
	}
	return nil
}
