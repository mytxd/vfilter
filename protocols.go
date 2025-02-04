package vfilter

import (
	"reflect"
	"regexp"
	"strings"
)

type _BoolDispatcher struct {
	implementations []BoolProtocol
}

func (self _BoolDispatcher) Bool(scope *Scope, a Any) bool {
	for _, impl := range self.implementations {
		if impl.Applicable(a) {
			return impl.Bool(scope, a)
		}
	}

	scope.Trace("Protocol Bool not found for %v (%T)", a, a)
	return false
}

func (self *_BoolDispatcher) AddImpl(elements ...BoolProtocol) {
	for _, impl := range elements {
		self.implementations = append(self.implementations, impl)
	}
}

// This protocol implements the truth value.
type BoolProtocol interface {
	Applicable(a Any) bool
	Bool(scope *Scope, a Any) bool
}

// Bool Implementations
type _BoolImpl struct{}

func (self _BoolImpl) Bool(scope *Scope, a Any) bool {
	return a.(bool)
}

func (self _BoolImpl) Applicable(a Any) bool {
	_, ok := a.(bool)
	return ok
}

type _BoolInt struct{}

func (self _BoolInt) Bool(scope *Scope, a Any) bool {
	a_val, _ := to_float(a)
	if a_val != 0 {
		return true
	}

	return false
}

func (self _BoolInt) Applicable(a Any) bool {
	_, a_ok := to_float(a)
	return a_ok
}

type _BoolString struct{}

func (self _BoolString) Bool(scope *Scope, a Any) bool {
	switch t := a.(type) {
	case string:
		return len(t) > 0
	case *string:
		return len(*t) > 0
	}
	return false
}

func (self _BoolString) Applicable(a Any) bool {
	switch a.(type) {
	case string, *string:
		return true
	}
	return false
}

type _BoolSlice struct{}

func (self _BoolSlice) Applicable(a Any) bool {
	return is_array(a)
}

func (self _BoolSlice) Bool(scope *Scope, a Any) bool {
	value_a := reflect.ValueOf(a)
	return value_a.Len() > 0
}

// Eq protocol
type EqProtocol interface {
	Applicable(a Any, b Any) bool
	Eq(scope *Scope, a Any, b Any) bool
}

type _EqDispatcher struct {
	impl []EqProtocol
}

func (self _EqDispatcher) Eq(scope *Scope, a Any, b Any) bool {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Eq(scope, a, b)
		}
	}

	scope.Trace("Protocol Equal not found for %v (%T) and %v (%T)",
		a, a, b, b)
	return false
}

func (self *_EqDispatcher) AddImpl(elements ...EqProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _StringEq struct{}

func (self _StringEq) Eq(scope *Scope, a Any, b Any) bool {
	return a == b
}

func (self _StringEq) Applicable(a Any, b Any) bool {
	_, a_ok := to_string(a)
	_, b_ok := to_string(b)
	return a_ok && b_ok
}

func to_string(x Any) (string, bool) {
	switch t := x.(type) {
	case string:
		return t, true
	case *string:
		return *t, true
	case []byte:
		return string(t), true
	default:
		return "", false
	}
}

func to_float(x Any) (float64, bool) {
	switch t := x.(type) {
	case bool:
		if t {
			return 1, true
		} else {
			return 0, true
		}
	case float64:
		return t, true
	case int:
		return float64(t), true
	case uint:
		return float64(t), true

	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true

	case uint32:
		return float64(t), true
	case int32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case int64:
		return float64(t), true

	default:
		return 0, false
	}
}

// Does x resemble a int?
func is_int(x Any) bool {
	switch x.(type) {
	case bool, int, int8, int16, int32, int64,
		uint8, uint16, uint32, uint64:
		return true
	}

	return false
}

func to_int64(x Any) (int64, bool) {
	switch t := x.(type) {
	case bool:
		if t {
			return 1, true
		} else {
			return 0, true
		}
	case int:
		return int64(t), true
	case uint8:
		return int64(t), true
	case int8:
		return int64(t), true
	case uint16:
		return int64(t), true
	case int16:
		return int64(t), true
	case uint32:
		return int64(t), true
	case int32:
		return int64(t), true
	case uint64:
		return int64(t), true
	case int64:
		return t, true
	case float64:
		return int64(t), true

	default:
		return 0, false
	}
}

// Specialized equivalence for integers - it is not reliable to
// compare floats to ints so we need to special case integers.
type _IntEq struct{}

func (self _IntEq) Applicable(a Any, b Any) bool {
	return is_int(a) && is_int(b)
}

func (self _IntEq) Eq(scope *Scope, a Any, b Any) bool {
	a_val, _ := to_int64(a)
	b_val, _ := to_int64(b)

	return a_val == b_val
}

type _NumericEq struct{}

func (self _NumericEq) Applicable(a Any, b Any) bool {
	_, a_ok := to_float(a)
	_, b_ok := to_float(b)
	return a_ok && b_ok
}

func (self _NumericEq) Eq(scope *Scope, a Any, b Any) bool {
	a_val, _ := to_float(a)
	b_val, _ := to_float(b)

	return a_val == b_val
}

type _ArrayEq struct{}

func (self _ArrayEq) Eq(scope *Scope, a Any, b Any) bool {
	value_a := reflect.ValueOf(a)
	value_b := reflect.ValueOf(b)

	if value_a.Len() != value_b.Len() {
		return false
	}

	for i := 0; i < value_a.Len(); i++ {
		if !scope.eq.Eq(
			scope,
			value_a.Index(i).Interface(),
			value_b.Index(i).Interface()) {
			return false
		}
	}

	return true
}

func is_array(a Any) bool {
	rt := reflect.TypeOf(a)
	if rt == nil {
		return false
	}
	return rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array
}

func (self _ArrayEq) Applicable(a Any, b Any) bool {
	return is_array(a) && is_array(b)
}

// Less than protocol
type LtProtocol interface {
	Applicable(a Any, b Any) bool
	Lt(scope *Scope, a Any, b Any) bool
}

type _LtDispatcher struct {
	impl []LtProtocol
}

func (self _LtDispatcher) Lt(scope *Scope, a Any, b Any) bool {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Lt(scope, a, b)
		}
	}

	return false
}

