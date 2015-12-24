# gopp
A go preprocessor that adds some object-oriented shortcuts to the go language.

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

	func Construct_(first string, last string) {
		parent::Construct_()
		this.first = first
		this.last = last
	}

	func Type string {
		return "Person"
	}

	func Name string {
		return this.first + " " + this.last
	}
}
```

becomes this in go code:

```
type Thing interface {
	gopp.Base
	WhoAmI () string
	Type () string
	Name () string
}

type Thing_ struct {
	gopp.Base_
}

/**
NewThing creates a new Thing object.
*/
func NewThing() Thing {
	this := Thing_{}
	this.Init_(&this)
	this.Construct_()
	return this.I_().(Thing)
}

func (this *Thing_) WhoAmI() string {
		return this.I_().(Thing).Type() + ":" + this.I_().(Thing).Name()
	}

func (this *Thing_) Type() string {
		return "Uknown"
	}

func (this *Thing_) Name() string {
		return "No Name"
	}


func (this *Thing_) IsA(className string) bool {
	if (className == "Thing") {
		return true
	}
	return this.Base_.IsA(className)
}

func (this *Thing_) Class() string {
	return "Thing"
}


type Person interface {
	Thing
	Construct_ (first string, last string)
	Type () string
	Name () string
}

type Person_ struct {
	Thing_
	first string
	last string
}

/**
NewPerson creates a new Person object.
*/
func NewPerson(first string, last string)  Person {
	this := Person_{}
	this.Init_(&this)
	this.Construct_(first,last)
	return this.I_().(Person)
}

func (this *Person_) Construct_(first string, last string) {
		this.Thing_.Construct_()
		this.first = first
		this.last = last
	}

func (this *Person_) Type() string {
		return "Person"
	}

func (this *Person_) Name() string {
		return this.first + " " + this.last
	}


func (this *Person_) IsA(className string) bool {
	if (className == "Person") {
		return true
	}
	return this.Thing_.IsA(className)
}

func (this *Person_) Class() string {
	return "Person"
}
```

With the expanded go code, you can do the following:

To create a new Person object, just call **NewPerson**. You will be working with the Person interface from then on,
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

This project is currently in its infancy. My hope is that you will find this useful, and also contribute to this project,
including ideas and suggestions, as well as code. At this stage of the project, everything is on the table,
including naming conventions. There are so many directions this could go. See
the open issues for ideas on how to contribute. Also, little attempt has been made to inline document the code, so feel
free to add documentation as you contribute.
