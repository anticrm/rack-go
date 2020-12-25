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
	fmt.Printf("PRINT: %s\n", vm.toString(val))
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

	bind(vm, code, func(sym sym, create bool) Binding {
		if sym == w.Sym() {
			return MakeStackBinding(-1)
		}
		return 0
	})

	var result Value

	fmt.Printf("series %s\n", vm.toString(series.Value()))

	for i := series.First(vm); i != 0; i = i.Next(vm) {
		val := i.Value(vm)
		fmt.Printf("value: %016x\n", val)
		vm.Push(val)
		result = vm.call(code)
		vm.sp = vm.sp - 1
	}

	return result
}

func makeObject(vm *VM) Value {
	block := Block(vm.Next())

	object := vm.AllocDict()

	bind(vm, block, func(sym sym, create bool) Binding {
		symValPtr := object.Find(vm, sym)
		if symValPtr == 0 {
			if create {
				symValPtr = object.Put(vm, sym, 0)
				// symValPtr = object.Find(vm, sym) // TODO: fix this garbage
			} else {
				return 0
			}
		}
		return makeMapBinding(ptr(symValPtr))
	})

	vm.call(block)
	return object.Value()
}

// func get(vm *VM) Value {
// 	w := vm.Next().Word()
// 	vm.bindFunc[GetWordType](vm, )
// }

// var (
// 	coreLibrary = map[string]procFunc{
// 		"add":         add,
// 		"sub":         sub,
// 		"gt":          gt,
// 		"either":      either,
// 		"fn":          fn,
// 		"make-object": makeObject,
// 		"foreach":     foreach,
// 		"print":       print,
// 	}
// )

func corePackage() *Pkg {
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
	return result
}

func BootVM(vm *VM) {
	vm.LoadPackage(corePackage(), vm.Dictionary)
}
