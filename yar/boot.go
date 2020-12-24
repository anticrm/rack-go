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

func add(vm *VM) Value {
	x := intValue(vm.Next())
	y := intValue(vm.Next())
	return makeInt(x + y)
}

func sub(vm *VM) Value {
	x := intValue(vm.Next())
	y := intValue(vm.Next())
	return makeInt(x - y)
}

func gt(vm *VM) Value {
	x := intValue(vm.Next())
	y := intValue(vm.Next())
	return makeBool(x > y)
}

func blockOfRefinements(vm *VM, code Block) []sym {
	var result []sym

	for i := code.First(); i != 0; i = i.Next(vm) {
		w := Word(i.Value(vm))
		result = append(result, w.Sym())
	}

	return result
}

// type pcAlias = pc

func fn(vm *VM) Value {
	params := Block(vm.Next())
	code := Block(vm.Next())

	defaults := blockOfRefinements(vm, params)
	stackSize := len(defaults)

	bind(vm, code, func(sym sym, create bool) bound {
		for i, def := range defaults {
			if def == sym {
				return makeStackBinding(i - stackSize)
			}
		}
		return 0
	})

	return makeProc(stackSize, ptr(code.First()))
}

func either(vm *VM) Value {
	cond := vm.Next()
	ifTrue := Block(vm.Next())
	ifFalse := Block(vm.Next())

	if boolValue(cond) {
		return vm.call(ifTrue)
	}

	return vm.call(ifFalse)
}

func makeObject(vm *VM) Value {
	block := Block(vm.Next())

	object := vm.allocDict()

	bind(vm, block, func(sym sym, create bool) bound {
		symValPtr := object.find(vm, sym)
		if symValPtr == 0 {
			if create {
				object.put(vm, sym, 0)
				symValPtr = object.find(vm, sym) // TODO: fix this garbage
			} else {
				return 0
			}
		}
		return makeMapBinding(ptr(symValPtr))
	})

	vm.call(block)
	return Value(vm.read(ptr(object)))
}

func BootVM(vm *VM) {
	vm.AddNative("add", add)
	vm.AddNative("sub", sub)
	vm.AddNative("gt", gt)
	vm.AddNative("either", either)
	vm.AddNative("fn", fn)
	vm.AddNative("make-object", makeObject)
}
