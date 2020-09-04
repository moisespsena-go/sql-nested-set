package nestedset

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"xorm.io/builder"
)

type DefaultDriver struct {
	InsertColumns []string
	NewID         func() ID
	scheme        *Scheme
}

func NewDefaultDriver() *DefaultDriver {
	s := DefaultScheme
	return &DefaultDriver{scheme: &s}
}

func (this DefaultDriver) Insert(ctx context.Context, node *Node, name string, iargs ...interface{}) (query string, args []interface{}) {
	columns := append([]string{
		this.scheme.Columns.Name.Name,
		"lft",
		"rgt",
		"depth",
		"parent_id",
	}, this.InsertColumns...)

	if node.ID == nil && this.NewID != nil {
		node.ID = this.NewID()
	}

	if node.ID != nil {
		columns = append([]string{"id"}, columns...)
		args = append(args, mv(node.ID))
	}

	args = append(args, name, node.Lft, node.Rgt, node.Depth, node.ParentID)
	query = "INSERT INTO {TB} (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Repeat("?,", len(columns)-1) + "?)"
	args = append(args, iargs...)
	return
}

func (this DefaultDriver) Detail(ctx context.Context, id ID) (query string, args []interface{}) {
	return "SELECT parent_id, " + this.scheme.Columns.Name.Name + " FROM {TB} WHERE id = ?", []interface{}{mv(id)}
}

func (this DefaultDriver) ControlDetail(ctx context.Context, id ID) (query string, args []interface{}) {
	return "SELECT depth, lft, rgt, " + this.scheme.Columns.DenyChildren.Name + " FROM {TB} WHERE id = ?", []interface{}{mv(id)}
}

func (DefaultDriver) Delete(ctx context.Context, id ID, lft, rgt int64, iargs ...interface{}) (query string, args []interface{}) {
	return "DELETE FROM {TB} WHERE lft BETWEEN ? AND ?", []interface{}{lft, rgt}
}

func (this DefaultDriver) Children(ctx context.Context, id ID, b builder.Builder) (query string, args []interface{}, err error) {
	b.From("{TB}").
		SelectAdd("id", "depth", "lft", "rgt", this.scheme.Columns.DenyChildren.Name).
		Where(builder.Eq{"id": mv(id)})
	args = []interface{}{mv(id)}
	return b.ToSQL()
}

func (this DefaultDriver) BuildPathQuery(b builder.Builder, root *Node) builder.Builder {
	nameColumn := this.scheme.Columns.DenyChildren.Name
	filter := builder.Select("id", "MAX(LENGTH(path))", "path").
		From("node_path").
		GroupBy("id")

	if root != nil {
		filter.Where(builder.And(
			builder.Gt{"lft": root.Lft},
			builder.Lt{"rgt": root.Rgt},
		))
	}

	b.SelectAdd("SUBSTR(paths.path, 1, length(paths.path)-LENGTH({TB}."+nameColumn+")-1) AS dir").
		From(filter,
			"paths").
		With(&builder.With{
			Name:      "node_path",
			Columns:   []string{"id", "parent_id", "path"},
			Recursive: true,
			As: builder.Select("id", "parent_id", "name").
				From("{TB}").
				Union("ALL",
					builder.Select("child.id", "child.parent_id", "path || '/' || child.name").
						From("{TB}", "child").
						Join("INNER", "node_path AS parent", "parent.id = child.parent_id"),
				),
		}).
		Join("INNER", "{TB}", "{TB}.id = paths.id").
		OrderBy("paths.path")
	return b
}

func (this DefaultDriver) PathOf(ctx context.Context, id ID) (query string, args []interface{}) {
	nameColumn := this.scheme.Columns.DenyChildren.Name
	query, args, _ = builder.Select("path").
		From("{TB}s_path").
		With(&builder.With{
			Name:      "{TB}s_path",
			Columns:   []string{"id", "parent_id", "path"},
			Recursive: true,
			As: builder.Select("id", "parent_id", nameColumn).
				From("{TB}").
				Where(builder.Eq{"id": mv(id)}).
				Union("ALL",
					builder.Select(
						"parent.id",
						"parent.parent_id",
						"parent."+nameColumn+" || '/' || child.path",
					).
						From("{TB} as parent").
						Join("INNER", "{TB}s_path as child", "ON parent.id = child.parent_id")),
		}).
		Where(builder.Eq{"parent_id": nil}).ToSQL()
	return
}

func (this DefaultDriver) BuildPathQueryFromDepth(depth int32) (column, from, join string) {
	var joins []string
	names := make([]string, depth+1)
	names[depth] = "n0." + this.scheme.Columns.Name.Name
	var s = func(i int32) string {
		return strconv.Itoa(int(i))
	}
	for i := int32(1); i <= depth; i++ {
		a, b := "n"+s(i-1), "n"+s(i)
		joins = append(joins, "JOIN {TB} "+b+" ON "+b+".id = "+a+".parent_id")
		names[depth-i] = b + "." + this.scheme.Columns.Name.Name
	}
	column = strings.Join(names, " || '/' || ")
	from = "n0"
	join = strings.Join(joins, "\n")
	return
}

