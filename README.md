# gopp
A go preprocessor that adds lightweight object-oriented shortcuts to the go language.

## Install

```shell
go get github.com/spekary/gopp/gopp
```

## About Gopp
The go authors separated the interface of a typical object from its implementation. This is great for some applications, but
for others, it can be a headache to put them back together. Gopp saves you a lot of duplicate typing and keeping
track of inheritance structures to make the whole process of creating hierarchical objects a lot easier.

In particular, gopp makes it easier to write your objects to support virtual function calls, something that is possible
in go, but kind of painful to actually code.

Gopp operates on .gpp files and converts them to standard .go files. .gpp files are formatted similarly to .go files,
but have a few additional key words. Here is an example.

### Gopp Object

The following .gpp code:

```
class Thing extends gopp.Base {

	func WhoAmI() string {
		return this.Type() + ":" + this.Name()
	}

	func Type string {
		return "Uknown"
	}

	func Name string {
		return "No Name"
	}

}

class Person extends Thing {
	first string
	last string

	func Construct(first string, last string) {
		parent::Construct()
		this.first = first
		this.last = last
	}

	override func Type string {
		return "Person"
	}

	override func Name string {
		return this.first + " " + this.last
	}
}
```

becomes this in go code:

```go
type ThingI interface {
	gopp.BaseI

	WhoAmI() string
	Type() string
	Name() string
}

type Thing struct {
	gopp.Base
}

// New Thing creates a new Thing object and returns its matching interface
func NewThing() ThingI {
	t_ := Thing{}
	t_.Init(&t_)
	t_.Construct()
	return t_.I().(ThingI)
}

func (t_ *Thing) WhoAmI() string {
	return t_.I().(ThingI).Type() + ":" + t_.I().(ThingI).Name()
}

func (t_ *Thing) Type() string {
	return "Uknown"
}

func (t_ *Thing) Name() string {
	return "No Name"
}

func (t_ *Thing) IsA(className string) bool {
	if className == "Thing" {
		return true
	}
	return t_.Base.IsA(className)
}

func (t_ *Thing) Class() string {
	return "Thing"
}

type PersonI interface {
	ThingI

	ComplexReturn(data interface{}) (string, interface{})
}

type Person struct {
	Thing
	first string
	last  string
}

// New Person creates a new Person object and returns its matching interface
func NewPerson(first string, last string) PersonI {
	p_ := Person{}
	p_.Init(&p_)
	p_.Construct(first, last)
	return p_.I().(PersonI)
}

func (p_ *Person) Construct(first string, last string) {
	p_.Thing.Construct()
	p_.first = first
	p_.last = last
}

func (p_ *Person) Type() string {
	return "Person"
}

func (p_ *Person) Name() string {
	return p_.first + " " + p_.last
}

func (p_ *Person) ComplexReturn(data interface{}) (string, interface{}) {
	return p_.first + " " + p_.last, 1
}

func (p_ *Person) IsA(className string) bool {
	if className == "Person" {
		return true
	}
	return p_.Thing.IsA(className)
}

func (p_ *Person) Class() string {
	return "Person"
}
```

With the expanded go code, you can do the following:

To create a new Person object, just call **NewPerson**. You will be working with the PersonI interface from then on,
and you do not need to worry about implementation details of the object.

Once you have a Person object, you can call WhoAmI(), and it will do the right thing and call into the Person's Type()
and Name() functions virtually.

See the doc for more specifics.

## Base file
Gopp includes a base file that provides some reflection capabilities and basic features to every object. All objects
should extend from another object you create, or the gopp.Base object.

## Usage

gopp file1 file2.. | -all

Either specify the specific files you want to gopp, or the -all flag will grab all .gpp files in the current directory

## Notes From the Author

While I realize this approach is going to be frowned upon by the go "community", object-oriented
patterns actually do have a purpose in the world. 
