package nestedset

import "context"

type Driver interface {
	FindID(ctx context.Context, lft, rgt int64) (query string, args []interface{})
	Insert(ctx context.Context, node *Node, name string, iargs ...interface{}) (query string, args []interface{})
	Detail(ctx context.Context, id ID) (query string, args []interface{})
	ControlDetail(ctx context.Context, id ID) (query string, args []interface{})
	Delete(ctx context.Context, id ID, lft, rgt int64, iargs ...interface{}) (query string, args []interface{})
	PathOf(ctx context.Context, id ID) (query string, args []interface{})
	PathToID(ctx context.Context, pth string) (query string, args []interface{})
	SelfSubTree(ctx context.Context, root, sub ID) (query string, args []interface{})
	SetParentID(ctx context.Context, id, parentID ID) (query string, args []interface{})
	TreeTempSet(ctx context.Context, node *Node, depthDiff *int32) (query string, args []interface{})
	TreeTempUnset(ctx context.Context, node *Node, newLeft int64, depthDiff *int32) (query string, args []interface{})
	MoveAftersLeft(ctx context.Context, start, length int64) struct {
		Lefts, Rigths struct {
			Query string
			Args  []interface{}
		}
	}
	MoveAftersRight(ctx context.Context, start, length int64) struct {
		Lefts, Rigths struct {
			Query string
			Args  []interface{}
		}
	}
	MaxRight(ctx context.Context) (query string, args []interface{})
	ValueToID(ctx context.Context, nodeId interface{}) ID
	Scheme() *Scheme
	DDL() []string
}
