package nestedset

import (
	"context"
	"database/sql"
	"os"
	"strings"

	type_scanner "github.com/moisespsena-go/type-scanner"
)

var nodeScanner = type_scanner.New(&Node{})

type Nested struct {
	TableName string
	Driver    Driver
}

func New(tableName string) *Nested {
	return &Nested{TableName: tableName, Driver: NewDefaultDriver()}
}

func (this Nested) Q(query string) string {
	return strings.ReplaceAll(query, "{TB}", this.TableName)
}

func (this Nested) Node(ctx context.Context, tx *sql.Tx, id ID) (n *Node, err error) {
	var r *sql.Rows
	q, args := this.Driver.ControlDetail(ctx, id)
	if r, err = tx.QueryContext(ctx, this.Q(q), args...); err != nil {
		return
	}
	defer r.Close()
	err = nodeScanner.Bulk(r, func(v interface{}) error {
		n = v.(*Node)
		return nil
	})
	return
}

func (this Nested) MaxRight(ctx context.Context, tx *sql.Tx) (max int64, err error) {
	q, args := this.Driver.MaxRight(ctx)
	var rows *sql.Rows
	if rows, err = tx.QueryContext(ctx, this.Q(q), args...); err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&max)
	}
	return
}

func (this Nested) FindID(ctx context.Context, tx *sql.Tx, lft, rgt int64, dst ID) (err error) {
	q, args := this.Driver.FindID(ctx, lft, rgt)
	var rows *sql.Rows
	if rows, err = tx.QueryContext(ctx, q, args...); err != nil {
		return
	}
	defer rows.Close()
	if !rows.Next() {
		err = os.ErrNotExist
		return
	}
	rows.Next()
	return rows.Scan(dst)
}
