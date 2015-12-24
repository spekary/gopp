/**
gopp

Gopp or go-plus-plus, is an attempt to make object oriented features of go, like virtual functions, more accessible.

It is currently structured as a pre-processor that can be run with go-generate, and that takes in a .gpp file and
outputs standard .go code.

The primary feature is a new keyword "class", that operates similarly to the "class" keyword in PHP, and begins the
definition of an object. This single object definition gets converted to both a struct and interface in standard go
code such that you can do things like:

this.DoVirtualFunc()

or

this.member = 7

and the resulting code will find the correct function or member of the struct and use it, including functions of subclasses.
The result is the ability for programmers who are accustomed to writing object oriented code like PHP or C++ to work the way they like, and
not have to worry about all the implementation details of writing similar code in strict go.

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

 */
package main
