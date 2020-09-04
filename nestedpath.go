package nestedset

import (
	"context"
	"database/sql"
	"os"
)

func (this *Nested) PathOf(ctx context.Context, tx *sql.Tx, id ID) (pth string, err error) {
	q, args := this.Driver.PathOf(ctx, id)
	var rows *sql.Rows
	if rows, err = tx.Query(this.Q(q), args...); err != nil {
		return
	}
	if !rows.Next() {
		err = os.ErrNotExist
		return
	}
	err = rows.Scan(&pth)
	return
}

func (this *Nested) PathToID(ctx context.Context, tx *sql.Tx, pth string, id ID) (err error) {
	q, args := this.Driver.PathToID(ctx, pth)
	var rows *sql.Rows
	if rows, err = tx.Query(this.Q(q), args...); err != nil {
		return
	}
	if !rows.Next() {
		return os.ErrNotExist
	}
	err = rows.Scan(id)
	return
}
