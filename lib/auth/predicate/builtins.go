/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package predicate

import (
	"reflect"
	"strings"

	"github.com/gravitational/trace"
	"github.com/vulcand/predicate"
)

func getIdentifier(obj any, selectors []string) (any, error) {
	for _, s := range selectors {
		if obj == nil || reflect.ValueOf(obj).IsNil() {
			return nil, trace.BadParameter("cannot take field of nil")
		}

		if m, ok := obj.(map[string]any); ok {
			obj = m[s]
			continue
		}

		val := reflect.ValueOf(obj)
		ty := reflect.TypeOf(obj)
		if ty.Kind() == reflect.Interface || ty.Kind() == reflect.Ptr {
			val = reflect.ValueOf(obj).Elem()
			ty = val.Type()
		}

		if ty.Kind() == reflect.Struct {
			for i := 0; i < ty.NumField(); i++ {
				tagValue := ty.Field(i).Tag.Get("json")
				parts := strings.Split(tagValue, ",")
				if parts[0] == s {
					obj = val.Field(i).Interface()
					break
				}
			}

			continue
		}

		return nil, trace.BadParameter("cannot take field of type: %T", obj)
	}

	return obj, nil
}

func getProperty(m any, k any) (any, error) {
	switch mT := m.(type) {
	case map[string]any:
		kS, ok := k.(string)
		if !ok {
			return nil, trace.BadParameter("unsupported key type: %T", k)
		}

		return mT[kS], nil
	default:
		return nil, trace.BadParameter("cannot take property of type: %T", m)
	}
}

func builtinOpEquals(a, b any) predicate.BoolPredicate {
	return func() bool { return reflect.DeepEqual(a, b) }
}

func builtinOpLT(a, b any) predicate.BoolPredicate {
	return func() bool {
		if reflect.TypeOf(a) != reflect.TypeOf(b) {
			return false
		}

		switch aT := a.(type) {
		case string:
			return aT < b.(string)
		case int:
			return aT < b.(int)
		case float32:
			return aT < b.(float32)
		default:
			return false
		}
	}
}

func builtinOpGT(a, b any) predicate.BoolPredicate {
	return builtinOpLT(b, a)
}

func builtinOpLE(a, b any) predicate.BoolPredicate {
	return func() bool {
		if reflect.TypeOf(a) != reflect.TypeOf(b) {
			return false
		}

		switch aT := a.(type) {
		case string:
			return aT <= b.(string)
		case int:
			return aT <= b.(int)
		case float32:
			return aT <= b.(float32)
		default:
			return false
		}
	}
}

func builtinOpGE(a, b any) predicate.BoolPredicate {
	return builtinOpLE(b, a)
}

func builtinAdd(a, b any) (any, error) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return nil, trace.BadParameter("cannot add types: %T and %T", a, b)
	}

	switch aT := a.(type) {
	case string:
		return aT + b.(string), nil
	case int:
		return aT + b.(int), nil
	case float32:
		return aT + b.(float32), nil
	default:
		return nil, trace.BadParameter("add unsupported for type type: %T", a)
	}
}

func builtinSub(a, b any) (any, error) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return nil, trace.BadParameter("cannot sub types: %T and %T", a, b)
	}

	switch aT := a.(type) {
	case int:
		return aT - b.(int), nil
	case float32:
		return aT - b.(float32), nil
	default:
		return nil, trace.BadParameter("sub unsupported for type type: %T", a)
	}
}

func builtinMul(a, b any) (any, error) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return nil, trace.BadParameter("cannot mul types: %T and %T", a, b)
	}

	switch aT := a.(type) {
	case int:
		return aT * b.(int), nil
	case float32:
		return aT * b.(float32), nil
	default:
		return nil, trace.BadParameter("mul unsupported for type type: %T", a)
	}
}

