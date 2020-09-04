package nestedset

import (
	"context"
	"database/sql"
	"errors"
	"os"
)

func (this *Nested) Delete(ctx context.Context, tx *sql.Tx, recursive bool, id ID, args ...interface{}) (err error) {
	var node *Node

	if node, err = this.Node(ctx, tx, id); err != nil {
		if os.IsNotExist(err) {
			return errors.New("nested: node does not exist")
		}
		return
	}

	if !recursive && node.Rgt-node.Lft > 1 {
		return errors.New("nested: node is tree")
	}

	q, args := this.Driver.Delete(ctx, id, node.Lft, node.Rgt, args...)
	if _, err = tx.ExecContext(ctx, this.Q(q), args...); err != nil {
		return
	}

	return this.moveAftersLeft(ctx, tx, node.Rgt, node.Rgt-node.Lft+1)
}

func (this *Nested) DeleteRecursive(ctx context.Context, tx *sql.Tx, id ID, args ...interface{}) (err error) {
	return this.Delete(ctx, tx, true, id, args...)
}

func (this *Nested) DeleteOne(ctx context.Context, tx *sql.Tx, id ID, args ...interface{}) (err error) {
	return this.Delete(ctx, tx, false, id, args...)
}
