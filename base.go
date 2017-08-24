package gopp

type BaseI interface {
	IsA(className string) bool
	InstanceOf(className string) bool
	Class() string
}

type Base struct {
	_i BaseI // the internal copy of the subclass struct inside of an interface
}

// Init saves a copy of the subclass inside of an interface. It is to be called by the "New" function of the subclass.
// We can type cast this interface at will to any intermediate class in order to virtually use its methods. You should
// not normally call this function yourself.
func (b *Base) Init(i BaseI) {
	b._i = i
}

// Construct is a typical constructor that can be used to initialize an inheritance hierarchy. Subclasses should call their
// superclasses.
func (b *Base) Construct() {
}

// I() is used internally to return the interface so that its methods can be called virtually.
func (b *Base) I() BaseI {
	return b._i
}

// InstanceOf is a synonym for IsA. b can also be used to test the virtual calling capabilities of the class system.
func (b *Base) InstanceOf(className string) bool {
	return b._i.IsA(className)
}

// IsA returns true if the object is a type that corresponds to the given type name. This will search all intermediate types
// as well.
func (b *Base) IsA(className string) bool {
	if className == "gopp.Base" {
		return true
	}
	return false
}

// Class returns the name of the class itself
func (b *Base) Class() string {
	return "gopp.Base"
}
