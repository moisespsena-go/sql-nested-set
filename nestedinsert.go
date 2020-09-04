package nestedset

import (
	"context"
	"database/sql"
	"errors"
	"os"
)

func (this *Nested) InsertRoot(ctx context.Context, tx *sql.Tx, id ID, name string, args ...interface{}) (node *Node, err error) {
	return this.Insert(ctx, tx, id, nil, name, args...)
}

func (this *Nested) Insert(ctx context.Context, tx *sql.Tx, id, toId ID, name string, args ...interface{}) (node *Node, err error) {
	if toId == nil {
		var max int64
		if max, err = this.MaxRight(ctx, tx); err != nil {
			return
		}
			max++
		node = &Node{
			ID:  id,
			Lft: max,
			Rgt: max + 1,
		}
	} else {
		var to *Node
		if to, err = this.Node(ctx, tx, toId); err != nil {
			if os.IsNotExist(err) {
				err = errors.New("nested.insert: destination node does not exists.")
			}
			return
		} else if to.DenyChindren {
			err = errors.New("nested.insert: destination does not accept children.")
			return
		}

		qs := this.Driver.MoveAftersRight(ctx, to.Rgt, 2)
		if _, err = tx.ExecContext(ctx, this.Q(qs.Lefts.Query), qs.Lefts.Args...); err != nil {
			return
		}
		if _, err = tx.ExecContext(ctx, this.Q(qs.Rigths.Query), qs.Rigths.Args...); err != nil {
			return
		}
		node = &Node{
			ID:       id,
			ParentID: toId,
			Lft:      to.Rgt,
			Rgt:      to.Rgt + 1,
			Depth:    to.Depth + 1,
		}
	}

	defer func() {
		if err != nil {
			node = nil
		}
	}()

	q, args := this.Driver.Insert(ctx, node, name, args...)
	if _, err = tx.ExecContext(ctx, this.Q(q), args...); err != nil {
		return
	}

	if node.ID == nil {
		err = this.FindID(ctx, tx, node.Lft, node.Rgt, node.ID)
	}
	return
}