func (this DefaultDriver) PathToID(ctx context.Context, pth string) (query string, args []interface{}) {
	parts := strings.Split(pth, "/")
	_, from, join := this.BuildPathQueryFromDepth(int32(len(parts)) - 1)
	query = "SELECT " + from + ".id FROM {TB} as " + from + "\n" +
		join + "\n" +
		"WHERE "

	var where []string
	for i, p := range parts {
		where = append(where, "n"+strconv.Itoa(len(parts)-i-1)+".name = ?")
		args = append(args, p)
	}
	query += strings.Join(where, " AND ")
	return
}

func (DefaultDriver) FindID(ctx context.Context, lft, rgt int64) (query string, args []interface{}) {
	return "SELECT id FROM {TB} WHERE lft = ? AND rgt = ?", []interface{}{lft, rgt}
}

func (DefaultDriver) SelfSubTree(ctx context.Context, root, sub ID) (query string, args []interface{}) {
	return "SELECT 1 " +
		"FROM {TB} as r, {TB} as s " +
		"WHERE r.id = ? " +
		"AND s.id = ? " +
		"AND s.lft > r.lft " +
		"AND s.rgt < r.rgt", []interface{}{mv(root), mv(sub)}
}

func (this DefaultDriver) SetParentID(ctx context.Context, id, parentID ID) (query string, args []interface{}) {
	return "UPDATE {TB} SET parent_id = ? WHERE id = ?", []interface{}{mv(parentID), mv(id)}
}

func (DefaultDriver) TreeTempSet(ctx context.Context, node *Node, depthDiff *int32) (query string, args []interface{}) {
	query = "UPDATE {TB} SET rgt = rgt - :rgt, lft = lft - :rgt"
	args = []interface{}{
		sql.Named("lft", node.Lft),
		sql.Named("rgt", node.Rgt),
	}

	if depthDiff != nil {
		query += ", depth = depth + :depthDiff"
		args = append(args, sql.Named("depthDiff", *depthDiff))
	}

	query += " WHERE lft >= :lft AND rgt <= :rgt"
	return
}

func (DefaultDriver) TreeTempUnset(ctx context.Context, node *Node, newLeft int64, depthDiff *int32) (query string, args []interface{}) {
	query = "UPDATE {TB} SET rgt = rgt + :len, lft = lft + :len"
	args = []interface{}{
		sql.Named("len", node.Rgt-node.Lft+newLeft),
		sql.Named("lft", 0-node.Rgt-node.Lft),
		sql.Named("rgt", 0),
	}

	if depthDiff != nil {
		query += ", depth = depth + :depthDiff"
		args = append(args, sql.Named("depthDiff", *depthDiff))
	}

	query += " WHERE lft >= :lft AND rgt <= :rgt"
	return
}

func (DefaultDriver) MoveTreeRight(ctx context.Context, length, lft, rgt int64) (query string, args []interface{}) {
	return "UPDATE {TB} SET rgt = rgt - :len, lft = lft - :len WHERE lft >= :lft AND rgt <= :rgt",
		[]interface{}{
			sql.Named("len", length),
			sql.Named("lft", lft),
			sql.Named("rgt", rgt),
		}
}

func (DefaultDriver) MoveAftersLeft(ctx context.Context, start, length int64) struct {
	Lefts, Rigths struct {
		Query string
		Args  []interface{}
	}
} {
	args := []interface{}{
		sql.Named("len", length),
		sql.Named("start", start),
	}
	return struct {
		Lefts, Rigths struct {
			Query string
			Args  []interface{}
		}
	}{
		struct {
			Query string
			Args  []interface{}
		}{
			"UPDATE {TB} SET rgt = rgt - :len WHERE rgt > :start", args,
		},
		struct {
			Query string
			Args  []interface{}
		}{"UPDATE {TB} SET lft = lft - :len WHERE lft > :start", args},
	}
}

func (DefaultDriver) MoveAftersRight(ctx context.Context, start, length int64) struct {
	Lefts, Rigths struct {
		Query string
		Args  []interface{}
	}
} {
	args := []interface{}{
		sql.Named("len", length),
		sql.Named("start", start),
	}
	return struct {
		Lefts, Rigths struct {
			Query string
			Args  []interface{}
		}
	}{
		struct {
			Query string
			Args  []interface{}
		}{
			"UPDATE {TB} SET rgt = rgt + :len WHERE rgt >= :start", args,
		},
		struct {
			Query string
			Args  []interface{}
		}{"UPDATE {TB} SET lft = lft + :len WHERE lft > :start", args},
	}
}

func (DefaultDriver) ValueToID(ctx context.Context, value interface{}) ID {
	return NewID(value)
}

func (DefaultDriver) MaxRight(ctx context.Context) (query string, args []interface{}) {
	return "SELECT (CASE WHEN m IS NULL THEN 0 ELSE m END) as m FROM (SELECT max(rgt) as m FROM {TB})", nil
}

func (this DefaultDriver) Scheme() *Scheme {
	return this.scheme
}

func (this DefaultDriver) DDL() []string {
	return this.DDLScheme(*this.scheme)
}

func (this DefaultDriver) DDLScheme(s Scheme) []string {
	return append([]string{
		"CREATE TABLE IF NOT EXISTS {TB} (\n  " +
			strings.Join(s.Columns.DDL(s), ",\n  ") +
			"\n" + ")",
	},
		s.Indexes.DDL(s)...)
}