func (self _LtDispatcher) Applicable(scope *Scope, a Any, b Any) bool {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return true
		}
	}

	scope.Trace("Protocol LessThan not found for %v (%T) and %v (%T)",
		a, a, b, b)
	return false
}

func (self *_LtDispatcher) AddImpl(elements ...LtProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _StringLt struct{}

func (self _StringLt) Lt(scope *Scope, a Any, b Any) bool {
	a_str, _ := to_string(a)
	b_str, _ := to_string(b)

	return a_str < b_str
}

func (self _StringLt) Applicable(a Any, b Any) bool {
	_, a_ok := to_string(a)
	_, b_ok := to_string(b)
	return a_ok && b_ok
}

type _NumericLt struct{}

func (self _NumericLt) Lt(scope *Scope, a Any, b Any) bool {
	a_val, _ := to_float(a)
	b_val, _ := to_float(b)
	return a_val < b_val
}
func (self _NumericLt) Applicable(a Any, b Any) bool {
	_, a_ok := to_float(a)
	_, b_ok := to_float(b)
	return a_ok && b_ok
}

// Add protocol
type AddProtocol interface {
	Applicable(a Any, b Any) bool
	Add(scope *Scope, a Any, b Any) Any
}

type _AddDispatcher struct {
	impl []AddProtocol
}

func (self _AddDispatcher) Add(scope *Scope, a Any, b Any) Any {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Add(scope, a, b)
		}
	}
	scope.Trace("Protocol Add not found for %v (%T) and %v (%T)",
		a, a, b, b)
	return Null{}
}

