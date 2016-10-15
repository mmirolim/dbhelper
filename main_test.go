package main

import (
	"fmt"
	"go/format"
	"os"
	"testing"
)

var (
	input = "package user; type User struct {ID int `db:\"id\"`; Name string `db:\"nickname\"`}; type Person struct{Fname string `db:\"fname\"`}"

	expectedOutput = `
// generated by sqlhelp; DO NOT EDIT
package user

type _PersonColumn struct {
    Fname string
}

type _UserColumn struct{
    ID string

    Name string
}

var (
    PersonColumns _PersonColumn

    UserColumns _UserColumn
)

func init() {

    // define Person column names

    PersonColumns.Fname = "fname"

    // define User column names

    UserColumns.ID = "id"

    UserColumns.Name = "nickname"

}
`
)

func TestGenColumnStruct(t *testing.T) {
	out, err := genColumnStruct("db", "db", input, []string{"User", "Person"})
	if err != nil {
		t.Errorf("unexpected error %v with input %v", err, input)
	}

	if string(out) != expectedOutput {
		t.Errorf("expected %v got %v", expectedOutput, string(out))
	}

}

func TestMapTags(t *testing.T) {
	tags := []struct {
		input    string
		expected map[string]string
	}{
		{
			input: "db:\"id\"",
			expected: map[string]string{
				"db": "id",
			},
		},

		{
			input: "db:\"id\" json:\"json_id\"",
			expected: map[string]string{
				"db":   "id",
				"json": "json_id",
			},
		},
	}

	for _, tag := range tags {
		for k := range tag.expected {
			val, ok := tagLookup(tag.input, k)
			if !ok || tag.expected[k] != val {
				t.Errorf("expected %v got %v for tag %v", tag.expected[k], val, k)
			}
		}
	}
}

func TestMain(m *testing.M) {
	// format expected output
	src, err := format.Source([]byte(expectedOutput))
	if err != nil {
		fmt.Printf("fmt err %+v\n", err)
		os.Exit(1)
	}
	expectedOutput = string(src)
	os.Exit(m.Run())
}
