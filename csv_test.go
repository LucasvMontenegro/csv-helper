package csvhelper

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// validModelT struct is valid when each element has is own csv_column_name tag value
type validModelT struct {
	Field1 string `csv_column_name:"column1"`
	Field2 string `csv_column_name:"column2"`
	Field3 string `csv_column_name:"column3"`
}

// randomizedModelT struct is independent for the csv headers order
type randomizedModelT struct {
	Field3 string `csv_column_name:"column3"`
	Field1 string `csv_column_name:"column1"`
	Field2 string `csv_column_name:"column2"`
}

// duplicatedTagModelT struct should have one unique tag value for element
type duplicatedTagModelT struct {
	Field1 string `csv_column_name:"column1"`
	Field2 string `csv_column_name:"column2"`
	Field3 string `csv_column_name:"column2"`
}

// missingTagModelT struct should have the csv_column_name tag
type missingTagModelT struct {
	Field1 string
	Field2 string
	Field3 string
}

// missingElementModelT struct should have the same amount of elements of the valid csv headers
type missingElementModelT struct {
	Field1 string `csv_column_name:"column1"`
	Field2 string `csv_column_name:"column2"`
}

// validCsvFile ordered valid csv
var validCsvFile = `column1,column2,column3
value1,value2,value3
value1,value2,value3
`

// invalidCsvFile invalid over column header
var invalidCsvFile = `column1,column2,column3,invalidColumn
value1,value2,value3
`

