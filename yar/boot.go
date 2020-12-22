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

func add(pc *PC) Value {
	x := intValue(pc.Next())
	y := intValue(pc.Next())
	return makeInt(x + y)
}

func sub(pc *PC) Value {
	x := intValue(pc.Next())
	y := intValue(pc.Next())
	return makeInt(x - y)
}

func gt(pc *PC) Value {
	x := intValue(pc.Next())
	y := intValue(pc.Next())
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

func fn(pc *PC) Value {
	params := Block(pc.Next())
	code := Block(pc.Next())

	defaults := blockOfRefinements(pc.VM, params)
	stackSize := len(defaults)

	bind(pc.VM, code, func(sym sym, create bool) bound {
		for i, def := range defaults {
			if def == sym {
				return makeStackBinding(i - stackSize)
			}
		}
		return 0
	})

	yfunc := func(pc *PC) Value {
		for i := 0; i < stackSize; i++ {
			pc.VM.push(cell(pc.Next()))
		}
		result := pc.VM.exec(code)
		pc.VM.sp -= uint(stackSize)
		return result
	}

	return pc.VM.addProc(yfunc)
}

func either(pc *PC) Value {
	cond := pc.Next()
	ifTrue := Block(pc.Next())
	ifFalse := Block(pc.Next())

	if boolValue(cond) {
		return pc.VM.exec(ifTrue)
	}

	return pc.VM.exec(ifFalse)
}

func BootVM(vm *VM) {
	vm.dictionary.put(vm, vm.getSymbolID("add"), vm.alloc(cell(vm.addProc(add))))
	vm.dictionary.put(vm, vm.getSymbolID("sub"), vm.alloc(cell(vm.addProc(sub))))

	vm.dictionary.put(vm, vm.getSymbolID("gt"), vm.alloc(cell(vm.addProc(gt))))

	vm.dictionary.put(vm, vm.getSymbolID("either"), vm.alloc(cell(vm.addProc(either))))

	vm.dictionary.put(vm, vm.getSymbolID("fn"), vm.alloc(cell(vm.addProc(fn))))
}
