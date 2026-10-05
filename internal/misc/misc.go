package internal_misc

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

func ConvertArgument(value any, expectedType reflect.Type) (converted any, err error) {

	if expectedType == nil {
		return nil, errors.New("expected type is nil")
	}

	if value == nil {
		return nil, errors.New("argument value is nil")
	}

	var actualType reflect.Type = reflect.TypeOf(value)

	// Значение уже имеет нужный тип.
	if actualType.AssignableTo(expectedType) {
		return value, nil
	}

	var encoded []byte

	encoded, err = yaml.Marshal(value)
	if err != nil {
		return nil, err
	}

	// Создаём указатель на значение ожидаемого типа.
	var destination reflect.Value = reflect.New(expectedType)

	err = yaml.Unmarshal(
		encoded,
		destination.Interface(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot convert %s to %s: %w",
			actualType,
			expectedType,
			err,
		)
	}

	return destination.Elem().Interface(), nil
}

func ReadFromFile(path string) ([]byte, error) {
	var (
		fileData []byte
		err      error
	)
	fileData, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return fileData, nil
}

func ReadStringFromFile(path string) (string, error) {
	var (
		content  string
		err      error
		fileData []byte
	)
	fileData, err = ReadFromFile(path)
	if err != nil {
		return "", err
	}
	content = string(fileData)
	return content, nil
}
