package hocon

import (
	"fmt"
	"github.com/go-akka/configuration"
	"os"
	"reflect"
	"strconv"
	"strings"
)

//todo path for whole containers

type tag struct {
	node            string
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

func (ptr *fieldWrapper) getTag() reflect.StructTag {
	if ptr.inner != nil {
		return ptr.inner.Tag
	}
	return ""
}

// LoadConfigFile loads configuration from given filename as HOCON.
func LoadConfigFile(filename string, receiver interface{}) error {
	if err := checkFileAccessibility(filename); err != nil {
		return fmt.Errorf("cannot read configuration file: %w", err)
	}
	config := configuration.LoadConfig(filename)
	return loadConfig(config, receiver)
}

// LoadConfigText parses given text as HOCON.
func LoadConfigText(text string, receiver interface{}) error {
	config := configuration.ParseString(text)
	return loadConfig(config, receiver)
}

func loadConfig(config *configuration.Config, receiver interface{}) error {
	wrapper := &fieldWrapper{
		single: reflect.ValueOf(receiver).Elem().Type(),
	}
	return loadStruct(wrapper, reflect.ValueOf(receiver), config)
}

func loadStruct(field *fieldWrapper, fieldValue reflect.Value, config *configuration.Config, nodes ...string) error {
	tag, err := mapTag(field.getTag())
	if err != nil {
		return err
	}

	if field.inner != nil {
		newNode := tag.node
		if newNode == "" {
			newNode = field.inner.Name
		}
		nodes = append(nodes, newNode)
	}

	for i := 0; i < field.getType().NumField(); i++ {
		innerField := field.getType().Field(i)
		if innerField.Type.Kind() == reflect.Struct {
			wrapper := &fieldWrapper{inner: &innerField}
			if err := loadStruct(wrapper, fieldValue.Elem().FieldByName(innerField.Name).Addr(), config, nodes...); err != nil {
				return err
			}
		} else {
			if err := loadValue(innerField, fieldValue.Elem().FieldByName(innerField.Name).Addr(), config, nodes...); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadValue loads value from configuration.Config to fieldValue according to its type
func loadValue(field reflect.StructField, fieldValue reflect.Value, config *configuration.Config, nodes ...string) error {
	tag, err := mapTag(field.Tag)
	if err != nil {
		return err
	}

	path := tag.path
	if !tag.pathProvided {
		newNode := tag.node
		if newNode == "" {
			newNode = field.Name
		}
		nodes = append(nodes, newNode)
		path = strings.Join(nodes, ".")
	}

	if !tag.defaultProvided && !config.HasPath(path) {
		return fmt.Errorf("no value either default value provided for %s [%s]", field.Name, field.Tag)
	}
	rawDefault := tag.defaultValue

	kind := fieldValue.Elem().Kind()
	switch kind {
	case reflect.Uint:
		return fmt.Errorf("cannot use %s. Use int32 or int64 instead for %s [%s]", kind, field.Name, field.Tag)

	case reflect.Int:
		return fmt.Errorf("cannot use %s. Use int32 or int64 explicitly instead for %s [%s]", kind, field.Name, field.Tag)

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bitSize := 0
		switch kind {
		case reflect.Uint8:
			bitSize = 8
		case reflect.Uint16:
			bitSize = 16
		case reflect.Uint32:
			bitSize = 32
		case reflect.Uint64:
			bitSize = 64
		}
		if rawDefault == "" {
			rawDefault = "0"
		}
		_, err := strconv.ParseUint(rawDefault, 0, bitSize)
		if err != nil {
			return fmt.Errorf("wrong default value for %s [%s]: %w", field.Name, field.Tag, err)
		}

		stringValue := config.GetString(path, rawDefault)
		typedValue, err := strconv.ParseUint(stringValue, 0, bitSize)
		if err != nil {
			return fmt.Errorf("wrong value for %s [%s]: %w", field.Name, field.Tag, err)
		}
		fieldValue.Elem().SetUint(typedValue)

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bitSize := 0
		switch kind {
		case reflect.Int8:
			bitSize = 8
		case reflect.Int16:
			bitSize = 16
		case reflect.Int32:
			bitSize = 32
		case reflect.Int64:
			bitSize = 64
		}

		if rawDefault == "" {
			rawDefault = "0"
		}
		_, err := strconv.ParseInt(rawDefault, 0, bitSize)
		if err != nil {
			return fmt.Errorf("wrong default value for %s [%s]: %w", field.Name, field.Tag, err)
		}

		stringValue := config.GetString(path, rawDefault)
		typedValue, err := strconv.ParseInt(stringValue, 0, bitSize)
		if err != nil {
			return fmt.Errorf("wrong value for %s [%s]: %w", field.Name, field.Tag, err)
		}
		fieldValue.Elem().SetInt(typedValue)
	case reflect.String:
		typedValue := config.GetString(path, rawDefault)
		fieldValue.Elem().SetString(typedValue)
	default:
		return fmt.Errorf("unimplemented data type %s", kind.String())
	}
	return nil
}

// mapTag parses StructTag to aux Tag struct.
func mapTag(structTag reflect.StructTag) (*tag, error) {
	stringTag := structTag.Get("hocon")
	var tag tag
	if stringTag != "" {
		for _, item := range strings.Split(stringTag, ",") {
			pair := strings.Split(item, "=")
			if len(pair) != 2 {
				return nil, fmt.Errorf("format error: %s", stringTag)
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
