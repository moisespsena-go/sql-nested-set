package nestedset

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"time"
)

func Dv(v interface{}) interface{} {
	if valuer, ok := v.(driver.Valuer); ok {
		v, _ = valuer.Value()
	}
	return v
}

func mv(v ID) interface{} {
	return v.MustValue()
}

type Epoch time.Time

func (this *Epoch) Scan(src interface{}) (err error) {
	if src != nil {
		switch t := src.(type) {
		case int64:
			*this = Epoch(time.Unix(t, 0))
		case int32:
			*this = Epoch(time.Unix(int64(t), 0))
		case string:
			return this.Scan([]byte(t))
		case []byte:
			var secs int64
			switch len(t) {
			case 4:
				secs = int64(binary.BigEndian.Uint32(t))
			case 8:
				secs = int64(binary.BigEndian.Uint64(t))
			default:
				return errors.New("bad size")
			}
			*this = Epoch(time.Unix(secs, 0))
		default:
			return errors.New("bad type")
		}
	}
	return
}
