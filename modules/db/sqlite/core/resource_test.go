package core

import (
	"context"
	"encoding/json"
	"testing"
)

func TestResourceValuesAreOpaqueAndStable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	if db.ResourceID() == 0 {
		t.Fatal("expected connection resource id")
	}
	if db.Copy() != db {
		t.Fatal("expected connection copy to preserve handle identity")
	}
	if db.String() != "<db.sqlite.connection>" {
		t.Fatalf("unexpected connection string: %s", db.String())
	}

	data, err := json.Marshal(db)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	var marshaled string
	if err := json.Unmarshal(data, &marshaled); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if marshaled != "<db.sqlite.connection>" {
		t.Fatalf("unexpected marshal output: %s", marshaled)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	defer tx.Close()

	if tx.ResourceID() == 0 {
		t.Fatal("expected transaction resource id")
	}
	if tx.Copy() != tx {
		t.Fatal("expected transaction copy to preserve handle identity")
	}
	if tx.String() != "<db.sqlite.transaction>" {
		t.Fatalf("unexpected transaction string: %s", tx.String())
	}
}
