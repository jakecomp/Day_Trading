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
	if c.command != "ADD" || c.ticket != 1 {
		t.Fatal(err)
	}
}
