package main

import (
	"testing"
)

func TestText(t *testing.T) {
	s :=
		`
/**
Test comment
 */
`
	sNew := ProcessString(s)
	if sNew != s {
		t.Error("Comment not passed through: " + sNew)
	}
}

func TestText2(t *testing.T) {
	s :=
		`
/**
Test comment
 */

 type test struct  {
 	blah
 }
`
	sNew := ProcessString(s)
	if sNew != s {
		t.Error("Text not passed through: " + sNew)
	}
}

func TestSimple(t *testing.T) {
	s :=
		`
class Test extends Base {
	testMember string

	func _construct() {
		this.Base._construct()
		this.testMember = "hi"
	}
}
`
	sExpected :=
		`
type Test interface {
	Base
}

type Test_ struct {
	Base_
	this Test
	testMember string
}

func NewTest() Test {
	s := Test_{}
	s.this := Test(&s)
	s._construct()
	return i
}
`
	sNew := ProcessString(s)
	if sNew != sExpected {
		t.Error("struct with member not created: " + sNew)
	}
}

func TestJustMember(t *testing.T) {
	s :=
		`
class Test extends Base {
	testMember string
}
`
	sExpected :=
		`
type Test interface {
	Base
}

type _Test struct {
	_Base
	testMember string
}
`
	sNew := ProcessString(s)
	if sNew != sExpected {
		t.Error("struct with member not created: " + sNew)
	}
}
