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

package y

func blockOfRefinements(vm *VM, code value) []sym {
	// result := make(map[string]Code)
	// current := "default"
	// result[current] = make(Code, 0)
	var result []sym

	first := blockFirst(code)
	entryPtr := first
	for entryPtr != 0 {
		entry := vm.read(entryPtr)
		value := blockEntryValue(vm, entry) // word
		sym := wordSym(value)
		result = append(result, sym)
		entryPtr = blockEntryNext(entry)
	}

	// for _, item := range code {
	// 	result[current] = append(result[current], item)
	// }
	// return result
	return result
}

// func native(pc *pc) value {
// 	params := pc.next()
// 	impl := pc.next()

// 	ref := blockOfRefinements(pc.vm, params)

// 	defaults := ref //len(ref["default"])

// 	yfunc := func(pc *pcAlias) value {
// 		for i := 0; i < defaults; i++ {
// 			pc.vm.push(pc.next())
// 		}
// 		// impl -- ProcType
// 		return getProcFunc(pc.vm, impl)(pc)
// 	}

// 	return pc.vm.addProc(yfunc)
// }

func fn(pc *pc) value {
	params := pc.next()
	code := pc.next()

	defaults := blockOfRefinements(pc.vm, params)
	stackSize := len(defaults)

	// fmt.Printf("rebinding fn...\n")
	bind(pc.vm, code, func(sym sym, create bool) bound {
		for i, def := range defaults {
			if def == sym {
				return newStackBindins(i - stackSize)
			}
		}
		return 0
	})

	yfunc := func(pc *pcAlias) value {
		for i := 0; i < stackSize; i++ {
			pc.vm.push(pc.next())
		}
		result := pc.vm.Exec(code)
		pc.vm.sp -= stackSize
		return result
	}

	return pc.vm.addProc(yfunc)
}

func wrap(params int, f procFunc) procFunc {
	return func(pc *pc) value {
		for i := 0; i < params; i++ {
			pc.vm.push(pc.next())
		}
		return f(pc)
	}
}

func add(pc *pc) value {
	x := intValue(pc.next())
	y := intValue(pc.next())
	return newInteger(x + y)
}

func sub(pc *pc) value {
	x := intValue(pc.next())
	y := intValue(pc.next())
	return newInteger(x - y)
}

func gt(pc *pc) value {
	x := intValue(pc.next())
	y := intValue(pc.next())
	return newBoolean(x > y)
}

func either(pc *pc) value {
	cond := pc.next()
	ifTrue := pc.next()
	ifFalse := pc.next()

	if boolValue(cond) {
		return pc.vm.Exec(ifTrue)
	}

	return pc.vm.Exec(ifFalse)
}

func BootVM(vm *VM) {
	mapPut(vm, vm.dictionary, vm.getSymbolID("add"), vm.alloc(vm.addProc(add)))
	mapPut(vm, vm.dictionary, vm.getSymbolID("sub"), vm.alloc(vm.addProc(sub)))
	mapPut(vm, vm.dictionary, vm.getSymbolID("gt"), vm.alloc(vm.addProc(gt)))
	mapPut(vm, vm.dictionary, vm.getSymbolID("fn"), vm.alloc(vm.addProc(fn)))
	mapPut(vm, vm.dictionary, vm.getSymbolID("either"), vm.alloc(vm.addProc(either)))
}
