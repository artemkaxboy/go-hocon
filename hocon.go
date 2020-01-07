package hocon

import (
	"fmt"
	"github.com/go-akka/configuration"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var boolVariants = map[string]bool{
	"true":  true,
	"on":    true,
	"yes":   true,
	"false": false,
	"off":   false,
	"no":    false,
}

type tag struct {
	node            string
	nodeProvided    bool
	path            string
	pathProvided    bool
	defaultValue    string
	defaultProvided bool
}

type fieldWrapper struct {
	inner  *reflect.StructField
	single reflect.Type
}

func (ptr *fieldWrapper) getType() reflect.Type {
	if ptr.inner != nil {
		return ptr.inner.Type
	}
	return ptr.single
}

// getPath is a facade method to call getPath with inner (StructField) or return empty string
// if there is no inner element.
func (ptr *fieldWrapper) getPath(parentPath string) (string, error) {
	if ptr.inner == nil {
		return "", nil
	}
	return getPath(parentPath, ptr.inner)
}

// getPath returns HOCON path for current element.
//
// There are a few methods to set it for each element:
//
// 1. Set path value in struct tag, then it will be taken as is
//
// 2. Set node value in struct tag, then it will be added to the parent path with '.' delimiter
//
// 3. Do not set any tag, then the name of struct field (as is) will be added to the parent path with '.' delimiter
func getPath(parentPath string, field *reflect.StructField) (string, error) {
	tag, err := mapTag(field.Tag)
	if err != nil {
		return "", err
	}

	if tag.pathProvided {
		return tag.path, nil
	}

	if tag.nodeProvided {
		return parentPath + "." + tag.node, nil
	}

	return parentPath + "." + field.Name, nil
}

// LoadConfigFile loads HOCON files parameters to given structure.
func LoadConfigFile(filename string, receiver interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Cannot parse config: Panic")
			panic(r)
		}
	}()
	if err := checkFileAccessibility(filename); err != nil {
		return fmt.Errorf("cannot read configuration file: %w", err)
	}
	config := configuration.LoadConfig(filename)
	return loadConfig(config, receiver)
}

// LoadConfigText parses given text as HOCON and loads parameters to given structure.
func LoadConfigText(text string, receiver interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Cannot parse config: Panic")
			panic(r)
		}
	}()
	config := configuration.ParseString(text)
	return loadConfig(config, receiver)
}

// loadConfig - is an entrypoint to a recursive function which walk through receiver structure to
// find and load needed parameters.
func loadConfig(config *configuration.Config, receiver interface{}) error {
	wrapper := &fieldWrapper{
		single: reflect.ValueOf(receiver).Elem().Type(),
	}
	return loadStruct("", wrapper, reflect.ValueOf(receiver), config)
}

// loadStruct recursively walk through receiver struct nested elements to fill them with the
// config data.
func loadStruct(parentPath string, field *fieldWrapper, fieldValue reflect.Value, config *configuration.Config) error {
	currentPath, err2 := field.getPath(parentPath)
	if err2 != nil {
		return err2
	}

	for i := 0; i < field.getType().NumField(); i++ {
		innerField := field.getType().Field(i)
		if innerField.Type.Kind() == reflect.Struct {
			wrapper := &fieldWrapper{inner: &innerField}
			if err := loadStruct(currentPath, wrapper, fieldValue.Elem().FieldByName(innerField.Name).Addr(), config); err != nil {
				return err
			}
		} else {
			if err := loadValue(currentPath, &innerField, fieldValue.Elem().FieldByName(innerField.Name).Addr(), config); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadValue loads value from config to fieldValue. It's a terminal method for recursive cycle of loadStruct.
func loadValue(parentPath string, field *reflect.StructField, fieldValue reflect.Value, config *configuration.Config) error {
	tag, err := mapTag(field.Tag)
	if err != nil {
		return err
	}

	// it's impossible to get error here while the only way to get it is give an element with incorrect tag and
	// map tag is doing before this statement.
	currentPath, _ := getPath(parentPath, field)

	typ := fieldValue.Elem().Type()

	rawDefault := tag.defaultValue
	if !tag.defaultProvided {
		if !config.HasPath(currentPath) {
			return fmt.Errorf("no value either default value provided for %s [%s]", field.Name, field.Tag)
		}
		rawDefault = getTypeDefault(typ)
	}

	switch typ.Kind() {
	case reflect.Uint:
		return fmt.Errorf("cannot use uint. Use uint32 or uint64 explicitly instead for %s [%s]", field.Name, field.Tag)

	case reflect.Int:
		return fmt.Errorf("cannot use int. Use int32 or int64 explicitly instead for %s [%s]", field.Name, field.Tag)

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		_, err := parseType(typ, rawDefault)
		if err != nil {
			return fmt.Errorf("wrong default value for %s [%s]: %w", field.Name, field.Tag, err)
		}

		stringValue := config.GetString(currentPath, rawDefault)

		value, err := parseType(typ, stringValue)
		if err != nil {
			return fmt.Errorf("wrong value for %s [%s]: %w", field.Name, field.Tag, err)
		}
		fieldValue.Elem().Set(value)

	case reflect.String:
		typedValue := config.GetString(currentPath, rawDefault)
		fieldValue.Elem().SetString(typedValue)
	default:
		return fmt.Errorf("unimplemented data type %s", typ.Kind().String())
	}
	return nil
}

func getBitSizeOf(kind reflect.Kind) int {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return 8
	case reflect.Int16, reflect.Uint16:
		return 16
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 32
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		return 64
	default:
		return 0
	}
}

// parseType parses given string according to given reflect.Type and returns reflect.Value of this type.
func parseType(typ reflect.Type, string string) (reflect.Value, error) {
	kind := typ.Kind()
	switch kind {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(string, 0, getBitSizeOf(kind))
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uintValue).Convert(typ), nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(string, 0, getBitSizeOf(kind))
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(intValue).Convert(typ), nil
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(string, getBitSizeOf(kind))
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(floatValue).Convert(typ), nil
	case reflect.Bool:
		boolValue, ok := boolVariants[strings.ToLower(string)]
		if !ok {
			return reflect.Value{}, fmt.Errorf("for boolean setting use only: true, yes, on, false, no, off values")
		}
		return reflect.ValueOf(boolValue), nil
	}
	return reflect.Value{}, fmt.Errorf("unimplemented Type")
}

func getTypeDefault(typ reflect.Type) string {
	v := reflect.Zero(typ)
	return fmt.Sprintf("%v", reflect.ValueOf(v).Interface())
}

// mapTag parses StructTag to aux Tag struct.
func mapTag(structTag reflect.StructTag) (*tag, error) {
	stringTag := structTag.Get("hocon")
	var tag tag
	if stringTag != "" {
		for _, item := range strings.Split(stringTag, ",") {
			pair := strings.Split(item, "=")
			if len(pair) != 2 {
				return nil, fmt.Errorf("tag format error: %s", stringTag)
			}
			key, value := pair[0], pair[1]

			switch key {
			case "path":
				tag.path = value
				tag.pathProvided = true
			case "default":
				tag.defaultValue = value
				tag.defaultProvided = true
			case "node":
				tag.node = value
				tag.nodeProvided = true
			}
		}
	}
	return &tag, nil
}

// checkFileAccessibility checks if a file accessible and is not a directory before we
// try using it to prevent further errors.
func checkFileAccessibility(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info.Mode()&(1<<8) == 0 {
		return fmt.Errorf("%s permission denied", filename)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}
	return nil
}
