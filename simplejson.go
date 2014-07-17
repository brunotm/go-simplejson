package simplejson

import (
	"encoding/json"
	"errors"
	"log"
)

// returns the current implementation version
func Version() string {
	return "0.5.0-alpha"
}

type Json struct {
	data interface{}
}

// NewJson returns a pointer to a new `Json` object
// after unmarshaling `body` bytes
func NewJson(body []byte) (*Json, error) {
	j := new(Json)
	err := j.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// New returns a pointer to a new, empty `Json` object
func New() *Json {
	return &Json{
		data: make(map[string]interface{}),
	}
}

// Interface returns the underlying data
func (j *Json) Interface() interface{} {
	return j.data
}

// Encode returns its marshaled data as `[]byte`
func (j *Json) Encode() ([]byte, error) {
	return j.MarshalJSON()
}

// EncodePretty returns its marshaled data as `[]byte` with indentation
func (j *Json) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

// Implements the json.Marshaler interface.
func (j *Json) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

// Set modifies `Json` map by `key` and `value`
// Useful for changing single key/value in a `Json` object easily.
func (j *Json) Set(key string, val interface{}) {
	m, err := j.CheckMap()
	if err != nil {
		return
	}
	m[key] = val
}

// SetPath modifies `Json`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value
func (j *Json) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		j.data = val
		return
	}

	// in order to insert our branch, we need map[string]interface{}
	if _, ok := (j.data).(map[string]interface{}); !ok {
		// have to replace with something suitable
		j.data = make(map[string]interface{})
	}
	curr := j.data.(map[string]interface{})

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(map[string]interface{}); !ok {
			// have to replace with something suitable
			n := make(map[string]interface{})
			curr[b] = n
		}

		curr = curr[b].(map[string]interface{})
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

// Del modifies `Json` map by deleting `key` if it is present.
func (j *Json) Del(key string) {
	m, err := j.CheckMap()
	if err != nil {
		return
	}
	delete(m, key)
}

// getKey returns a pointer to a new `Json` object
// for `key` in its `map` representation
// and a bool identifying success or failure
func (j *Json) getKey(key string) (*Json, bool) {
	m, err := j.CheckMap()
	if err == nil {
		if val, ok := m[key]; ok {
			return &Json{val}, true
		}
	}
	return nil, false
}

// getIndex returns a pointer to a new `Json` object
// for `index` in its `array` representation
// and a bool identifying success or failure
func (j *Json) getIndex(index int) (*Json, bool) {
	a, err := j.CheckArray()
	if err == nil {
		if len(a) > index {
			return &Json{a[index]}, true
		}
	}
	return nil, false
}

// Get searches for the item as specified by the branch
// within a nested Json and returns a new Json pointer
// the pointer is always a valid Json, allowing for chained operations
//
//   newJs := js.Get("top_level", "entries", 3, "dict")
func (j *Json) Get(branch ...interface{}) *Json {
	jin, ok := j.CheckGet(branch...)
	if ok {
		return jin
	}
	return &Json{nil}
}

// CheckGet is like Get, except it also returns a bool
// indicating whenever the branch was found or not
// the Json pointer mai be nil
//
//   newJs, ok := js.Get("top_level", "entries", 3, "dict")
func (j *Json) CheckGet(branch ...interface{}) (*Json, bool) {
	jin := j
	var ok bool
	for _, p := range branch {
		switch p.(type) {
		case string:
			jin, ok = jin.getKey(p.(string))
		case int:
			jin, ok = jin.getIndex(p.(int))
		default:
			ok = false
		}
		if !ok {
			return nil, false
		}
	}
	return jin, true
}

// CheckJsonMap returns a copy of a Json map, but with values as Jsons
func (j *Json) CheckJsonMap() (map[string]*Json, error) {
	m, err := j.CheckMap()
	if err != nil {
		return nil, err
	}
	jm := make(map[string]*Json)
	for key, val := range m {
		jm[key] = &Json{val}
	}
	return jm, nil
}

// CheckJsonArray returns a copy of an array, but with each value as a Json
func (j *Json) CheckJsonArray() ([]*Json, error) {
	a, err := j.CheckArray()
	if err != nil {
		return nil, err
	}
	ja := make([]*Json, len(a))
	for key, val := range a {
		ja[key] = &Json{val}
	}
	return ja, nil
}

// CheckMap type asserts to `map`
func (j *Json) CheckMap() (map[string]interface{}, error) {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// CheckArray type asserts to an `array`
func (j *Json) CheckArray() ([]interface{}, error) {
	if a, ok := (j.data).([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// CheckBool type asserts to `bool`
func (j *Json) CheckBool() (bool, error) {
	if s, ok := (j.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

// CheckString type asserts to `string`
func (j *Json) CheckString() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// CheckBytes type asserts to `[]byte`
func (j *Json) CheckBytes() ([]byte, error) {
	if s, ok := (j.data).(string); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

// JsonArray guarantees the return of a `[]interface{}` (with optional default)
func (j *Json) JsonArray(args ...[]*Json) []*Json {
	var def []*Json

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("JsonArray() received too many arguments %d", len(args))
	}

	a, err := j.CheckJsonArray()
	if err == nil {
		return a
	}

	return def
}

// JsonMap guarantees the return of a `map[string]interface{}` (with optional default)
func (j *Json) JsonMap(args ...map[string]*Json) map[string]*Json {
	var def map[string]*Json

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("JsonMap() received too many arguments %d", len(args))
	}

	a, err := j.CheckJsonMap()
	if err == nil {
		return a
	}

	return def
}

// Array guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, v := range js.Get("results").Array() {
//			fmt.Println(i, v)
//		}
func (j *Json) Array(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Array() received too many arguments %d", len(args))
	}

	a, err := j.CheckArray()
	if err == nil {
		return a
	}

	return def
}

// Map guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//		for k, v := range js.Get("dictionary").Map() {
//			fmt.Println(k, v)
//		}
func (j *Json) Map(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Map() received too many arguments %d", len(args))
	}

	a, err := j.CheckMap()
	if err == nil {
		return a
	}

	return def
}

// String guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.Get("param1").String(), js.Get("optional_param").String("my_default"))
func (j *Json) String(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("String() received too many arguments %d", len(args))
	}

	s, err := j.CheckString()
	if err == nil {
		return s
	}

	return def
}

// Int guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.Get("param1").Int(), js.Get("optional_param").Int(5150))
func (j *Json) Int(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Int() received too many arguments %d", len(args))
	}

	i, err := j.CheckInt()
	if err == nil {
		return i
	}

	return def
}

// Float64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Get("param1").Float64(), js.Get("optional_param").Float64(5.150))
func (j *Json) Float64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Float64() received too many arguments %d", len(args))
	}

	f, err := j.CheckFloat64()
	if err == nil {
		return f
	}

	return def
}

// Bool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.Get("param1").Bool(), js.Get("optional_param").Bool(true))
func (j *Json) Bool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Bool() received too many arguments %d", len(args))
	}

	b, err := j.CheckBool()
	if err == nil {
		return b
	}

	return def
}

// Int64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Get("param1").Int64(), js.Get("optional_param").Int64(5150))
func (j *Json) Int64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Int64() received too many arguments %d", len(args))
	}

	i, err := j.CheckInt64()
	if err == nil {
		return i
	}

	return def
}

// UInt64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Get("param1").Uint64(), js.Get("optional_param").Uint64(5150))
func (j *Json) Uint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Uint64() received too many arguments %d", len(args))
	}

	i, err := j.CheckUint64()
	if err == nil {
		return i
	}

	return def
}
