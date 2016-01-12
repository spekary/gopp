package gopp

type Base interface {
	IsA(className string) bool
	InstanceOf(className string) bool
	Class() string
}

type Base_ struct {
	_i Base	// the internal copy of the subclass struct inside of an interface
}

/**
Init saves a copy of the subclass inside of an interface. It is to be called by the "New" function of the subclass.
We can type cast this interface at will to any intermediate class in order to virtually use its methods. You should
not normally call this function yourself.
*/
func (this *Base_) Init_(i Base) {
	this._i = i
}

/**
Construct is a typical constructor that can be used to initialize an inheritance hierarchy. Subclasses should call their
superclasses.
 */
func (this *Base_) Construct_() {
}

/**
I_ is used internally to return the interface so that its methods can be called virtually.
 */
func (this *Base_) I_() Base {
	return this._i
}

/**
InstanceOf is a synonym for IsA. This can also be used to test the virtual calling capabilities of the class system.
 */
func (this *Base_) InstanceOf(className string) bool {
	return this._i.IsA(className)
}

/**
IsA returns true if the object is a type that corresponds to the given type name. This will search all intermediate types
as well.
 */
func (this *Base_) IsA(className string) bool {
	if (className == "gopp.Base") {
		return true
	}
	return false
}

func (this *Base_) Class() string {
	return "gopp.Base"
}