func (self *_AddDispatcher) AddImpl(elements ...AddProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _AddStrings struct{}

func (self _AddStrings) Applicable(a Any, b Any) bool {
	_, a_ok := to_string(a)
	_, b_ok := to_string(b)
	return a_ok && b_ok
}

func (self _AddStrings) Add(scope *Scope, a Any, b Any) Any {
	a_str, _ := to_string(a)
	b_str, _ := to_string(b)

	return a_str + b_str
}

type _AddInts struct{}

func (self _AddInts) Applicable(a Any, b Any) bool {
	return is_int(a) && is_int(b)
}

func (self _AddInts) Add(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_int64(a)
	b_val, _ := to_int64(b)
	return a_val + b_val
}

type _AddFloats struct{}

func (self _AddFloats) Applicable(a Any, b Any) bool {
	_, a_ok := to_float(a)
	_, b_ok := to_float(b)
	return a_ok && b_ok
}

func (self _AddFloats) Add(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_float(a)
	b_val, _ := to_float(b)
	return a_val + b_val
}

// Add two slices together.
type _AddSlices struct{}

func (self _AddSlices) Applicable(a Any, b Any) bool {
	return is_array(a) && is_array(b)
}

func (self _AddSlices) Add(scope *Scope, a Any, b Any) Any {
	var result []Any
	a_slice := reflect.ValueOf(a)
	b_slice := reflect.ValueOf(b)

	for i := 0; i < a_slice.Len(); i++ {
		result = append(result, a_slice.Index(i).Interface())
	}

	for i := 0; i < b_slice.Len(); i++ {
		result = append(result, b_slice.Index(i).Interface())
	}

	return result
}

func is_null(a Any) bool {
	if a == nil {
		return true
	}

	switch a.(type) {
	case Null, *Null:
		return true
	}

	return false
}

// Add a slice to null. We treat null as the empty array.
type _AddNull struct{}

func (self _AddNull) Applicable(a Any, b Any) bool {
	return (is_array(a) && is_null(b)) || (is_null(a) && is_array(b))
}

func (self _AddNull) Add(scope *Scope, a Any, b Any) Any {
	if is_null(a) {
		return b
	}
	return a
}

// Add a slice to Any will expand the slice and add each item with the
// any.
type _AddSliceAny struct{}

func (self _AddSliceAny) Applicable(a Any, b Any) bool {
	return is_array(a) || is_array(b)
}

func (self _AddSliceAny) Add(scope *Scope, a Any, b Any) Any {
	var result []Any

	if is_array(a) {
		a_slice := reflect.ValueOf(a)

		for i := 0; i < a_slice.Len(); i++ {
			result = append(result, a_slice.Index(i).Interface())
		}

		return append(result, b)
	}

	result = append(result, a)
	b_slice := reflect.ValueOf(b)

	for i := 0; i < b_slice.Len(); i++ {
		result = append(result, b_slice.Index(i).Interface())
	}

	return result
}

// Sub protocol
type SubProtocol interface {
	Applicable(a Any, b Any) bool
	Sub(scope *Scope, a Any, b Any) Any
}

type _SubDispatcher struct {
	impl []SubProtocol
}

func (self _SubDispatcher) Sub(scope *Scope, a Any, b Any) Any {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Sub(scope, a, b)
		}
	}

	scope.Trace("Protocol Sub not found for %v (%T) and %v (%T)",
		a, a, b, b)
	return Null{}
}

func (self *_SubDispatcher) AddImpl(elements ...SubProtocol) {
	for _, impl := range elements {

		self.impl = append(self.impl, impl)
	}
}

type _SubFloats struct{}

func (self _SubFloats) Applicable(a Any, b Any) bool {
	_, a_ok := to_float(a)
	_, b_ok := to_float(b)
	return a_ok && b_ok
}

func (self _SubFloats) Sub(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_float(a)
	b_val, _ := to_float(b)
	return a_val - b_val
}

type _SubInts struct{}

func (self _SubInts) Applicable(a Any, b Any) bool {
	return is_int(a) && is_int(b)
}

func (self _SubInts) Sub(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_int64(a)
	b_val, _ := to_int64(b)
	return a_val - b_val
}

// Multiply protocol
type MulProtocol interface {
	Applicable(a Any, b Any) bool
	Mul(scope *Scope, a Any, b Any) Any
}

type _MulDispatcher struct {
	impl []MulProtocol
}

func (self _MulDispatcher) Mul(scope *Scope, a Any, b Any) Any {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Mul(scope, a, b)
		}
	}
	scope.Trace("Protocol Mul not found for %v (%T) and %v (%T)",
		a, a, b, b)

	return Null{}
}

func (self *_MulDispatcher) AddImpl(elements ...MulProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _MulInt struct{}

func (self _MulInt) Applicable(a Any, b Any) bool {
	return is_int(a) && is_int(b)
}

func (self _MulInt) Mul(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_int64(a)
	b_val, _ := to_int64(b)
	return a_val * b_val
}

type _NumericMul struct{}

func (self _NumericMul) Applicable(a Any, b Any) bool {
	_, a_ok := to_float(a)
	_, b_ok := to_float(b)
	return a_ok && b_ok
}

func (self _NumericMul) Mul(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_float(a)
	b_val, _ := to_float(b)
	return a_val * b_val
}

// Divtiply protocol
type DivProtocol interface {
	Applicable(a Any, b Any) bool
	Div(scope *Scope, a Any, b Any) Any
}

type _DivDispatcher struct {
	impl []DivProtocol
}

func (self _DivDispatcher) Div(scope *Scope, a Any, b Any) Any {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Div(scope, a, b)
		}
	}

	scope.Trace("Protocol Div not found for %v (%T) and %v (%T)",
		a, a, b, b)

	return Null{}
}

