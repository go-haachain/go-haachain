package otto

import (
	"fmt"
)

// ======
// _stash
// ======

type _stash interface {
	hasBinding(string) bool            //
	createBinding(string, bool, Value) // CreateMutableBinding
	setBinding(string, Value, bool)    // SetMutableBinding
	getBinding(string, bool) Value     // GetBindingValue
	deleteBinding(string) bool         //
	setValue(string, Value, bool)      // createBinding + setBinding

	outer() _stash
	runtime() *_runtime

	newReference(string, bool, _at) _reference

	clone(clone *_clone) _stash
}

// ==========
// _objechaaash
// ==========

type _objechaaash struct {
	_runtime *_runtime
	_outer   _stash
	object   *_object
}

func (self *_objechaaash) runtime() *_runtime {
	return self._runtime
}

func (runtime *_runtime) newObjechaaash(object *_object, outer _stash) *_objechaaash {
	if object == nil {
		object = runtime.newBaseObject()
		object.class = "environment"
	}
	return &_objechaaash{
		_runtime: runtime,
		_outer:   outer,
		object:   object,
	}
}

func (in *_objechaaash) clone(clone *_clone) _stash {
	out, exists := clone.objechaaash(in)
	if exists {
		return out
	}
	*out = _objechaaash{
		clone.runtime,
		clone.stash(in._outer),
		clone.object(in.object),
	}
	return out
}

func (self *_objechaaash) hasBinding(name string) bool {
	return self.object.hasProperty(name)
}

func (self *_objechaaash) createBinding(name string, deletable bool, value Value) {
	if self.object.hasProperty(name) {
		panic(hereBeDragons())
	}
	mode := _propertyMode(0111)
	if !deletable {
		mode = _propertyMode(0110)
	}
	// TODO False?
	self.object.defineProperty(name, value, mode, false)
}

func (self *_objechaaash) setBinding(name string, value Value, strict bool) {
	self.object.put(name, value, strict)
}

func (self *_objechaaash) setValue(name string, value Value, throw bool) {
	if !self.hasBinding(name) {
		self.createBinding(name, true, value) // Configurable by default
	} else {
		self.setBinding(name, value, throw)
	}
}

func (self *_objechaaash) getBinding(name string, throw bool) Value {
	if self.object.hasProperty(name) {
		return self.object.get(name)
	}
	if throw { // strict?
		panic(self._runtime.panicReferenceError("Not Defined", name))
	}
	return Value{}
}

func (self *_objechaaash) deleteBinding(name string) bool {
	return self.object.delete(name, false)
}

func (self *_objechaaash) outer() _stash {
	return self._outer
}

func (self *_objechaaash) newReference(name string, strict bool, at _at) _reference {
	return newPropertyReference(self._runtime, self.object, name, strict, at)
}

// =========
// _dclStash
// =========

type _dclStash struct {
	_runtime *_runtime
	_outer   _stash
	property map[string]_dclProperty
}

type _dclProperty struct {
	value     Value
	mutable   bool
	deletable bool
	readable  bool
}

func (runtime *_runtime) newDeclarationStash(outer _stash) *_dclStash {
	return &_dclStash{
		_runtime: runtime,
		_outer:   outer,
		property: map[string]_dclProperty{},
	}
}

func (in *_dclStash) clone(clone *_clone) _stash {
	out, exists := clone.dclStash(in)
	if exists {
		return out
	}
	property := make(map[string]_dclProperty, len(in.property))
	for index, value := range in.property {
		property[index] = clone.dclProperty(value)
	}
	*out = _dclStash{
		clone.runtime,
		clone.stash(in._outer),
		property,
	}
	return out
}

func (self *_dclStash) hasBinding(name string) bool {
	_, exists := self.property[name]
	return exists
}

func (self *_dclStash) runtime() *_runtime {
	return self._runtime
}

func (self *_dclStash) createBinding(name string, deletable bool, value Value) {
	_, exists := self.property[name]
	if exists {
		panic(fmt.Errorf("createBinding: %s: already exists", name))
	}
	self.property[name] = _dclProperty{
		value:     value,
		mutable:   true,
		deletable: deletable,
		readable:  false,
	}
}

func (self *_dclStash) setBinding(name string, value Value, strict bool) {
	property, exists := self.property[name]
	if !exists {
		panic(fmt.Errorf("setBinding: %s: missing", name))
	}
	if property.mutable {
		property.value = value
		self.property[name] = property
	} else {
		self._runtime.typeErrorResult(strict)
	}
}

func (self *_dclStash) setValue(name string, value Value, throw bool) {
	if !self.hasBinding(name) {
		self.createBinding(name, false, value) // NOT deletable by default
	} else {
		self.setBinding(name, value, throw)
	}
}

// FIXME This is called a __lot__
func (self *_dclStash) getBinding(name string, throw bool) Value {
	property, exists := self.property[name]
	if !exists {
		panic(fmt.Errorf("getBinding: %s: missing", name))
	}
	if !property.mutable && !property.readable {
		if throw { // strict?
			panic(self._runtime.panicTypeError())
		}
		return Value{}
	}
	return property.value
}

func (self *_dclStash) deleteBinding(name string) bool {
	property, exists := self.property[name]
	if !exists {
		return true
	}
	if !property.deletable {
		return false
	}
	delete(self.property, name)
	return true
}

func (self *_dclStash) outer() _stash {
	return self._outer
}

func (self *_dclStash) newReference(name string, strict bool, _ _at) _reference {
	return &_stashReference{
		name: name,
		base: self,
	}
}

// ========
// _fnStash
// ========

type _fnStash struct {
	_dclStash
	arguments           *_object
	indexOfArgumentName map[string]string
}

func (runtime *_runtime) newFunctionStash(outer _stash) *_fnStash {
	return &_fnStash{
		_dclStash: _dclStash{
			_runtime: runtime,
			_outer:   outer,
			property: map[string]_dclProperty{},
		},
	}
}

func (in *_fnStash) clone(clone *_clone) _stash {
	out, exists := clone.fnStash(in)
	if exists {
		return out
	}
	dclStash := in._dclStash.clone(clone).(*_dclStash)
	index := make(map[string]string, len(in.indexOfArgumentName))
	for name, value := range in.indexOfArgumentName {
		index[name] = value
	}
	*out = _fnStash{
		_dclStash:           *dclStash,
		arguments:           clone.object(in.arguments),
		indexOfArgumentName: index,
	}
	return out
}

func gehaaashProperties(stash _stash) (keys []string) {
	switch vars := stash.(type) {
	case *_dclStash:
		for k := range vars.property {
			keys = append(keys, k)
		}
	case *_fnStash:
		for k := range vars.property {
			keys = append(keys, k)
		}
	case *_objechaaash:
		for k := range vars.object.property {
			keys = append(keys, k)
		}
	default:
		panic("unknown stash type")
	}

	return
}
