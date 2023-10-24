# Welcome to csv-helper 👋

> Convert csv files to structs.

# How to use

<br />

## Install the module
```
go get github.com/lucasvmontenegro/csv-helper.git
```

<br />

## Import the lib
``` go
import (
	csvhelper "github.com/lucasvmontenegro/csv-helper.git"
)
```

<br />

## Create the struct type you want to fill using the TAG to reference the csv columns

``` go
type person struct {
	Name     string `csv_column_name:"name"`
	LastName string `csv_column_name:"lastname"`
}
```

<br />

## Instantiate the helper by passing the type as a parameter
``` go
helper := csvhelper.New[person]()
```

<br />

## [Optional] create a .csv file in your local project
name:
file.csv

content:
``` csv
name,lastname
Lucas,Montenegro
Vinicius,Vieira
```
<br />

## [Optional] Read the csv file
``` go
file, err := ioutil.ReadFile("...file.csv")
	if err != nil {
        // handle err
		fmt.Println("err reading file")
		return
	}

file = bytes.NewBuffer(file)
```

<br />

## [Optional] Or simulates one manually in memory
``` go
file := `name,lastname
Lucas,Montenegro
Vinicius,Vieira
`

file = bytes.NewBufferString(file)
```

<br />

## Use the helper to Read the file or the variable and fill the model with their lines
``` go
models, err := helper.
ReadAll(file).
Marshal(csvhelper.MarshalConfig{})
if err != nil {
    // handle err
    fmt.Println("err reading/marshalling all")
    return
}
```

## About the methods
The helper exposes the following interface

``` go
type CsvHelper[T any] interface {
	ReadAll(buffer *bytes.Buffer) CsvHelper[T]
	Validate() (bool, error)
	Records() ([][]string, error)
	Error() error
	Marshal(cfg MarshalConfig) ([]T, error)
}
```

### ReadAll
The first method to be called. It uses a reader to convert the csv buffer to a map of records.

###	Validate
Verifies if some error happened in the ReadAll method, or in any other that was called before. Also validates the type inserted, looking for its tags and the amount of columns compared to the fields in the given struct type.

### Records
Returns the read records in map format.

### Error
Returns the error if it exists.

### Marshal
Returns a list of the given struct type.
# csv-helper
# csv-helper