func (self *_DivDispatcher) AddImpl(elements ...DivProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _NumericDiv struct{}

func (self _NumericDiv) Applicable(a Any, b Any) bool {
	_, a_ok := to_float(a)
	_, b_ok := to_float(b)
	return a_ok && b_ok
}

func (self _NumericDiv) Div(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_float(a)
	b_val, _ := to_float(b)
	if b_val == 0 {
		return false
	}

	return a_val / b_val
}

type _DivInt struct{}

func (self _DivInt) Applicable(a Any, b Any) bool {
	return is_int(a) && is_int(b)
}

func (self _DivInt) Div(scope *Scope, a Any, b Any) Any {
	a_val, _ := to_int64(a)
	b_val, _ := to_int64(b)
	if b_val == 0 {
		return false
	}

	return a_val / b_val
}

// Membership protocol
type MembershipProtocol interface {
	Applicable(a Any, b Any) bool
	Membership(scope *Scope, a Any, b Any) bool
}

type _MembershipDispatcher struct {
	impl []MembershipProtocol
}

func (self _MembershipDispatcher) Membership(scope *Scope, a Any, b Any) bool {

	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			return impl.Membership(scope, a, b)
		}
	}

	// Default behavior: Test lhs against each member in RHS -
	// slow but works.
	rt := reflect.TypeOf(b)
	if rt == nil {
		return false
	}

	kind := rt.Kind()
	value := reflect.ValueOf(b)
	if kind == reflect.Slice || kind == reflect.Array {
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i).Interface()
			if scope.eq.Eq(scope, a, item) {
				return true
			}
		}
	} else {
		scope.Trace("Protocol Membership not found for %v (%T) and %v (%T)",
			a, a, b, b)
	}

	return false
}

