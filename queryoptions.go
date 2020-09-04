package nestedset

import (
	"strconv"
	"strings"
	"text/scanner"
)

type SelectOptions struct {
	Args    []interface{}
	Scanner scanner.Scanner
	Columns []string
	Order   []struct {
		Name string
		Asc  bool
	}
	Limit, Offset int
}

func (this SelectOptions) Scaner(s scanner.Scanner) SelectOptions {
	this.Scanner = s
	return this
}

func (this SelectOptions) Arg(args ...interface{}) SelectOptions {
	this.Args = append(this.Args, args...)
	return this
}

func (this SelectOptions) Asc(v string) SelectOptions {
	this.Order = append(this.Order, struct {
		Name string
		Asc  bool
	}{Name: v, Asc: true})
	return this
}

func (this SelectOptions) Desc(v string) SelectOptions {
	this.Order = append(this.Order, struct {
		Name string
		Asc  bool
	}{Name: v})
	return this
}

func (this SelectOptions) Lim(limit int) SelectOptions {
	this.Limit = limit
	return this
}

func (this SelectOptions) Off(offset int) SelectOptions {
	this.Offset = offset
	return this
}

func (this SelectOptions) Select(column ...string) SelectOptions {
	this.Columns = append(this.Columns, column...)
	return this
}

func CompileQueryOptions(opts ...SelectOptions) (selec, order, limit string) {
	var opt SelectOptions
	for _, opt = range opts {
	}

	selec = strings.Join(opt.Columns, ", ")

	var orderby []string
	for _, ob := range opt.Order {
		var v = ob.Name + " "
		if ob.Asc {
			v += "ASC"
		} else {
			v += "DESC"
		}
		orderby = append(orderby, v)
	}
	order = strings.Join(orderby, ", ")

	if opt.Limit > 0 {
		limit = "LIMIT " + strconv.Itoa(opt.Limit) + " "
	}
	if opt.Offset > 0 {
		limit = "OFFSET " + strconv.Itoa(opt.Offset) + " "
	}
	return
}
