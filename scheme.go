package nestedset

import "strings"

type SchemeColumn struct {
	Name       string
	Definition string
}

func (this SchemeColumn) DDL(s Scheme) (r string) {
	return this.Name + " " + this.Definition
}

type SchemeColumns struct {
	ID,
	Lft,
	Rgt,
	Name,
	Depth,
	ParentID,
	DenyChildren SchemeColumn
	Other map[string]string
}

func (this SchemeColumns) DDL(s Scheme) (r []string) {
	r = append(r,
		this.ID.DDL(s),
		this.Lft.DDL(s),
		this.Rgt.DDL(s),
		this.Name.DDL(s),
		this.Depth.DDL(s),
		this.ParentID.DDL(s),
		this.DenyChildren.DDL(s))
	if this.Other != nil {
		for name, def := range this.Other {
			r = append(r, (SchemeColumn{name, def}).DDL(s))
		}
	}
	return
}

type SchemeIndex struct {
	Name        string
	Definition  func(s Scheme, i SchemeIndex) string
	Columns     []string
	ColumnsFunc func(s Scheme, i SchemeIndex) []string
	Unique      bool
}

func (this SchemeIndex) DDL(s Scheme) (r string) {
	r = "CREATE "
	if this.Unique {
		r += "UNIQUE "
	}
	r += "INDEX IF NOT EXISTS "
	var columns = this.Columns
	if this.ColumnsFunc != nil {
		columns = this.ColumnsFunc(s, this)
	}
	if this.Name == "" {
		r += "ix"
		if this.Unique {
			r += "u"
		}
		r += "_{TB}_"
		r += strings.Join(columns, "_")
	} else {
		r += this.Name
	}
	r += " ON {TB} "
	if this.Definition == nil {
		r += "(" + strings.Join(columns, ", ") + ")"
	} else {
		r += this.Definition(s, this)
	}
	return
}

type SchemeIndexes []SchemeIndex

func (this SchemeIndexes) DDL(s Scheme) (r []string) {
	for _, i := range this {
		r = append(r, i.DDL(s))
	}
	return
}

type Scheme struct {
	Columns SchemeColumns
	Indexes SchemeIndexes
}

func (this Scheme) DDL() []string {
	return append(this.Columns.DDL(this), this.Indexes.DDL(this)...)
}

var DefaultScheme = Scheme{
	Columns: SchemeColumns{
		ID:           SchemeColumn{"id", "INTEGER PRIMARY KEY AUTOINCREMENT"},
		Lft:          SchemeColumn{"lft", "INT NOT NULL"},
		Rgt:          SchemeColumn{"rgt", "INT NOT NULL"},
		Name:         SchemeColumn{"name", "VARCHAR(64) NOT NULL"},
		Depth:        SchemeColumn{"depth", "INT NOT NULL DEFAULT 0"},
		ParentID:     SchemeColumn{"parent_id", "BIGINT NOT NULL DEFAULT 0"},
		DenyChildren: SchemeColumn{"deny_children", "BOOLEAN NOT NULL DEFAULT FALSE"},
	},
	Indexes: []SchemeIndex{
		{Columns: []string{"name"}},
		{Columns: []string{"depth"}},
		{Columns: []string{"lft"}},
		{Columns: []string{"rgt"}},
		{ColumnsFunc: func(s Scheme, i SchemeIndex) []string {
			return []string{"parent_id", s.Columns.Name.Name}
		}, Unique: true},
	},
}
