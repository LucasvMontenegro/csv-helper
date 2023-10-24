package csvhelper

import (
	"bytes"
	"encoding/csv"
	"errors"
	"reflect"
	"strings"
)

type MarshalConfig struct {
	SkipValidation bool
}

type ValidationConfig struct {
	SkipValidation bool
}

var csvColumnNameTag = "csv_column_name"

var ErrInvalidHeaderSize = errors.New("invalid header size")
var ErrInvalidHeaderValues = errors.New("invalid header values")
var ErrMissingRequiredTag = errors.New("missing required tag")
var ErrDuplicatedTag = errors.New("duplicated tag")
var ErrUninitializedRecords = errors.New("uninitialized records")

type CsvHelper[T any] interface {
	ReadAll(buffer *bytes.Buffer) CsvHelper[T]
	Validate() (bool, error)
	Records() ([][]string, error)
	Error() error
	Marshal(cfg MarshalConfig) ([]T, error)
}

type csvHelperImpl[T any] struct {
	records [][]string
	err     error
}

func New[T any]() CsvHelper[T] {
	return &csvHelperImpl[T]{}
}

func (c csvHelperImpl[T]) ReadAll(buffer *bytes.Buffer) CsvHelper[T] {
	c.records, c.err = csv.NewReader(buffer).ReadAll()
	return c
}

func (c csvHelperImpl[T]) Validate() (bool, error) {
	if err := c.validate(ValidationConfig{}); err != nil {
		return false, err
	}

	return true, nil
}

func (c csvHelperImpl[T]) Records() ([][]string, error) {
	if err := c.validate(ValidationConfig{}); err != nil {
		return nil, err
	}

	return c.records, c.err
}

func (c csvHelperImpl[T]) Error() error {
	if err := c.validate(ValidationConfig{}); err != nil {
		return err
	}

	return c.err
}

func (c csvHelperImpl[T]) Marshal(cfg MarshalConfig) ([]T, error) {
	if err := c.validate(ValidationConfig{cfg.SkipValidation}); err != nil {
		return nil, err
	}

	indexToFieldName, err := c.mapIndexToField(c.records)
	if err != nil {
		return nil, err
	}

	return c.assign(c.records, indexToFieldName)
}

func (c *csvHelperImpl[T]) validate(cfg ValidationConfig) error {
	if err := c.validateIntegrity(); err != nil {
		return err
	}

	if !cfg.SkipValidation {
		if err := c.validateModelTags(); err != nil {
			return err
		}

		var model T
		csvHeaders := c.records[0]
		modelSize := reflect.ValueOf(model).NumField()

		if len(csvHeaders) != modelSize {
			return ErrInvalidHeaderSize
		}
	}

	return nil
}

func (c *csvHelperImpl[T]) validateIntegrity() error {
	if c.err != nil {
		return c.err
	} else {
		if len(c.records) == 0 {
			return ErrUninitializedRecords
		}
	}

	return nil
}

func (*csvHelperImpl[T]) validateModelTags() error {
	var model T
	tags := make(map[string]int)
	modelSize := reflect.ValueOf(model).NumField()

	for i := 0; i < modelSize; i++ {
		tag := reflect.TypeOf(model).Field(i).Tag.Get(csvColumnNameTag)

		if tag == "" {
			return ErrMissingRequiredTag
		} else {
			tags[tag] = i
		}
	}

	if len(tags) != modelSize {
		return ErrDuplicatedTag
	}

	return nil
}

func (csvHelperImpl[T]) mapIndexToField(csv [][]string) (map[int]string, error) {
	var model T
	var modelFieldNames []string
	modelFieldsToColumns := make(map[string]string)
	csvHeaders := csv[0]
	rv := reflect.ValueOf(model)
	rt := reflect.TypeOf(model)
	modelSize := rv.NumField()
	indexToFieldName := make(map[int]string)

	for i := 0; i < modelSize; i++ {
		modelColumnName := rt.Field(i).Tag.Get(csvColumnNameTag)
		modelFieldName := rt.Field(i).Name

		modelFieldNames = append(modelFieldNames, modelFieldName)
		modelFieldsToColumns[modelFieldName] = modelColumnName
	}

	for i := 0; i < len(modelFieldsToColumns); i++ {
		for j, header := range csvHeaders {
			if strings.EqualFold(modelFieldsToColumns[modelFieldNames[i]], header) {
				indexToFieldName[j] = modelFieldNames[i]
			}
		}
	}

	return indexToFieldName, nil
}

func (csvHelperImpl[T]) assign(records [][]string, indexToFieldName map[int]string) ([]T, error) {
	var output []T

	for index, record := range records {
		var model T
		headerLine := 0

		if index == headerLine {
			continue
		}

		for index, data := range record {
			modelField := indexToFieldName[index]
			reflect.ValueOf(&model).Elem().FieldByName(modelField).SetString(data)
		}

		output = append(output, model)
	}

	return output, nil
}
