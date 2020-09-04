package nestedset

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

func NewID(value interface{}) ID {
	var id Id
	if valuer, ok := value.(driver.Valuer); ok {
		id.Valuer = valuer.Value
	} else {
		id.Valuer = func() (v driver.Value, e error) {
			return value, nil
		}
	}

	if scanner, ok := value.(sql.Scanner); ok {
		id.Scanner = scanner.Scan
	} else {
		id.Scanner = func(v interface{}) (err error) {
			defer func() {
				if r := recover(); r != nil {
					switch et := r.(type) {
					case error:
						err = et
					case string:
						err = errors.New(et)
					default:
						err = errors.New(fmt.Sprint(et))
					}
				}
			}()
			i := binary.BigEndian.Uint64(v.([]byte))
			reflect.Indirect(reflect.ValueOf(value)).Set(reflect.ValueOf(i))
			return
		}
	}

	id.Rawer = func() interface{} {
		return value
	}
	return id
}
