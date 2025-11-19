package handler

import (
	"encoding/json"
	"testing"
)

func TestNormalizeJSONKeysToCamel(t *testing.T) {
	raw := []byte(`{"receipt_data":{"qr_code_url":"https://example.com?q=1","store_name":"Loja X","items":[{"temp_id":1,"description":"Item A","quantity":1,"unit":"UN","unit_price":5.0,"total":5.0}]}}`)

	normalized, err := normalizeJSONKeysToCamel(raw)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(normalized, &data); err != nil {
		t.Fatalf("unmarshal normalized json failed: %v", err)
	}

	// top-level should have "receiptData"
	if _, ok := data["receiptData"]; !ok {
		t.Fatalf("expected receiptData key after normalization, got keys: %v", data)
	}

	receipt := data["receiptData"].(map[string]interface{})
	if _, ok := receipt["qrCodeUrl"]; !ok {
		t.Fatalf("expected qrCodeUrl key inside receiptData after normalization, got keys: %v", receipt)
	}
}
