# Implementing a simple CLASS - Why GOPP

The GO developers have asked for real world examples of large scale codebases running into roadblocks using GO. This is an attempt to explain
why I created GOPP(https://github.com/spekary/gopp/), and the problem it is trying to solve. This isn't an example of a running GO code base, but more a description of
one way to make GO easier to do what I have attempted to do and struggled with in a large code base.

## Porting code to GO

Large code bases don't appear over night. There is a lot of code out in the wild that works great, but its not written in GO. The project
I particular like working on is the PHP QCubed framework. It is unique enough in its features for quickly creating business web sites, 
and its approach to the whole problem of RADD design, that I continue to support it. However, PHP as many have pointed out has its
limitations. The particular limitations that are pet peaves of mine is the lack of compile time type checking, slow speed for data intensive
parts of a web app, and the lack of built-in asynchronous operations. The blocking in particular prevents us from elegently doing live 
updates to a web site like Node can do (there are ugly workarounds, like polling, or 3rd party message casting), and also create limits
to using services like PubNub or FireBase that are inherently asynchronous.

As I have looked around for a solution, the idea of porting to GO came up. GO has a lot of features that fit well with the goals of QCubed.
Also, the qcubed codebase is becoming well structured enough that writing a porting tool would be able to get us very close to pushing
a button and automatically getting GO code, or at least close enough that porting would be infinitely easier.

## Object Oriented??

The one problem though is the paradigm shift of object oriented programming. GO has objects as structs, but lacks many common features available in 
other languages related to object use. Many of these can be worked around, but the one that is a very difficult shift is the lack of 
virtual functions. GO puts virtual functions into interfaces, but not into structs. To change our codebase to fit this model requires
a major rethinking of our approach to the entire problem we are solving. Something the GO developers seem to be happy to encourage,
but is largely too difficult for us to complete in a timely way to justify the porting process.

**However**, it doesn't have to be this way. GO has way of turning a struct's functions into virtual functions by simply creating a matching 
interface. Thus was born GOPP, a simple way to declare that a struct has a matching interface. Its usable, but still somewhat messy
when trying to remember where to use the struct and where to use the interface. It would be SO much more convenient if GO could sort this
out automatically.

## The Proposal

Create a *class* keyword that simply means something is both a struct and an interface. When accessing the methods, it treats it like 
an interface, and when accessing the members, it treats it like a struct. Subclasses could then be created, and superclasses would be able
to call into the virtual functions at will. Perhaps some keywords like "virtual" could be used to further limit which methods are part of
the interface, but you could always just keep it simple.

This would make it SOO much easier to port OO code to GO. And, give us time to rethink parts of our code base later,
