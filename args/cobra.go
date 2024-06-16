package args

import (
	"reflect"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/oleiade/reflections"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

func RegisterArgs[T any](v *T, flagSet *pflag.FlagSet) error {
	fields, err := reflections.Fields(v)
	if err != nil {
		return err
	}

	for _, fieldName := range fields {
		// grab the field
		field, err := reflections.GetField(v, fieldName)
		if err != nil {
			return err
		}

		// grab the tag
		tag, err := reflections.GetFieldTag(v, fieldName, "arg")
		if err != nil {
			return err
		}

		parsedTag, err := parseTag(tag)
		if err != nil {
			return err
		}

		// parse the tag
		var name string
		if parsedTag.Name != nil {
			name = *parsedTag.Name
		} else {
			name = stringy.New(fieldName).KebabCase().ToLower()
		}

		switch {
		case reflect.TypeOf(field) == reflect.TypeOf(true):
			flagSet.BoolVarP(reflect.ValueOf(v).Elem().FieldByName(fieldName).Addr().Interface().(*bool), name, lo.FromPtr(parsedTag.Short), reflect.ValueOf(field).Bool(), lo.FromPtr(parsedTag.Description))
		case reflect.TypeOf(field) == reflect.TypeOf(""):
			flagSet.StringVarP(reflect.ValueOf(v).Elem().FieldByName(fieldName).Addr().Interface().(*string), name, lo.FromPtr(parsedTag.Short), reflect.ValueOf(field).String(), lo.FromPtr(parsedTag.Description))
		case reflect.TypeOf(field) == reflect.TypeOf(0):
			flagSet.IntVarP(reflect.ValueOf(v).Elem().FieldByName(fieldName).Addr().Interface().(*int), name, lo.FromPtr(parsedTag.Short), int(reflect.ValueOf(field).Int()), lo.FromPtr(parsedTag.Description))
		case reflect.TypeOf(field) == reflect.TypeOf([]string{}):
			flagSet.StringSliceVarP(reflect.ValueOf(v).Elem().FieldByName(fieldName).Addr().Interface().(*[]string), name, lo.FromPtr(parsedTag.Short), reflect.ValueOf(field).Interface().([]string), lo.FromPtr(parsedTag.Description))
		case reflect.TypeOf(field) == reflect.TypeOf([]int{}):
			flagSet.IntSliceVarP(reflect.ValueOf(v).Elem().FieldByName(fieldName).Addr().Interface().(*[]int), name, lo.FromPtr(parsedTag.Short), reflect.ValueOf(field).Interface().([]int), lo.FromPtr(parsedTag.Description))
		default:
			return errors.Errorf("type %s not supported by this yet", reflect.TypeOf(field))
		}

	}

	return nil
}

type Tag struct {
	Name        *string
	Short       *string
	Description *string
}

func parseTag(tag string) (Tag, error) {
	t := Tag{}

	split := strings.Split(tag, ",")
	if len(split) == 1 {
		if len(tag) != 0 {
			t.Name = &tag
			return t, nil
		}
	}

	// otherwise we've more than 1 thing
	if len(split[0]) > 0 {
		t.Name = &split[0]
	}

	for _, s := range split[1:] {
		parts := strings.Split(s, "=")
		if len(parts) != 2 {
			return t, errors.Errorf("invalid tag: %s", tag)
		}

		switch parts[0] {
		case "short":
			t.Short = &parts[1]
		case "desc", "description":
			t.Description = &parts[1]
		default:
			return t, errors.Errorf("invalid tag: %s", tag)
		}
	}

	return t, nil
}
