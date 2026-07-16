package importer

import (
	"testing"

	"github.com/dagimg-dot/oc-sync/internal/types"
)

func TestLookupMapping_found(t *testing.T) {
	mappings := []types.Mapping{
		{RemoteProjectID: "proj_a", LocalProjectID: "proj_x"},
		{RemoteProjectID: "proj_b", LocalProjectID: "proj_y"},
	}
	m := lookupMapping(mappings, "proj_a")
	if m == nil {
		t.Fatal("lookupMapping returned nil")
	}
	if m.LocalProjectID != "proj_x" {
		t.Errorf("want LocalProjectID 'proj_x', got %q", m.LocalProjectID)
	}
}

func TestLookupMapping_notFound(t *testing.T) {
	mappings := []types.Mapping{
		{RemoteProjectID: "proj_a", LocalProjectID: "proj_x"},
	}
	m := lookupMapping(mappings, "proj_z")
	if m != nil {
		t.Fatal("lookupMapping should return nil for unknown ID")
	}
}

func TestLookupMapping_empty(t *testing.T) {
	m := lookupMapping(nil, "proj_a")
	if m != nil {
		t.Fatal("lookupMapping on nil should return nil")
	}
}

func TestNullStr(t *testing.T) {
	if s := nullStr(""); s != nil {
		t.Errorf("nullStr('') should be nil, got %v", s)
	}
	if s := nullStr("hello"); s == nil || *s != "hello" {
		t.Errorf("nullStr('hello') = %v, want 'hello'", s)
	}
}
