package nestedset

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (this Nested) Setup(ctx context.Context, db *sql.DB) (err error) {
	ddl := strings.Join(this.DDL(), ";\n")+";"
	fmt.Println(ddl)
	_, err = db.ExecContext(ctx, ddl)
	return
}
