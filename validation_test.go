package validate

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func ExampleStruct() {
	u := &UserForm{
		Name: "inhere",
	}

	v := Struct(u)
	ok := v.Validate()

	fmt.Println(ok)
}

var mpSample = M{
	"age":   100,
	"name":  "inhere",
	"oldSt": 1,
	"newSt": 2,
	"email": "some@e.com",
}

func TestMap(t *testing.T) {
	is := assert.New(t)

	m := M{
		"name":  "inhere",
		"age":   100,
		"oldSt": 1,
		"newSt": 2,
		"email": "some@e.com",
	}

	v := New(m)
	v.AddRule("name", "required")
	v.AddRule("name", "minLen", 7)
	v.AddRule("age", "max", 99)
	v.AddRule("age", "min", 1)

	ok := v.Validate()
	is.False(ok)
	is.Equal("name min length is 7", v.Errors.Get("name"))
	is.Empty(v.SafeData())

	v = New(nil)
	is.Contains(v.Errors.String(), "invalid input data")
	is.False(v.Validate())
}

func TestValidation_StringRule(t *testing.T) {
	is := assert.New(t)

	v := Map(mpSample)
	v.StringRules(MS{
		"name":  "string|len:6|minLen:2|maxLen:10",
		"oldSt": "lt:5|gt:0|in:1,2,3|notIn:4,5",
		"age":   "max:100",
	})
	v.StringRule("newSt", "required|int:1|gtField:oldSt")
	ok := v.Validate()
	is.True(ok)

	v = Map(mpSample)
	v.StringRule("newSt", "required|int:1,5")
	is.True(v.Validate())
}

// UserForm struct
type UserForm struct {
	Name     string    `validate:"required|minLen:7"`
	Email    string    `validate:"email"`
	CreateAt int       `validate:"email"`
	Safe     int       `validate:"-"`
	UpdateAt time.Time `validate:"required"`
	Code     string    `validate:"customValidator"`
	Status   int       `validate:"required|gtField:Extra.Status1"`
	Extra    ExtraInfo `validate:"required"`
}

// ExtraInfo data
type ExtraInfo struct {
	Github  string `validate:"required|url"`
	Status1 int    `validate:"required|int"`
}

// custom validator in the source struct.
func (f UserForm) CustomValidator(val string) bool {
	return len(val) == 4
}

func (f UserForm) ConfigValidation(v *Validation) {
	v.AddTranslates(MS{
		"Safe": "Safe-Name",
	})
}

// Messages you can custom define validator error messages.
func (f UserForm) Messages() map[string]string {
	return MS{
		"required":      "oh! the {field} is required",
		"Name.required": "message for special field",
	}
}

// Translates you can custom field translates.
func (f UserForm) Translates() map[string]string {
	return MS{
		"Name":  "User Name",
		"Email": "User Email",
	}
}

func TestStruct(t *testing.T) {
	is := assert.New(t)

	u := &UserForm{
		Name: "inhere",
	}
	v := Struct(u)

	// check trans data
	is.True(v.Trans().HasField("Name"))
	is.True(v.Trans().HasField("Safe"))
	is.True(v.Trans().HasMessage("Name.required"))

	// validate
	ok := v.Validate()
	is.False(ok)
	is.Equal("User Name min length is 7", v.Errors.Field("Name")[0])
	is.Empty(v.SafeData())
}

func TestJSON(t *testing.T) {
	is := assert.New(t)

	v := JSON(`{
	"name": "inhere",
	"age": 100
}`)

	v.StopOnError = false
	v.StringRules(MS{
		"name": "required|minLen:7",
		"age":  "required|int|range:1,99",
	})

	is.False(v.Validate())
	is.Empty(v.SafeData())

	is.Contains(v.Errors, "age")
	is.Contains(v.Errors, "name")
	is.Contains(v.Errors.String(), "name min length is 7")
	is.Contains(v.Errors.String(), "age value must be in the range 1 - 99")
}

func TestFromQuery(t *testing.T) {
	is := assert.New(t)

	data := url.Values{
		"name": []string{"inhere"},
		"age":  []string{"10"},
	}

	v := FromQuery(data).Create()
	v.StopOnError = false
	v.FilterRule("age", "int")
	v.StringRules(MS{
		"name": "required|minLen:7",
		"age":  "required|int|min:10",
	})

	is.False(v.Validate())
	is.Equal("name min length is 7", v.Errors.Field("name")[0])
	is.Empty(v.SafeData())
}

func TestValidationScene(t *testing.T) {
	is := assert.New(t)
	mp := M{
		"name": "inhere",
		"age":  100,
	}

	v := Map(mp)
	v.StopOnError = false
	v.StringRules(MS{
		"name": "minLen:7",
		"age":  "min:101",
	})
	v.WithScenes(SValues{
		"create": []string{"name", "age"},
		"update": []string{"name"},
	})

	// on scene "create"
	ok := v.Validate("create")
	is.False(ok)
	is.False(v.Errors.Empty())
	is.Contains(v.Errors.Error(), "age")
	is.Contains(v.Errors.Error(), "name")

	// on scene "update"
	v.ResetResult()
	ok = v.Validate("update")
	is.False(ok)
	is.Contains(v.Errors, "name")
	is.NotContains(v.Errors, "age")
	is.Equal("", v.Errors.Get("age"))
	is.Equal("name min length is 7", v.Errors.One())
}