// randomizedCsvFile valid csv in a randomized columns order
var randomizedCsvFile = `column2,column1,column3
value1,value2,value3
value1,value2,value3
`

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want CsvHelper[validModelT]
	}{
		{
			name: "success - should return the impl object",
			want: &csvHelperImpl[validModelT]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New[validModelT](); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadAll(t *testing.T) {
	csvHelper := New[validModelT]()
	type args struct {
		buffer *bytes.Buffer
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    csvHelperImpl[validModelT]
	}{
		{
			name: "success - should return the impl object with its records and nil error",
			args: args{
				bytes.NewBufferString(validCsvFile),
			},
			want: csvHelperImpl[validModelT]{
				records: [][]string{
					{
						"column1", "column2", "column3",
					},
					{
						"value1", "value2", "value3",
					},
					{
						"value1", "value2", "value3",
					},
				},
				err: nil,
			},
		},
		{
			name: "error - should return the impl object with empty records and ErrFieldCount error",
			args: args{
				bytes.NewBufferString(invalidCsvFile),
			},
			wantErr: true,
			want: csvHelperImpl[validModelT]{
				err: csv.ErrFieldCount,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := csvHelper.ReadAll(tt.args.buffer); tt.wantErr {
				rec, err := got.Records()
				assert.Equal(t, rec, tt.want.records)
				assert.ErrorIs(t, err, tt.want.err)
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ReadAll() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// T is validModelT
func TestValidate1(t *testing.T) {
	var csvHelper CsvHelper[validModelT]
	tests := []struct {
		name           string
		want           bool
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name:    "error - fail to validate because an error exists - *should readAll again",
			want:    false,
			wantErr: true,
			err:     csv.ErrFieldCount,
			before: func() {
				csvHelper = New[validModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(invalidCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.ErrorIs(t, gotErr, err)
			},
		},
		{
			name:    "error - fail to validate because the csv wasn't read yet",
			want:    false,
			wantErr: true,
			err:     ErrUninitializedRecords,
			before: func() {
				csvHelper = New[validModelT]()
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.Equal(t, gotErr, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()

			got, err := csvHelper.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
			tt.assertBehavior(t, err, tt.err)
		})
	}
}

// T is missingTagModelT
func TestValidate2(t *testing.T) {
	var csvHelper CsvHelper[missingTagModelT]
	tests := []struct {
		name           string
		want           bool
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name:    "error - fail to validate because model T is misconfigured - missing tag",
			want:    false,
			wantErr: true,
			err:     ErrMissingRequiredTag,
			before: func() {
				csvHelper = New[missingTagModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(validCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.Equal(t, gotErr, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()

			got, err := csvHelper.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
			tt.assertBehavior(t, err, tt.err)
		})
	}
}

// T is duplicatedTagModelT
func TestValidate3(t *testing.T) {
	var csvHelper CsvHelper[duplicatedTagModelT]
	tests := []struct {
		name           string
		want           bool
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name:    "error - fail to validate because model T is misconfigured - duplicated tag",
			want:    false,
			wantErr: true,
			err:     ErrDuplicatedTag,
			before: func() {
				csvHelper = New[duplicatedTagModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(validCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.Equal(t, gotErr, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()

			got, err := csvHelper.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
			tt.assertBehavior(t, err, tt.err)
		})
	}
}

// T is missingElementModelT
func TestValidate4(t *testing.T) {
	var csvHelper CsvHelper[missingElementModelT]
	tests := []struct {
		name           string
		want           bool
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name:    "error - fail to validate because model has different amount of fields than csv header",
			want:    false,
			wantErr: true,
			err:     ErrInvalidHeaderSize,
			before: func() {
				csvHelper = New[missingElementModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(validCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.Equal(t, gotErr, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()

			got, err := csvHelper.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
			tt.assertBehavior(t, err, tt.err)
		})
	}
}

func TestRecords(t *testing.T) {
	var csvHelper CsvHelper[validModelT]
	tests := []struct {
		name           string
		want           [][]string
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name: "error - should return Uninitialized records",
			// want: Uninitialized - zero value
			wantErr: true,
			err:     ErrUninitializedRecords,
			before: func() {
				csvHelper = New[validModelT]()
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.ErrorIs(t, gotErr, err)
			},
		},
		{
			name: "ok - should return the records",
			want: [][]string{
				{
					"column1", "column2", "column3",
				},
				{
					"value1", "value2", "value3",
				},
				{
					"value1", "value2", "value3",
				},
			}, wantErr: false,
			err: nil,
			before: func() {
				csvHelper = New[validModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(validCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.ErrorIs(t, gotErr, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			got, err := csvHelper.Records()
			if (err != nil) != tt.wantErr {
				t.Errorf("Records() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Records() = %v, want %v", got, tt.want)
			}
			tt.assertBehavior(t, err, tt.err)
		})
	}
}

func TestError(t *testing.T) {
	var csvHelper CsvHelper[validModelT]

	tests := []struct {
		name           string
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name:    "error - should return Uninitialized records",
			wantErr: true,
			err:     ErrUninitializedRecords,
			before: func() {
				csvHelper = New[validModelT]()
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.ErrorIs(t, gotErr, err)
			},
		},
		{
			name:    "error - should return Uninitialized records",
			wantErr: true,
			err:     csv.ErrFieldCount,
			before: func() {
				csvHelper = New[validModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(invalidCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.ErrorIs(t, gotErr, err)
			},
		},
		{
			name:    "ok - should return nil",
			wantErr: false,
			err:     nil,
			before: func() {
				csvHelper = New[validModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(validCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr error, err error) {
				assert.ErrorIs(t, gotErr, err)
			},
		},
	}
	for _, tt := range tests {
		tt.before()
		t.Run(tt.name, func(t *testing.T) {
			err := csvHelper.Error()

			if (err != nil) != tt.wantErr {
				t.Errorf("Error() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.assertBehavior(t, err, tt.err)
		})
	}
}

func TestMarshal(t *testing.T) {
	var csvHelper CsvHelper[validModelT]

	type args struct {
		cfg MarshalConfig
	}
	tests := []struct {
		name           string
		args           args
		want           []validModelT
		wantErr        bool
		err            error
		before         func()
		assertBehavior func(t *testing.T, gotErr error, err error)
	}{
		{
			name: "error - records uninitialized - should be read before marshaling",
			args: args{
				cfg: MarshalConfig{SkipValidation: true}, // in this test case it could be false too
			},
			// want:    uninitialized - zero value,
			wantErr: true,
			err:     ErrUninitializedRecords,
			before: func() {
				csvHelper = New[validModelT]()
			},
			assertBehavior: func(t *testing.T, gotErr, err error) {
			},
		},
		{
			name: "ok - should return the fulfilled model",
			args: args{
				cfg: MarshalConfig{SkipValidation: false},
			},
			want: []validModelT{
				{Field1: "value1", Field2: "value2", Field3: "value3"},
				{Field1: "value1", Field2: "value2", Field3: "value3"},
			},
			wantErr: false,
			err:     nil,
			before: func() {
				csvHelper = New[validModelT]()
				csvHelper = csvHelper.ReadAll(bytes.NewBufferString(validCsvFile))
			},
			assertBehavior: func(t *testing.T, gotErr, err error) {
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			got, err := csvHelper.Marshal(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
			tt.assertBehavior(t, err, tt.err)
		})
	}
}
