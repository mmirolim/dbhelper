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

func TestMain(m *testing.M) {
	// preformat expected output
	src, err := format.Source([]byte(expectedOutput))
	if err != nil {
		fmt.Printf("fmt err %+v\n", err)
		os.Exit(1)
	}
	expectedOutput = string(src)
	os.Exit(m.Run())
}