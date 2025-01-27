package validate

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// New a Validation
func New(data interface{}, scene ...string) *Validation {
	switch td := data.(type) {
	case DataFace:
		return NewValidation(td, scene...)
	case M:
		return FromMap(td).Create().SetScene(scene...)
	case map[string]interface{}:
		return FromMap(td).Create().SetScene(scene...)
	case SValues:
		return FromURLValues(url.Values(td)).Create().SetScene(scene...)
	case url.Values:
		return FromURLValues(td).Create().SetScene(scene...)
	case map[string][]string:
		return FromURLValues(td).Create().SetScene(scene...)
	}

	return Struct(data, scene...)
}

// TODO since v1.2 ...
// func NewWithOptions(data interface{}, func(Options))  {
// }

// Map validation create
func Map(m map[string]interface{}, scene ...string) *Validation {
	return FromMap(m).Create().SetScene(scene...)
}

// JSON create validation from JSON string.
func JSON(s string, scene ...string) *Validation {
	return newWithError(FromJSON(s)).SetScene(scene...)
}

// Struct validation create
func Struct(s interface{}, scene ...string) *Validation {
	return newWithError(FromStruct(s)).SetScene(scene...)
}

// Request validation create
func Request(r *http.Request) *Validation {
	return newWithError(FromRequest(r))
}

// Config global options
func Config(fn func(opt *GlobalOption)) {
	fn(globalOpt)
}

/*************************************************************
 * create data instance
 *************************************************************/

// FromMap build data instance.
func FromMap(m map[string]interface{}) *MapData {
	data := &MapData{}
	if m != nil {
		data.Map = m
		data.value = reflect.ValueOf(m)
	}
	return data
}

// FromJSON string build data instance.
func FromJSON(s string) (*MapData, error) {
	return FromJSONBytes([]byte(s))
}

// FromJSONBytes string build data instance.
func FromJSONBytes(bs []byte) (*MapData, error) {
	mp := map[string]interface{}{}
	if err := json.Unmarshal(bs, &mp); err != nil {
		return nil, err
	}

	data := &MapData{
		Map:   mp,
		value: reflect.ValueOf(mp),
		// save JSON bytes
		bodyJSON: bs,
	}

	return data, nil
}

// FromStruct create a Data from struct
func FromStruct(s interface{}) (*StructData, error) {
	data := &StructData{
		ValidateTag: globalOpt.ValidateTag,
		// init map
		fieldNames:  make(map[string]int),
		fieldValues: make(map[string]interface{}),
	}

	if s == nil {
		return data, ErrInvalidData
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	typ := val.Type()
	if val.Kind() != reflect.Struct || typ == timeType {
		return data, ErrInvalidData
	}

	data.src = s
	data.value = val
	data.valueTpy = typ

	return data, nil
}

// FromRequest collect data from request instance
func FromRequest(r *http.Request, maxMemoryLimit ...int64) (DataFace, error) {
	// no body. like GET DELETE ....
	if r.Method != "POST" && r.Method != "PUT" && r.Method != "PATCH" {
		return FromURLValues(r.URL.Query()), nil
	}

	cType := r.Header.Get("Content-Type")

	// contains file uploaded form
	// strings.HasPrefix(mediaType, "multipart/")
	if strings.Contains(cType, "multipart/form-data") {
		maxMemory := defaultMaxMemory
		if len(maxMemoryLimit) > 0 {
			maxMemory = maxMemoryLimit[0]
		}

		if err := r.ParseMultipartForm(maxMemory); err != nil {
			return nil, err
		}

		// collect from values
		data := FromURLValues(r.MultipartForm.Value)
		// collect uploaded files
		data.AddFiles(r.MultipartForm.File)
		// add queries data
		data.AddValues(r.URL.Query())
		return data, nil
	}

	// basic POST form. content type: application/x-www-form-urlencoded
	if strings.Contains(cType, "form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}

		data := FromURLValues(r.PostForm)
		// add queries data
		data.AddValues(r.URL.Query())
		return data, nil
	}

	// JSON body request
	if strings.Contains(cType, "application/json") {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		return FromJSONBytes(bs)
	}

	return nil, ErrEmptyData
}

// FromURLValues build data instance.
func FromURLValues(values url.Values) *FormData {
	data := newFormData()
	for key, vals := range values {
		for _, val := range vals {
			data.Add(key, val)
		}
	}

	return data
}

// FromQuery build data instance.
// Usage:
// 	validate.FromQuery(r.URL.Query()).Create()
func FromQuery(values url.Values) *FormData {
	return FromURLValues(values)
}
