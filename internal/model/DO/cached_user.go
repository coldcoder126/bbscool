package DO

import (
	"reflect"
	"strconv"
	"strings"
)

// CachedUser 放入缓存中的用户信息，用于鉴权等
type CachedUser struct {
	Username string
	Status   int8
	School   []string
}

func (cu *CachedUser) GetValByFieldName(field string) string {
	t := reflect.ValueOf(*cu)
	v := t.FieldByName(field)
	k := v.Kind()
	switch k {
	case reflect.String:
		return v.String()
	case reflect.Int:
		return strconv.Itoa(int(v.Int()))
	case reflect.Slice:
		temp := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			temp = append(temp, v.Index(i).String())
		}
		return strings.Join(temp, ",")
	}
	return ""
}