func builtinDiv(a, b any) (any, error) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return nil, trace.BadParameter("cannot div types: %T and %T", a, b)
	}

	switch aT := a.(type) {
	case int:
		return aT / b.(int), nil
	case float32:
		return aT / b.(float32), nil
	default:
		return nil, trace.BadParameter("div unsupported for type type: %T", a)
	}
}

func builtinXor(a, b any) (any, error) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return nil, trace.BadParameter("cannot xor types: %T and %T", a, b)
	}

	switch aT := a.(type) {
	case int:
		return aT ^ b.(int), nil
	case bool:
		return aT != b.(bool), nil
	default:
		return nil, trace.BadParameter("xor unsupported for type type: %T", a)
	}
}

func builtinSplit(a, b any) (any, error) {
	aS, ok := a.(string)
	if !ok {
		return nil, trace.BadParameter("cannot split type: %T", a)
	}

	bS, ok := b.(string)
	if !ok {
		return nil, trace.BadParameter("cannot split by type: %T", b)
	}

	return strings.Split(aS, bS), nil
}

func builtinUpper(a any) (any, error) {
	aS, ok := a.(string)
	if !ok {
		return nil, trace.BadParameter("upper not valid for type: %T", a)
	}

	return strings.ToUpper(aS), nil
}

func builtinLower(a any) (any, error) {
	aS, ok := a.(string)
	if !ok {
		return nil, trace.BadParameter("lower not valid for type: %T", a)
	}

	return strings.ToLower(aS), nil
}

func builtinContains(a, b any) (any, error) {
	var bS string
	if bT, ok := b.(string); ok {
		bS = bT
	} else {
		return nil, trace.BadParameter("cannot check if type: %T contains type: %T", a, b)
	}

	switch aT := a.(type) {
	case string:
		return strings.Contains(aT, bS), nil
	case []string:
		for _, s := range aT {
			if s == bS {
				return true, nil
			}
		}

		return false, nil
	default:
		return nil, trace.BadParameter("contains not valid for type: %T", a)
	}
}

func builtinFirst(a any) (any, error) {
	switch aT := a.(type) {
	case []string:
		if len(aT) == 0 {
			return nil, nil
		}

		return aT[0], nil
	default:
		return nil, trace.BadParameter("first not valid for type: %T", a)
	}
}

func builtinAppend(a, b any) (any, error) {
	var bS string
	if bT, ok := b.(string); ok {
		bS = bT
	} else {
		return nil, trace.BadParameter("cannot append type %T", b)
	}

	switch aT := a.(type) {
	case []string:
		return append(aT, bS), nil
	default:
		return nil, trace.BadParameter("append not valid for type: %T", a)
	}
}

func builtinArray(elements ...any) (any, error) {
	arr := make([]string, len(elements))
	for i, e := range elements {
		s, ok := e.(string)
		if !ok {
			return nil, trace.BadParameter("cannot create array with element type %T", e)
		}

		arr[i] = s
	}

	return arr, nil
}

func builtinReplace(in, match, with any) (any, error) {
	matchS, ok := match.(string)
	if !ok {
		return nil, trace.BadParameter("cannot replace with non-string match of type %T", match)
	}

	withS, ok := with.(string)
	if !ok {
		return nil, trace.BadParameter("cannot replace with non-string with of type %T", with)
	}

	switch inT := in.(type) {
	case string:
		return strings.Replace(inT, matchS, withS, -1), nil
	case []string:
		for i, s := range inT {
			if s == matchS {
				inT[i] = withS
			}
		}

		return inT, nil
	default:
		return nil, trace.BadParameter("replace not valid for type: %T", in)
	}
}

func builtinLen(a any) (any, error) {
	switch aT := a.(type) {
	case string:
		return len(aT), nil
	case []string:
		return len(aT), nil
	default:
		return nil, trace.BadParameter("len not valid for type: %T", a)
	}
}

// TODO(joel): implement elemental functions:
// - regex
// - matches(string, regex, regexes?)
// - contains_regex(array, regex, regexes?)
// - map_insert
// - map_remove
