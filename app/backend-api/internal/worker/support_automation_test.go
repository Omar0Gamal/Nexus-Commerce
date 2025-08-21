package worker

import (
	"reflect"
	"testing"
)

func TestParseKeywords_ValidArray(t *testing.T) {
	raw := []byte(`["refund","shipping","broken"]`)
	got := parseKeywords(raw)
	want := []string{"refund", "shipping", "broken"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseKeywords: got %v, want %v", got, want)
	}
}

func TestParseKeywords_SingleKeyword(t *testing.T) {
	raw := []byte(`["password"]`)
	got := parseKeywords(raw)
	if len(got) != 1 || got[0] != "password" {
		t.Errorf("parseKeywords single: got %v", got)
	}
}

func TestParseKeywords_Empty(t *testing.T) {
	raw := []byte(`[]`)
	got := parseKeywords(raw)
	if len(got) != 0 {
		t.Errorf("parseKeywords empty array: got %v", got)
	}
}

func TestParseKeywords_KeywordWithComma(t *testing.T) {
	// Old naive string-split would break on commas inside values.
	raw := []byte(`["hello, world","foo"]`)
	got := parseKeywords(raw)
	if len(got) != 2 {
		t.Fatalf("parseKeywords comma-keyword: got %d items", len(got))
	}
	if got[0] != "hello, world" {
		t.Errorf("parseKeywords comma-keyword: got[0]=%q, want %q", got[0], "hello, world")
	}
}

func TestParseKeywords_InvalidJSON(t *testing.T) {
	got := parseKeywords([]byte(`not json`))
	if got != nil {
		t.Errorf("parseKeywords invalid JSON: expected nil, got %v", got)
	}
}

func TestParseKeywords_Null(t *testing.T) {
	got := parseKeywords([]byte(`null`))
	if got != nil {
		t.Errorf("parseKeywords null: expected nil, got %v", got)
	}
}

func TestParseKeywords_NilInput(t *testing.T) {
	got := parseKeywords(nil)
	if got != nil {
		t.Errorf("parseKeywords nil input: expected nil, got %v", got)
	}
}
