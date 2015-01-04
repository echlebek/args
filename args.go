package args

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

/*
Parse unmarshals command-line arguments into data. It uses reflection to
unmarshal arguments into various data types.

Parse can unmarshal into three different types:
	* struct
	* map
	* slice

The main use case is struct. Example:

	import (
		"fmt"
		"github.com/echlebek/args"
		"os"
	)

	type Args struct {
		Foo int     `args:"this is a foo,-f"` // Foo has a short flag, -f
		Bar float32 `args:"this is a bar,r"`  // Bar is required
		Baz string  `args:"a baz!,r"`         // Baz is required too
	}

	var defaultArgs = Args{Foo: 5}

	const progDesc = "A mock program"

	func main() {
		a := defaultArgs

		if err := args.Parse(&a); err != nil {
			fmt.Errorf("%s\n", err)
			if err := args.Usage(os.Stdout, defaultArgs, progDesc); err != nil {
				fmt.Errorf("%s\n", err)
			}
			return
		}

		fmt.Printf("%+v\n", a)
	}

$ ./main --foo 5 --bar 3.5 --baz asdf

{Foo:5 Bar:3.5 Baz:asdf}
*/
func Parse(strukt interface{}) error {
	return parse(strukt, os.Args[1:])
}

// Usage writes the usage for a program to w, given a struct with members
// and tags outlining its purpose. Description should be a short one-line
// description of what the program does.
func Usage(w io.Writer, strukt interface{}) error {
	val := reflect.ValueOf(strukt)
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf(
			"can only print usage with struct, not %s", typ.Kind().String())
	}
	if _, err := fmt.Fprint(w, "usage:\n"); err != nil {
		return err
	}
	for i := 0; i < typ.NumField(); i++ {
		ftype := typ.Field(i)
		fval := val.Field(i)
		if ftype.PkgPath == "" {
			if err := usageForField(w, ftype, fval); err != nil {
				return err
			}
		}
	}

	return nil
}

// Embed Positionals in your args struct to add positional argument support.
type Positionals struct {
	data []string
}

// usageForField writes the usage for a single argument from a struct field
func usageForField(w io.Writer, field reflect.StructField, fieldVal reflect.Value) error {
	td := parseTagData(field.Tag)
	if td.ShortFlag != "" {
		if _, err := fmt.Fprintf(w, " -%s,\t", td.ShortFlag); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprint(w, " \t"); err != nil {
			return err
		}
	}
	if fieldVal.Kind() != reflect.Ptr {
		var defaultValue string
		if fieldVal.Kind() == reflect.String {
			defaultValue = fmt.Sprintf("%q", fieldVal.Interface())
		} else {
			defaultValue = fmt.Sprintf("%v", fieldVal.Interface())
		}
		_, err := fmt.Fprintf(
			w, "--%s\t(default: %s)\t%s\n", strings.ToLower(field.Name),
			defaultValue, td.Description)
		if err != nil {
			return err
		}
	} else {
		_, err := fmt.Fprintf(
			w, "--%s\t\t%s\n", strings.ToLower(field.Name), td.Description)

		if err != nil {
			return err
		}
	}
	return nil
}

// implementation of Parse
func parse(data interface{}, args []string) error {
	typ, err := getType(data)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(data).Elem()
	switch typ.Kind() {
	case reflect.Struct:
		return parseStruct(v, args)
	case reflect.Slice:
		return fillSlice(v, args)
	case reflect.Map:
		return parseMap(v, args)
	default: // should never be reached
		return fmt.Errorf("invalid type for unmarshal: %s", typ.Kind().String())
	}
}

// checkArgLen returns an error if the length of the data != 1
func checkArgLen(data []string, name string) error {
	if len(data) == 0 {
		return fmt.Errorf("args: option %s specified but not set", name)
	} else if len(data) > 1 {
		return fmt.Errorf("args: option %s specified more than once", name)
	}
	return nil
}

type tagData struct {
	Description string
	Required    bool
	ShortFlag   string
}

func parseTagData(tag reflect.StructTag) tagData {
	td := tagData{}
	parts := strings.Split(tag.Get("args"), ",")
	if len(parts) > 0 {
		td.Description = parts[0]
	}
	if len(parts) > 1 {
		for i := 1; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "-") && len(parts[i]) == 2 {
				td.ShortFlag = parts[i][1:]
			} else if parts[i] == "r" {
				td.Required = true
			}
		}
	}
	return td
}

