package namedlist

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/veggiemonk/strcase"
)

type MapperFunc func(string) string

type NamedList struct {
	columnSeparator string
	mapperFunc      MapperFunc
	tagKey          string
}

type Option func(list *NamedList)

func New(opts ...Option) (*NamedList, error) {
	namedList := &NamedList{
		columnSeparator: "_",
		mapperFunc:      SnakeCaseMapper,
		tagKey:          "db",
	}

	for _, opt := range opts {
		opt(namedList)
	}

	return namedList, nil
}

func WithColumnSeparator(sep string) Option {
	return func(list *NamedList) {
		list.columnSeparator = sep
	}
}

func WithMapperFunc(mapper MapperFunc) Option {
	return func(list *NamedList) {
		list.mapperFunc = mapper
	}
}

func WithTagKey(key string) Option {
	return func(list *NamedList) {
		list.tagKey = key
	}
}

func (nl *NamedList) FromStruct(src any) ([]any, error) {
	var results = make([]any, 0)

	srcType := reflect.TypeOf(src)
	if isStruct(srcType) == false {
		return nil, errors.New("src is not struct or pointer to a struct")
	}

	srcIsPointer := false
	if srcType.Kind() == reflect.Ptr {
		// We need the underlying Struct in order to get the visible fields.
		srcType = srcType.Elem()
		srcIsPointer = true
	}

	fields := reflect.VisibleFields(srcType)
	for _, field := range fields {
		if field.Anonymous == true {
			continue
		}

		rv := reflect.ValueOf(src)
		var fieldValue any
		if srcIsPointer == true {
			// When the input is a pointer to a struct we need to dereference
			// the pointer (rv.Elem) before we attempt to get the value of the
			// field by index.
			fieldValue = rv.Elem().FieldByIndex(field.Index).Interface()
		} else {
			// When the input is a value by copy, we can simply retrieve the
			// value at the field index.
			fieldValue = rv.FieldByIndex(field.Index).Interface()
		}

		tag, tagPresent := field.Tag.Lookup(nl.tagKey)
		fieldValueAsIs := false
		convertedFieldName := nl.mapperFunc(field.Name)
		if tagPresent == true && tag != "" {
			if tag == "-" {
				// Skip ignored field.
				continue
			}
			tagParts := strings.Split(tag, ",")
			if tagParts[0] != "." {
				// The `.` tag lets us keep the name from the mapper function while
				// being able to supply further tag options like `asis`.
				convertedFieldName = tagParts[0]
			}
			// As-is fields are intended to support things like a time.Time that have
			// non-public fields that we cannot iterate. In other words, we copy the
			// value of as-is fields directly into the value parameter of [sql.Named]
			// and let the [sql.driver.Valuer] handle the serialization.
			fieldValueAsIs = slices.Contains(tagParts[1:], "asis")
		}

		if isStruct(field.Type) == true && fieldValueAsIs == false {
			list, err := nl.FromStruct(fieldValue)
			if err != nil {
				return nil, err
			}
			for _, e := range list {
				arg := e.(sql.NamedArg)
				arg.Name = fmt.Sprintf("%s%s%s", convertedFieldName, nl.columnSeparator, arg.Name)
				results = append(results, arg)
			}
		} else {
			wrapped := sql.Named(convertedFieldName, fieldValue)
			results = append(results, wrapped)
		}

	}

	return results, nil
}

func SnakeCaseMapper(input string) string {
	return strcase.Snake(input)
}

func isStruct(value reflect.Type) bool {
	kind := value.Kind()
	if kind == reflect.Struct {
		return true
	}
	if kind == reflect.Ptr {
		return isStruct(value.Elem())
	}
	return false
}
