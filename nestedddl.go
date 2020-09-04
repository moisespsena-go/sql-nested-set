package nestedset

func (this *Nested) DDL() []string {
	sql := this.Driver.DDL()
	for i, v := range sql {
		sql[i] = this.Q(v)
	}
	return sql
}
