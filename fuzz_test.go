package quuid

import (
	"strings"
	"testing"
)

func FuzzParseID(f *testing.F) {
	f.Add(DeriveID("seed", []byte("one")).String())
	f.Add("")
	f.Add("qid_invalid")
	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseID(input)
		if err != nil {
			return
		}
		roundTrip, err := ParseID(id.String())
		if err != nil {
			t.Fatalf("canonical value did not parse: %v", err)
		}
		if !id.Equal(roundTrip) {
			t.Fatal("round trip mismatch")
		}
	})
}

func FuzzParseStrongID(f *testing.F) {
	f.Add(DeriveStrongID("seed", []byte("one")).String())
	f.Add("qsid_invalid")
	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseStrongID(input)
		if err != nil {
			return
		}
		roundTrip, err := ParseStrongID(id.String())
		if err != nil || !id.Equal(roundTrip) {
			t.Fatalf("round trip failed: %v", err)
		}
	})
}

func FuzzParseSortableID(f *testing.F) {
	f.Add("qsort_0000000000000000000000000000000000000000000000000000000000000000")
	f.Add("qsort_invalid")
	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseSortableID(input)
		if err != nil {
			return
		}
		roundTrip, err := ParseSortableID(id.String())
		if err != nil || !id.Equal(roundTrip) {
			t.Fatalf("round trip failed: %v", err)
		}
	})
}

func FuzzParseUUID(f *testing.F) {
	f.Add("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	f.Add("")
	f.Add(" 018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseUUID(input)
		if err != nil {
			return
		}
		canonical := id.String()
		if canonical != input && strings.ToLower(input) != canonical {
			t.Fatalf("strict parser accepted non-canonical input %q", input)
		}
		roundTrip, err := ParseUUID(canonical)
		if err != nil || roundTrip != id {
			t.Fatalf("round trip failed: %v", err)
		}
	})
}

func FuzzParseUUIDBytes(f *testing.F) {
	f.Add([]byte("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22"))
	f.Add([]byte(""))
	f.Add([]byte(" 018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22"))
	f.Fuzz(func(t *testing.T, input []byte) {
		id, err := ParseUUIDBytes(input)
		if err != nil {
			return
		}
		canonical := id.String()
		inputString := string(input)
		if canonical != inputString && strings.ToLower(inputString) != canonical {
			t.Fatalf("strict byte parser accepted non-canonical input %q", input)
		}
		roundTrip, err := ParseUUID(canonical)
		if err != nil || roundTrip != id {
			t.Fatalf("round trip failed: %v", err)
		}
	})
}

func FuzzParseUUIDLoose(f *testing.F) {
	f.Add("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	f.Add("{018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22}")
	f.Add("urn:uuid:018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseUUIDLoose(input)
		if err != nil {
			return
		}
		roundTrip, err := ParseUUID(id.String())
		if err != nil || roundTrip != id {
			t.Fatalf("round trip failed: %v", err)
		}
	})
}
