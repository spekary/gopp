/**
gopp

Installation:

go get github.com/spekary/gopp/gopp

Gopp some lightweight object oriented features to go code that allow you to structure your code using
inheritance, and use virtual functions.

It is currently structured as a pre-processor that can be run with go-generate, and that takes in a .gpp file and
outputs standard .go code.

The primary feature is a new keyword "class", that operates similarly to the "class" keyword in other OO languages, and begins the
definition of an object. This single object definition gets converted to both a struct and interface in standard go
code such that you can do things like:

this.DoVirtualFunc()

or

this.member = 7

and the resulting code will find the correct function or member of the struct and use it, including functions of subclasses.
The result is the ability for programmers who are accustomed to writing object oriented code like Java, PHP or C++ to work the way they like, and
not have to worry about all the implementation details of writing similar code in strict go.

The effect is a very lightweight, bu

Syntax:

A class definition begins with the word "class", followed by a class name, the required word "extends" and a superclass.

To create a base class, you should extend the "gopp.Base" class. The Base class is a struct and interface combination that
implement basic object functions that are often found in object oriented languages.

Within a class, declare members the same way you would declare a member of a go struct, with a name followed by a type.

For example:
	member1 string
	member2 int

Then declare methods just like declaring functions in go. Do not include the target selector, that will be added automatically.

Within a method, use the 'this' keyword to refer to the current object. Whenever you refer to a a member of the object, the
.gpp pre-processor will get the member of the struct. If you call a method in the object, the preprocessor will call a
method on the interface to the object, so that if any subclasses override that method, the subclass method will be called.
The overall effect is similar to any other object oriented language, and you really don't need to worry about the details,
except of course when debugging, since you will be debugging the go code and not the gpp code.

See the files in the test directory for more examples of what you can do, and the results.

Special Key Words and Transformations:

The Construct() function is an initializer like any other constructor. Call the parent constructor from within your constructor using:

parent::Construct(<vars>)

parent:: will always get substituted by the class after the "extends" keyword.

IsA() and Class() functions are automatically added so you can test whether a particular object belongs to a class hierarchy
or is a particular class without having to do type juggling or reflection.

The resulting struct name is the same as the class name, and the interface name is the class name followed by "I". So,
a Duck interface is DuckI (pronounced Duckee, as in "duck like"). Generally you will work with the interface when
creating these objects from standard Go code.

*/
package main
