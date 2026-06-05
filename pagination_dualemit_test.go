package response

import (
	"encoding/json"
	"testing"
)

// TestPaginationDualEmit verifies that NewPagination produces BOTH the canonical
// 8-field snake_case shape AND the three legacy aliases (page, page_size, total)
// in the serialized JSON, so existing consumers keep working while new ones can
// migrate to the richer shape.
func TestPaginationDualEmit(t *testing.T) {
	// 25 items, page 2, 10 per page -> 3 total pages, middle page (has both neighbours).
	p := NewPagination(25, 2, 10)

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Canonical 8-field shape must all be present.
	canonical := []string{
		"current_page", "items_per_page", "total_items", "total_pages",
		"has_next_page", "has_prev_page", "next_page", "prev_page",
	}
	for _, k := range canonical {
		if _, ok := m[k]; !ok {
			t.Errorf("canonical key %q missing from JSON: %s", k, raw)
		}
	}

	// Legacy aliases must ALSO be present (dual-emit, no consumer breaks).
	legacy := []string{"page", "page_size", "total"}
	for _, k := range legacy {
		if _, ok := m[k]; !ok {
			t.Errorf("legacy alias %q missing from JSON: %s", k, raw)
		}
	}

	// Values: canonical and legacy must agree.
	assertInt := func(key string, want int) {
		var got int
		if err := json.Unmarshal(m[key], &got); err != nil {
			t.Fatalf("%s not an int: %v", key, err)
		}
		if got != want {
			t.Errorf("%s = %d, want %d", key, got, want)
		}
	}
	assertInt("current_page", 2)
	assertInt("page", 2)
	assertInt("items_per_page", 10)
	assertInt("page_size", 10)
	assertInt("total_items", 25)
	assertInt("total", 25)
	assertInt("total_pages", 3)

	// Middle page: both flags true, both neighbours non-null.
	var hasNext, hasPrev bool
	if err := json.Unmarshal(m["has_next_page"], &hasNext); err != nil || !hasNext {
		t.Errorf("has_next_page = %v (err %v), want true", hasNext, err)
	}
	if err := json.Unmarshal(m["has_prev_page"], &hasPrev); err != nil || !hasPrev {
		t.Errorf("has_prev_page = %v (err %v), want true", hasPrev, err)
	}
	if string(m["next_page"]) != "3" {
		t.Errorf("next_page = %s, want 3", m["next_page"])
	}
	if string(m["prev_page"]) != "1" {
		t.Errorf("prev_page = %s, want 1", m["prev_page"])
	}
}

// TestPaginationNullableNeighbours verifies next_page/prev_page serialize as
// JSON null at the boundaries rather than 0 or being omitted.
func TestPaginationNullableNeighbours(t *testing.T) {
	// Single page: no next, no prev.
	p := NewPagination(5, 1, 10)
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if string(m["next_page"]) != "null" {
		t.Errorf("next_page = %s, want null", m["next_page"])
	}
	if string(m["prev_page"]) != "null" {
		t.Errorf("prev_page = %s, want null", m["prev_page"])
	}
	var hasNext, hasPrev bool
	_ = json.Unmarshal(m["has_next_page"], &hasNext)
	_ = json.Unmarshal(m["has_prev_page"], &hasPrev)
	if hasNext {
		t.Error("has_next_page = true, want false on single page")
	}
	if hasPrev {
		t.Error("has_prev_page = true, want false on single page")
	}

	// First page of many: next present, prev null.
	first := NewPagination(25, 1, 10)
	rawFirst, _ := json.Marshal(first)
	var mf map[string]json.RawMessage
	_ = json.Unmarshal(rawFirst, &mf)
	if string(mf["prev_page"]) != "null" {
		t.Errorf("first page prev_page = %s, want null", mf["prev_page"])
	}
	if string(mf["next_page"]) != "2" {
		t.Errorf("first page next_page = %s, want 2", mf["next_page"])
	}

	// Last page of many: prev present, next null.
	last := NewPagination(25, 3, 10)
	rawLast, _ := json.Marshal(last)
	var ml map[string]json.RawMessage
	_ = json.Unmarshal(rawLast, &ml)
	if string(ml["next_page"]) != "null" {
		t.Errorf("last page next_page = %s, want null", ml["next_page"])
	}
	if string(ml["prev_page"]) != "2" {
		t.Errorf("last page prev_page = %s, want 2", ml["prev_page"])
	}
}

// TestErrorDetailDualEmitTraceID verifies the error envelope emits BOTH the
// legacy camelCase traceId AND the canonical snake_case trace_id with the same
// value, so consumers reading either key keep working.
func TestErrorDetailDualEmitTraceID(t *testing.T) {
	resp := InternalServerError(CLSTInternalServerError, "boom", "trace-xyz-123")

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var outer struct {
		Error map[string]json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(raw, &outer); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, k := range []string{"traceId", "trace_id"} {
		v, ok := outer.Error[k]
		if !ok {
			t.Errorf("error.%s missing from JSON: %s", k, raw)
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			t.Fatalf("error.%s not a string: %v", k, err)
		}
		if s != "trace-xyz-123" {
			t.Errorf("error.%s = %q, want trace-xyz-123", k, s)
		}
	}
}
