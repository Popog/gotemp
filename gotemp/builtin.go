package gotemp

import (
	"fmt"
	"reflect"
	"regexp"
	"text/template"
)

var builtins = template.FuncMap{
	"index":        index,
	"rindex":       builtinReverseIndex,
	"filter":       builtinFilter,
	"regexpfilter": builtinRegexpFilter,
	"append":       builtinAppend,
	"rappend":      builtinReverseAppend,
	"appendslice":  builtinAppendSlice,
	"prepend":      builtinPrepend,
	"prependslice": builtinPrependSlice,
	"set":          builtinSet,
	"rset":         builtinReverseSet,
	"errorf":       builtinErrorf,
	"nop":          builtinNop,
}

func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
		if v.Kind() == reflect.Interface && v.NumMethod() > 0 {
			break
		}
	}
	return v, false
}
func index(item interface{}, indices ...interface{}) (interface{}, error) {
	v := reflect.ValueOf(item)
	for _, i := range indices {
		index := reflect.ValueOf(i)
		var isNil bool
		if v, isNil = indirect(v); isNil {
			return nil, fmt.Errorf("index of nil pointer")
		}
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			var x int64
			switch index.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				x = index.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				x = int64(index.Uint())
			default:
				return nil, fmt.Errorf("cannot index slice/array with type %s", index.Type())
			}
			if x < 0 || x >= int64(v.Len()) {
				return nil, fmt.Errorf("index out of range: %d", x)
			}
			v = v.Index(int(x))
		case reflect.Map:
			if !index.Type().AssignableTo(v.Type().Key()) {
				return nil, fmt.Errorf("%s is not index type for %s", index.Type(), v.Type())
			}
			if x := v.MapIndex(index); x.IsValid() {
				v = x
			} else {
				v = reflect.Zero(v.Type().Elem())
			}
		default:
			return nil, fmt.Errorf("can't index item of type %s", index.Type())
		}
	}
	return v.Interface(), nil
}

func builtinReverseIndex(indices_item ...interface{}) (interface{}, error) {
	if len(indices_item) == 0 {
		return nil, fmt.Errorf("no items passed")
	}

	item_index := len(indices_item) - 1
	return index(indices_item[item_index], indices_item[:item_index]...)
}

func builtinFilter(item interface{}, indices ...interface{}) (interface{}, error) {
	v := reflect.ValueOf(item)

	v, isNil := indirect(v)
	if isNil {
		return nil, fmt.Errorf("filter of nil pointer")
	} else if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("can't filter item of type %s", v.Type())
	}

	m := reflect.MakeMap(v.Type())
	for _, i := range indices {
		index := reflect.ValueOf(i)

		if x := v.MapIndex(index); x.IsValid() {
			m.SetMapIndex(index, x)
		}
	}
	return m.Interface(), nil
}

func builtinRegexpFilter(item interface{}, regex string) (interface{}, error) {
	v := reflect.ValueOf(item)

	v, isNil := indirect(v)
	if isNil {
		return nil, fmt.Errorf("regexpfilter of nil pointer")
	} else if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("can't regexpfilter item of type %s", v.Type())
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	m := reflect.MakeMap(v.Type())

	for _, key := range v.MapKeys() {
		if r.MatchString(key.String()) {
			m.SetMapIndex(key, v.MapIndex(key))
		}
	}

	return m.Interface(), nil
}

func builtinAppend(s interface{}, x ...interface{}) (interface{}, error) {
	sval := reflect.ValueOf(s)
	if sval.Kind() != reflect.Slice {
		return nil, fmt.Errorf("can't append to item of type %s", sval.Type())
	}

	xval := make([]reflect.Value, len(x))
	for i, e := range x {
		xval[i] = reflect.ValueOf(e)
	}

	return reflect.Append(sval, xval...).Interface(), nil
}

func builtinReverseAppend(x_s ...interface{}) (interface{}, error) {
	if len(x_s) == 0 {
		return nil, fmt.Errorf("no items passed")
	}

	s_index := len(x_s) - 1
	return builtinAppend(x_s[s_index], x_s[:s_index]...)
}

func builtinAppendSlice(s interface{}, t interface{}) (interface{}, error) {
	return reflect.AppendSlice(reflect.ValueOf(s), reflect.ValueOf(t)).Interface(), nil
}

func builtinPrepend(s interface{}, x ...interface{}) (interface{}, error) {
	sval := reflect.ValueOf(s)
	if sval.Kind() != reflect.Slice {
		return nil, fmt.Errorf("can't index item of type %s", sval.Type())
	}

	result := reflect.MakeSlice(sval.Type(), 0, len(x)+sval.Len())

	xval := make([]reflect.Value, len(x))
	for i, e := range x {
		xval[i] = reflect.ValueOf(e)
	}

	return reflect.AppendSlice(reflect.Append(result, xval...), sval).Interface(), nil
}

func builtinPrependSlice(s interface{}, t interface{}) (interface{}, error) {
	return builtinAppendSlice(t, s)
}

func builtinSet(item interface{}, key interface{}, val interface{}) (interface{}, error) {
	ritem := reflect.ValueOf(item)
	rkey := reflect.ValueOf(key)
	rval := reflect.ValueOf(val)

	switch ritem.Kind() {
	case reflect.Array, reflect.Slice:
		var index int64
		switch rkey.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			index = rkey.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			index = int64(rkey.Uint())
		default:
			return nil, fmt.Errorf("cannot index slice/array with type %s", rkey.Type())
		}
		if index < 0 || index >= int64(ritem.Len()) {
			return nil, fmt.Errorf("index out of range: %d", index)
		}
		if !rval.Type().AssignableTo(ritem.Type().Elem()) {
			return nil, fmt.Errorf("%s is not elem type for %s", rval.Type(), ritem.Type())
		}

		ritem.Index(int(index)).Set(rval)
	case reflect.Map:
		if !rkey.Type().AssignableTo(ritem.Type().Key()) {
			return nil, fmt.Errorf("%s is not index type for %s", rkey.Type(), ritem.Type())
		}
		if !rval.Type().AssignableTo(ritem.Type().Elem()) {
			return nil, fmt.Errorf("%s is not elem type for %s", rval.Type(), ritem.Type())
		}

		ritem.SetMapIndex(rkey, rval)
	default:
		return nil, fmt.Errorf("can't index item of type %s", rkey.Type())
	}

	return ritem.Interface(), nil
}

func builtinReverseSet(key interface{}, val interface{}, item interface{}) (interface{}, error) {
	return builtinSet(item, key, val)
}

func builtinErrorf(format string, a ...interface{}) (interface{}, error) {
	return nil, fmt.Errorf(format, a...)
}

func builtinNop(x interface{}) interface{} { return x }