func (self *_MembershipDispatcher) AddImpl(elements ...MembershipProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _SubstringMembership struct{}

func (self _SubstringMembership) Applicable(a Any, b Any) bool {
	_, a_ok := to_string(a)
	_, b_ok := to_string(b)
	return a_ok && b_ok
}

func (self _SubstringMembership) Membership(scope *Scope, a Any, b Any) bool {
	a_str, _ := to_string(a)
	b_str, _ := to_string(b)

	return strings.Contains(b_str, a_str)
}

// Associative protocol.
type AssociativeProtocol interface {
	Applicable(a Any, b Any) bool

	// Returns a value obtained by dereferencing field b from
	// object a. If not present return pres == false and possibly
	// a default value in res. If no default is present res must
	// be nil.
	Associative(scope *Scope, a Any, b Any) (res Any, pres bool)
	GetMembers(scope *Scope, a Any) []string
}

type _AssociativeDispatcher struct {
	impl []AssociativeProtocol
}

func (self *_AssociativeDispatcher) Associative(
	scope *Scope, a Any, b Any) (Any, bool) {
	for _, impl := range self.impl {
		if impl.Applicable(a, b) {
			res, pres := impl.Associative(scope, a, b)
			return res, pres
		}
	}
	res, pres := DefaultAssociative{}.Associative(scope, a, b)
	return res, pres
}

func (self *_AssociativeDispatcher) GetMembers(
	scope *Scope, a Any) []string {
	for _, impl := range self.impl {
		if impl.Applicable(a, "") {
			return impl.GetMembers(scope, a)
		}
	}
	return DefaultAssociative{}.GetMembers(scope, a)
}

func (self *_AssociativeDispatcher) AddImpl(elements ...AssociativeProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

// Last resort associative - uses reflect package to resolve struct
// fields.
type DefaultAssociative struct{}

func (self DefaultAssociative) Applicable(a Any, b Any) bool {
	return false
}

func (self DefaultAssociative) Associative(scope *Scope, a Any, b Any) (res Any, pres bool) {
	defer func() {
		// If an error occurs we return false - not found.
		recover()
	}()
	switch field_name := b.(type) {
	case *float64:
		a_value := reflect.Indirect(reflect.ValueOf(a))
		idx := int(*field_name)
		if idx < 0 || idx > a_value.Len() {
			return &Null{}, false
		}
		value := a_value.Index(idx)
		if value.Kind() == reflect.Ptr && value.IsNil() {
			return &Null{}, true
		}
		return value.Interface(), true

	case *int64:
		a_value := reflect.Indirect(reflect.ValueOf(a))
		idx := int(*field_name)
		if idx < 0 || idx > a_value.Len() {
			return &Null{}, false
		}
		value := a_value.Index(idx)
		if value.Kind() == reflect.Ptr && value.IsNil() {
			return &Null{}, true
		}
		return value.Interface(), true

	case string:
		if !is_exported(field_name) {
			field_name = strings.Title(field_name)
		}

		a_value := reflect.Indirect(reflect.ValueOf(a))
		a_type := a_value.Type()

		// A struct with regular exportable field.
		if a_type.Kind() == reflect.Struct {
			field_value := a_value.FieldByName(field_name)
			if field_value.IsValid() && field_value.CanInterface() {
				if field_value.Kind() == reflect.Ptr && field_value.IsNil() {
					return &Null{}, true
				}
				return field_value.Interface(), true
			}
		}

		// A method we call. Usually this is a Getter.
		method_value := reflect.ValueOf(a).MethodByName(field_name)
		if _Callable(method_value, field_name) {
			if method_value.Type().Kind() == reflect.Ptr {
				method_value = method_value.Elem()
			}

			results := method_value.Call([]reflect.Value{})

			// In Go, a common pattern is to
			// return (value, err). We try to
			// guess here by taking the first
			// return value as the value.
			if len(results) == 1 || len(results) == 2 {
				res := results[0]
				if res.CanInterface() {
					if res.Kind() == reflect.Ptr && res.IsNil() {
						return &Null{}, true
					}

					return res.Interface(), true
				}
			}
		}

		// An array - we call Associative on each member.
		if a_type.Kind() == reflect.Slice {
			var result []Any

			for i := 0; i < a_value.Len(); i++ {
				element := a_value.Index(i).Interface()
				if item, pres := scope.Associative(element, b); pres {
					result = append(result, item)
				}
			}

			return result, true
		}
	}

	return Null{}, false
}

// Get the members which are callable by VFilter.
func (self DefaultAssociative) GetMembers(scope *Scope, a Any) []string {
	var result []string

	a_value := reflect.Indirect(reflect.ValueOf(a))
	if a_value.Kind() == reflect.Struct {
		for i := 0; i < a_value.NumField(); i++ {
			field_type := a_value.Type().Field(i)
			if is_exported(field_type.Name) {
				result = append(result, field_type.Name)
			}
		}
	}

	a_value = reflect.ValueOf(a)

	// If a value is a slice, we get the members of the
	// first item. Hopefully they are the same as the
	// other items. A common use case is storing the
	// output of a query in the scope environment and then
	// selecting from it, in which case the value is a
	// list of Rows, each row has a Dict.
	if a_value.Type().Kind() == reflect.Slice {
		for i := 0; i < a_value.Len(); i++ {
			return scope.GetMembers(a_value.Index(i).Interface())
		}
	}

	for i := 0; i < a_value.NumMethod(); i++ {
		method_type := a_value.Type().Method(i)
		method_value := a_value.Method(i)
		if _Callable(method_value, method_type.Name) {
			result = append(result, method_type.Name)
		}
	}

	return result
}

// Regex Match protocol
type RegexProtocol interface {
	Applicable(pattern Any, target Any) bool
	Match(scope *Scope, pattern Any, target Any) bool
}

type _RegexDispatcher struct {
	impl []RegexProtocol
}

func (self _RegexDispatcher) Match(scope *Scope, pattern Any, target Any) bool {
	for _, impl := range self.impl {
		if impl.Applicable(pattern, target) {
			return impl.Match(scope, pattern, target)
		}
	}

	scope.Trace("Protocol Regex not found for %v (%T) and %v (%T)",
		pattern, pattern, target, target)

	return false
}

func (self *_RegexDispatcher) AddImpl(elements ...RegexProtocol) {
	for _, impl := range elements {
		self.impl = append(self.impl, impl)
	}
}

type _SubstringRegex struct{}

func (self _SubstringRegex) Applicable(pattern Any, target Any) bool {
	_, a_ok := to_string(pattern)
	_, b_ok := to_string(target)

	return a_ok && b_ok
}

func (self _SubstringRegex) Match(scope *Scope, pattern Any, target Any) bool {
	pattern_string, _ := to_string(pattern)
	target_string, _ := to_string(target)

	re, pres := scope.regexp_cache[pattern_string]
	if !pres {
		var err error
		re, err = regexp.Compile("(?i)" + pattern_string)
		if err != nil {
			scope.Log("Compile regexp: %v", err)
			return false
		}

		scope.regexp_cache[pattern_string] = re
	}

	return re.MatchString(target_string)
}

type _ArrayRegex struct{}

func (self _ArrayRegex) Applicable(pattern Any, target Any) bool {
	_, pattern_ok := to_string(pattern)
	return pattern_ok && is_array(target)
}

func (self _ArrayRegex) Match(scope *Scope, pattern Any, target Any) bool {
	a_slice := reflect.ValueOf(target)
	for i := 0; i < a_slice.Len(); i++ {
		if scope.Match(pattern, a_slice.Index(i).Interface()) {
			return true
		}
	}

	return false
}

type StringProtocol interface {
	ToString(scope *Scope) string
}
