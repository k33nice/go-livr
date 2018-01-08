package test

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

func deepValueEqual(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := unsafe.Pointer(v1.UnsafeAddr())
		addr2 := unsafe.Pointer(v2.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Slice:
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Map:
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !deepValueEqual(v1.MapIndex(k), v2.MapIndex(k), visited, depth+1) {
				return false
			}
		}
		return true
	default:
		switch v1.Kind() {
		case reflect.Bool:
			switch v2.Kind() {
			case reflect.Bool:
				return v2.Bool() == v1.Bool()
			case reflect.String:
				if b, err := strconv.ParseBool(v2.String()); err == nil {
					return v1.Bool() == b
				}
				return false
			case reflect.Ptr:
				fmt.Printf("v2.Addr() = %+v\n", v2.Addr())
			}
		case reflect.String:
			switch v2.Kind() {
			case reflect.Bool:
				return strconv.FormatBool(v2.Bool()) == v1.String()
			case reflect.String:
				return v1.String() == v2.String()
			case reflect.Float64:
				return strconv.FormatFloat(v2.Float(), 'f', -1, 64) == v1.String()
			}
		case reflect.Float64:
			switch v2.Kind() {
			case reflect.Float64:
				return v1.Float() == v2.Float()
			case reflect.String:
				return strconv.FormatFloat(v1.Float(), 'f', -1, 64) == v2.String()
			case reflect.Ptr:
				fmt.Printf("v2.Addr() = %+v\n", v2.Addr())
			}
		}
		return false
	}
}

// JSONDuckEqual reports if two map[string]interface{} value are equal.
// On equality skip type checks and try to cast leaf to the same basic type.
func JSONDuckEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}
	return deepValueEqual(v1, v2, make(map[visit]bool), 0)
}
