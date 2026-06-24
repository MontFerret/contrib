package resource

import "testing"

func TestIDGenerator(t *testing.T) {
	var ids IDGenerator

	if got := ids.Next(); got != 1 {
		t.Fatalf("first id = %d, want 1", got)
	}
	if got := ids.Next(); got != 2 {
		t.Fatalf("second id = %d, want 2", got)
	}
}

func TestHash(t *testing.T) {
	got := Hash("db.sqlite.connection", 1)
	if got != 9895857375763297966 {
		t.Fatalf("hash = %d, want stable value", got)
	}

	if Hash("db.sqlite.connection", 1) == Hash("db.sqlite.connection", 2) {
		t.Fatal("expected different IDs to hash differently")
	}
	if Hash("db.sqlite.connection", 1) == Hash("db.sqlite.transaction", 1) {
		t.Fatal("expected different type names to hash differently")
	}
}

func TestDisplayAndJSON(t *testing.T) {
	display := Display("http.client")
	if display != "<http.client>" {
		t.Fatalf("display = %q", display)
	}

	encoded, err := MarshalDisplayJSON("http.client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(encoded) != `"<http.client>"` {
		t.Fatalf("json = %s", encoded)
	}

	encoded, err = MarshalStringJSON(`a"b`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(encoded) != `"a\"b"` {
		t.Fatalf("json = %s", encoded)
	}
}
