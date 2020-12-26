//
// Copyright © 2020 Anticrm Platform Contributors.
//
// Licensed under the Eclipse Public License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may
// obtain a copy of the License at https://www.eclipse.org/legal/epl-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//

package yar

import "fmt"

func add(vm *VM) Value {
	x := vm.Next().Val()
	y := vm.Next().Val()
	return MakeInt(x + y).Value()
}

func sub(vm *VM) Value {
	x := vm.Next().Val()
	y := vm.Next().Val()
	return MakeInt(x - y).Value()
}

func gt(vm *VM) Value {
	x := vm.Next().Val()
	y := vm.Next().Val()
	return MakeBool(x > y).Value()
}

func blockOfRefinements(vm *VM, code Block) []sym {
	var result []sym

	for i := code.First(vm); i != 0; i = i.Next(vm) {
		w := Word(i.Value(vm))
		result = append(result, w.Sym())
	}

	return result
}

// type pcAlias = pc

func fn(vm *VM) Value {
	params := vm.Next().Block()
	code := vm.Next().Block()

	defaults := blockOfRefinements(vm, params)
	stackSize := len(defaults)

	bind(vm, code, func(sym sym, create bool) Binding {
		for i, def := range defaults {
			if def == sym {
				return MakeStackBinding(i - stackSize)
			}
		}
		return 0
	})

	return makeProc(stackSize, ptr(code.First(vm)))
}

func either(vm *VM) Value {
	cond := vm.Next().Bool()
	ifTrue := vm.Next().Block()
	ifFalse := Block(vm.Next())

	if cond.Val() {
		return vm.call(ifTrue)
	}

	return vm.call(ifFalse)
}

func print(vm *VM) Value {
	val := vm.Next()
	fmt.Printf("PRINT: %s\n", vm.ToString(val))
	return val
}

func _append(vm *VM) Value {
	series := vm.Next().Block()
	value := vm.Next()

	series.Add(vm, value)

	return series.Value()
}

func foreach(vm *VM) Value {
	w := vm.ReadNext().Word()
	series := vm.Next().Block()
	code := vm.Next().Block()

	offset := int(vm.bp)
	bind(vm, code, func(sym sym, create bool) Binding {
		if sym == w.Sym() {
			return MakeWordBinding(offset)
		}
		return 0
	})

	var result Value
	vm.bp++

	for i := series.First(vm); i != 0; i = i.Next(vm) {
		vm.bindStack[offset] = i.Value(vm)
		result = vm.call(code)
	}

	vm.bp = vm.bp - 1

	return result
}

func repeat(vm *VM) Value {
	w := vm.ReadNext().Word()
	times := vm.Next().Integer().Value().Val()
	code := vm.Next().Block()

	offset := int(vm.bp)
	bind(vm, code, func(sym sym, create bool) Binding {
		if sym == w.Sym() {
			return MakeWordBinding(offset)
		}
		return 0
	})

	var result Value
	vm.bp++

	for i := 0; i < times; i++ {
		vm.bindStack[offset] = MakeInt(i).Value()
		result = vm.call(code)
	}

	vm.bp = vm.bp - 1

	return result
}

func makeObject(vm *VM) Value {
	block := vm.Next().Block()

	object := vm.AllocDict()
	bind(vm, block, func(sym sym, create bool) Binding {
		symValPtr := object.Find(vm, sym)
		if symValPtr == 0 {
			if create {
				symValPtr = object.Put(vm, sym, 0)
			} else {
				return 0
			}
		}
		return makeMapBinding(ptr(symValPtr))
	})

	vm.call(block)
	return object.Value()
}

func in(vm *VM) Value {
	m := vm.Next().Dict()
	w := vm.Next().Word()
	sym := w.Sym()

	symval := m.Find(vm, sym)
	if symval == 0 {
		return 0
	}
	binding := makeMapBinding(ptr(symval))
	return Value(_makeWord(sym, pBinding(vm.alloc(cell(binding))), QuoteType))
}

func get(vm *VM) Value {
	w := vm.Next()

	return getWordExec(vm, w)
	// return vm.execFunc[bound.Kind()](vm, bound)
}

func none(vm *VM) Value {
	return 0
}

func CorePackage() *Pkg {
	result := NewPackage("core")
	result.AddFunc("add", add)
	result.AddFunc("sub", sub)
	result.AddFunc("gt", gt)
	result.AddFunc("either", either)
	result.AddFunc("fn", fn)
	result.AddFunc("make-object", makeObject)
	result.AddFunc("foreach", foreach)
	result.AddFunc("print", print)
	result.AddFunc("append", _append)
	result.AddFunc("repeat", repeat)
	result.AddFunc("in", in)
	result.AddFunc("get", get)
	result.AddFunc("none", none)
	return result
}

const coreY = `
add: load-native "core/add"
sub: load-native "core/sub"
gt: load-native "core/gt"
either: load-native "core/either"
fn: load-native "core/fn"
make-object: load-native "core/make-object"
foreach: load-native "core/foreach"
print: load-native "core/print"
append: load-native "core/append"
repeat: load-native "core/repeat"
in: load-native "core/in"
get: load-native "core/get"
none: load-native "core/none"
`

func CoreModule(vm *VM) Value {
	code := vm.Parse(coreY)
	return vm.BindAndExec(code)
}

func BootVM(vm *VM) {
	vm.Library.Add(CorePackage())
	CoreModule(vm)
}
