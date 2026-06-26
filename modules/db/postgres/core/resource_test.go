package core

import "testing"

func TestResourceValues(t *testing.T) {
	t.Parallel()

	db := NewConnection(nil)
	if db.String() != "<db.postgres.connection>" {
		t.Fatalf("unexpected connection string: %s", db.String())
	}
	if db.Copy() != db {
		t.Fatal("expected connection copy to preserve handle identity")
	}

	tx := NewTransaction(db, nil)
	if tx.String() != "<db.postgres.transaction>" {
		t.Fatalf("unexpected transaction string: %s", tx.String())
	}
	if tx.Copy() != tx {
		t.Fatal("expected transaction copy to preserve handle identity")
	}
}
