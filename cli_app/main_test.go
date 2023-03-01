package main

import (
	"testing"
)

// func Test_(t *testing.T) {

// }
func Test_parsing(t *testing.T) {
	c, err := parseCmd("[1] ADD,userid,2000")
	if err != nil {
		t.Fatal(err)
	}
	if c.Command != "ADD" || c.Ticket != 1 {
		t.Fatal(err)
	}
}
