package core

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"hash/fnv"
	"sync"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Transaction is an opaque SQLite transaction handle exposed to Ferret.
type Transaction struct {
	parent         *Connection
	tx             *sql.Tx
	mu             sync.Mutex
	id             uint64
	finished       bool
	closedByParent bool
}

// NewTransaction creates a tracked SQLite transaction handle.
func NewTransaction(parent *Connection, tx *sql.Tx) *Transaction {
	return &Transaction{
		parent: parent,
		tx:     tx,
		id:     newResourceID(),
	}
}

func (t *Transaction) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	tx, err := t.activeTx("QUERY")
	if err != nil {
		return nil, err
	}

	return querySQL(ctx, "QUERY", tx, q)
}

func (t *Transaction) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return runtime.DefaultQueryOne(ctx, q, t.Query)
}

func (t *Transaction) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return runtime.DefaultQueryCount(ctx, q, t.Query)
}

func (t *Transaction) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return runtime.DefaultQueryExists(ctx, q, t.Query)
}

func (t *Transaction) Commit() error {
	tx, err := t.finish("COMMIT")
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return OperationError("COMMIT", err)
	}

	return nil
}

func (t *Transaction) Rollback() error {
	tx, err := t.finish("ROLLBACK")
	if err != nil {
		return err
	}

	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return OperationError("ROLLBACK", err)
	}

	return nil
}

func (t *Transaction) Close() error {
	tx := t.closeActive(false)
	if tx == nil {
		return nil
	}

	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return OperationError("ROLLBACK", err)
	}

	return nil
}

func (t *Transaction) ResourceID() uint64 {
	return t.id
}

func (t *Transaction) String() string {
	return "<db.sqlite.transaction>"
}

func (t *Transaction) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte("db.sqlite.transaction:"))

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, t.id)
	h.Write(bytes)

	return h.Sum64()
}

func (t *Transaction) Copy() runtime.Value {
	return t
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return []byte(`"<db.sqlite.transaction>"`), nil
}

func (t *Transaction) activeTx(operation string) (*sql.Tx, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closedByParent {
		return nil, OperationError(operation, errParentClosed)
	}
	if t.finished {
		return nil, OperationError(operation, errTransactionDone)
	}
	if t.parent.isClosed() {
		t.finished = true
		t.closedByParent = true

		return nil, OperationError(operation, errParentClosed)
	}

	return t.tx, nil
}

func (t *Transaction) finish(operation string) (*sql.Tx, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closedByParent {
		return nil, OperationError(operation, errParentClosed)
	}
	if t.finished {
		return nil, OperationError(operation, errTransactionDone)
	}
	if t.parent.isClosed() {
		t.finished = true
		t.closedByParent = true

		return nil, OperationError(operation, errParentClosed)
	}

	tx := t.tx
	t.tx = nil
	t.finished = true
	t.parent.removeTransaction(t)

	return tx, nil
}

func (t *Transaction) closeFromParent() error {
	tx := t.closeActive(true)
	if tx == nil {
		return nil
	}

	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}

	return nil
}

func (t *Transaction) closeActive(parent bool) *sql.Tx {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.finished {
		return nil
	}

	tx := t.tx
	t.tx = nil
	t.finished = true
	t.closedByParent = parent
	if !parent {
		t.parent.removeTransaction(t)
	}

	return tx
}