// parseStruct walks the struct fields of v, and tries to assign items
// from args to them.
func parseStruct(v reflect.Value, args []string) error {
	rawData, err := rawArgsMap(args)
	if err != nil {
		return err
	}
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" || field.Anonymous {
			// Non-empty PkgPath implies unexported field
			continue
		}
		name := strings.ToLower(field.Name)
		ftype := field.Type
		fval := v.Field(i)
		tagData := parseTagData(field.Tag)
		// Try to find a mapping from the struct field to the command line flag
		data, ok := rawData[name]
		if !ok {
			// Didn't find any flag that matches this name, try the short flag
			if data, ok = rawData[tagData.ShortFlag]; !ok {
				// Nothing was found. If it's not required, that's OK. Otherwise, error.
				if !tagData.Required {
					continue
				} else {
					return fmt.Errorf("%s: required argument was not supplied: --%s", os.Args[0], name)
				}
			}
		}
		if ftype.Kind() == reflect.Ptr {
			ftype = ftype.Elem()
			fval.Set(reflect.New(fval.Type().Elem()))
			fval = fval.Elem()
		}
		switch ftype.Kind() {
		case reflect.String:
			if err := checkArgLen(data, name); err != nil {
				return fmt.Errorf("args: %s", err)
			}
			fval.Set(reflect.ValueOf(data[0]))

		case reflect.Bool:
			if len(data) != 0 {
				return fmt.Errorf("args: boolean option %s has parameter", name)
			}
			fval.Set(reflect.ValueOf(true))

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if err := checkArgLen(data, name); err != nil {
				return fmt.Errorf("args: %s", err)
			}
			n, err := strconv.ParseInt(data[0], 0, ftype.Bits())
			if err != nil {
				return fmt.Errorf("args: %s", err)
			}
			fval.SetInt(n)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if err := checkArgLen(data, name); err != nil {
				return fmt.Errorf("args: %s", err)
			}
			n, err := strconv.ParseUint(data[0], 0, ftype.Bits())
			if err != nil {
				return fmt.Errorf("args: %s", err)
			}
			fval.SetUint(n)

		case reflect.Float32, reflect.Float64:
			if err := checkArgLen(data, name); err != nil {
				return fmt.Errorf("args: %s", err)
			}
			n, err := strconv.ParseFloat(data[0], ftype.Bits())
			if err != nil {
				return fmt.Errorf("args: %s", err)
			}
			fval.SetFloat(n)

		case reflect.Slice:
			if err := fillSlice(fval, data); err != nil {
				return err
			}

		default:
			return fmt.Errorf(
				"args: unsupported type: %s", field.Type.Kind().String())
		}
	}
	return nil
}

func fillSlice(v reflect.Value, args []string) error {
	typ := v.Type()
	elem := typ.Elem()
	slice := reflect.MakeSlice(typ, len(args), len(args))
	switch elem.Kind() {
	case reflect.String:
		for i, s := range args {
			slice.Index(i).Set(reflect.ValueOf(s))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i, s := range args {
			n, err := strconv.ParseInt(s, 0, elem.Bits())
			if err != nil {
				return fmt.Errorf("args: %s", err)
			}
			slice.Index(i).SetInt(n)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for i, s := range args {
			n, err := strconv.ParseUint(s, 0, elem.Bits())
			if err != nil {
				return fmt.Errorf("args: %s", err)
			}
			slice.Index(i).SetUint(n)
		}

	case reflect.Float32, reflect.Float64:
		for i, s := range args {
			n, err := strconv.ParseFloat(s, elem.Bits())
			if err != nil {
				return fmt.Errorf("args: %s", err)
			}
			slice.Index(i).SetFloat(n)
		}

	default:
		return fmt.Errorf("args: unsupported slice type %s", typ.Kind().String())
	}

	v.Set(slice)

	return nil
}

// munch items from args, while a new key does not appear.
// return the munched items.
func munchArgs(args []string) []string {
	result := []string{}
	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			result = append(result, args[0])
			args = args[1:]
		} else {
			break
		}
	}
	return result
}

// rawArgsMap parses a command-line string into a map
func rawArgsMap(args []string) (map[string][]string, error) {
	result := make(map[string][]string)
	for len(args) > 0 {
		if strings.HasPrefix(args[0], "--") {
			key := args[0][2:]
			args = args[1:]
			vals := munchArgs(args)
			result[key] = append(result[key], vals...)
			args = args[len(vals):]
		} else if strings.HasPrefix(args[0], "-") {
			if len(args[0]) > 2 {
				// Bunch of switches stuck together
				for i := 2; i < len(args[0]); i++ {
					key := args[0][i-1 : i]
					result[key] = []string{}
				}
				args = args[1:]
			} else {
				key := args[0][1:]
				args = args[1:]
				vals := munchArgs(args)
				result[key] = append(result[key], vals...)
				args = args[len(vals):]
			}
		}
	}
	return result, nil
}

func parseMap(v reflect.Value, args []string) error {
	elem := v.Type().Elem()
	rawData, err := rawArgsMap(args)
	if err != nil {
		return err
	}
	switch elem.Kind() {
	case reflect.String:
		for key, value := range rawData {
			if len(value) == 0 {
				v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(""))
			} else if len(value) > 0 {
				v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value[0]))
			}
			if len(value) > 1 {
				return fmt.Errorf("args: option %q specified more than once", key)
			}
		}

	case reflect.Interface:
		for key, value := range rawData {
			if len(value) == 0 {
				v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(struct{}{}))
			} else if len(value) == 1 {
				val := value[0]
				// try int, then float, otherwise leave as string
				if n, err := strconv.ParseInt(val, 0, 64); err == nil {
					v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(n))
				} else if n, err := strconv.ParseFloat(val, 64); err == nil {
					v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(n))
				} else {
					v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
				}
			} else {
				v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
			}
		}
	default:
		return fmt.Errorf(
			"args: invalid type for unmarshal: %s of %s",
			v.Type().Kind().String(), elem.Kind().String())
	}
	return nil
}

func getType(data interface{}) (reflect.Type, error) {
	typ := reflect.TypeOf(data)
	var err error

	if typ.Kind() != reflect.Ptr {
		return typ, fmt.Errorf("args: non-pointer %s", typ.Kind().String())
	}

	typ = typ.Elem()
	switch typ.Kind() {
	case reflect.Struct, reflect.Slice:
		return typ, nil

	case reflect.Map:
		if keyKind := typ.Key().Kind(); keyKind != reflect.String {
			err = fmt.Errorf(
				"args: map key type must be string, not %s", keyKind.String())
		}
		return typ, err

	default:
		break
	}

	err = fmt.Errorf("args: invalid type for unmarshal: %s", typ.Kind().String())
	return nil, err
}
