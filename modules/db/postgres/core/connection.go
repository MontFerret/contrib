package core

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Connection is an opaque Postgres database handle exposed to Ferret.
type Connection struct {
	db     *sql.DB
	txs    map[*Transaction]struct{}
	mu     sync.Mutex
	id     uint64
	closed bool
}

// NewConnection creates a tracked Postgres connection handle.
func NewConnection(db *sql.DB) *Connection {
	return &Connection{
		db:  db,
		txs: make(map[*Transaction]struct{}),
		id:  newResourceID(),
	}
}

func (c *Connection) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	db, err := c.database("QUERY")
	if err != nil {
		return nil, err
	}

	return querySQL(ctx, "QUERY", db, q)
}

func (c *Connection) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return runtime.DefaultQueryOne(ctx, q, c.Query)
}

func (c *Connection) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return runtime.DefaultQueryCount(ctx, q, c.Query)
}

func (c *Connection) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return runtime.DefaultQueryExists(ctx, q, c.Query)
}

func (c *Connection) Begin(ctx context.Context) (*Transaction, error) {
	db, err := c.database("BEGIN")
	if err != nil {
		return nil, err
	}

	sqlTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, OperationError("BEGIN", err)
	}

	tx := NewTransaction(c, sqlTx)

	if err := c.addTransaction(tx); err != nil {
		_ = tx.Close()
		return nil, err
	}

	return tx, nil
}

func (c *Connection) Close() error {
	c.mu.Lock()

	if c.closed {
		c.mu.Unlock()

		return nil
	}

	c.closed = true
	txs := make([]*Transaction, 0, len(c.txs))

	for tx := range c.txs {
		txs = append(txs, tx)
	}

	c.txs = make(map[*Transaction]struct{})
	db := c.db
	c.mu.Unlock()

	errs := make([]error, 0)

	for _, tx := range txs {
		if err := tx.closeFromParent(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := db.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := errors.Join(errs...); err != nil {
		return OperationError("CLOSE", err)
	}

	return nil
}

func (c *Connection) ResourceID() uint64 {
	return c.id
}

func (c *Connection) String() string {
	return commonresource.Display("db.postgres.connection")
}

func (c *Connection) Hash() uint64 {
	return commonresource.Hash("db.postgres.connection", c.id)
}

func (c *Connection) Copy() runtime.Value {
	return c
}

func (c *Connection) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalDisplayJSON("db.postgres.connection")
}

func (c *Connection) database(operation string) (*sql.DB, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, OperationError(operation, errConnectionClosed)
	}

	return c.db, nil
}

func (c *Connection) addTransaction(tx *Transaction) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.txs[tx] = struct{}{}

		return nil
	}

	return OperationError("BEGIN", errConnectionClosed)
}

func (c *Connection) removeTransaction(tx *Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.txs, tx)
}

func (c *Connection) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}
