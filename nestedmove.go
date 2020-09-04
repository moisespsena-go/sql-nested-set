package nestedset

import (
	"context"
	"database/sql"
	"errors"
	"os"
)

func (this *Nested) MoveToRoot(ctx context.Context, tx *sql.Tx, id ID) (err error) {
	return this.Move(ctx, tx, id, nil)
}

func (this *Nested) moveAftersLeft(ctx context.Context, tx *sql.Tx, start, width int64) (err error) {
	qs := this.Driver.MoveAftersLeft(ctx, start, width)
	if _, err = tx.ExecContext(ctx, this.Q(qs.Lefts.Query), qs.Lefts.Args...); err != nil {
		return
	}
	if _, err = tx.ExecContext(ctx, this.Q(qs.Rigths.Query), qs.Rigths.Args...); err != nil {
		return
	}
	return
}

func (this *Nested) moveAftersRight(ctx context.Context, tx *sql.Tx, start, width int64) (err error) {
	qs := this.Driver.MoveAftersRight(ctx, start, width)
	if _, err = tx.ExecContext(ctx, this.Q(qs.Lefts.Query), qs.Lefts.Args...); err != nil {
		return
	}
	if _, err = tx.ExecContext(ctx, this.Q(qs.Rigths.Query), qs.Rigths.Args...); err != nil {
		return
	}
	return
}

func (this *Nested) Move(ctx context.Context, tx *sql.Tx, id, toId ID) (err error) {
	var rows *sql.Rows
	if toId != nil {
		q, args := this.Driver.SelfSubTree(ctx, id, toId)

		if rows, err = tx.QueryContext(ctx, this.Q(q), args...); err != nil {
			return
		} else if rows.Next() {
			return errors.New("nested: bad move tree into self subtree.")
		}
	}

	var node *Node

	if node, err = this.Node(ctx, tx, id); err != nil {
		if os.IsNotExist(err) {
			return errors.New("nested.move: node does not exist.")
		}
		return
	}

	q, args := this.Driver.TreeTempSet(ctx, node, nil)
	if _, err = tx.ExecContext(ctx, this.Q(q), args...); err != nil {
		return
	}
	width := node.Rgt - node.Lft + 1

	if err = this.moveAftersLeft(ctx, tx, node.Rgt, width); err != nil {
		return
	}

	var (
		newLeft   int64
		depthDiff int32
	)

	if toId == nil {
		var max int64
		if max, err = this.MaxRight(ctx, tx); err != nil {
			return
		}
		newLeft = max + 1
		depthDiff = -node.Depth
	} else {
		var to *Node
		if to, err = this.Node(ctx, tx, toId); err != nil {
			if os.IsNotExist(err) {
				return errors.New("nested.move: new parent node does not exist.")
			}
			return
		} else if to.DenyChindren {
			err = errors.New("nested.move: destination does not accept children.")
			return
		}
		newLeft = to.Rgt
		depthDiff = to.Depth - node.Depth + 1
	}

	if toId != nil {
		if err = this.moveAftersRight(ctx, tx, newLeft, width); err != nil {
			return
		}
	}

	q, args = this.Driver.TreeTempUnset(ctx, node, newLeft, &depthDiff)
	if _, err = tx.ExecContext(ctx, this.Q(q), args...); err != nil {
		return
	}

	if toId != nil {
		q, args = this.Driver.SetParentID(ctx, id, toId)
		if _, err = tx.ExecContext(ctx, this.Q(q), args...); err != nil {
			return
		}
	}
	return nil
}
