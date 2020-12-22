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

package x

func add(pc *pc) value {
	x := intValue(pc.next())
	y := intValue(pc.next())
	return makeInt(x + y)
}

func sub(pc *pc) value {
	x := intValue(pc.next())
	y := intValue(pc.next())
	return makeInt(x - y)
}

func gt(pc *pc) value {
	x := intValue(pc.next())
	y := intValue(pc.next())
	return makeBool(x > y)
}

func blockOfRefinements(vm *VM, code block) []sym {
	var result []sym

	for i := code.first(); i != 0; i = i.next(vm) {
		w := word(i.value(vm))
		result = append(result, w.sym())
	}

	return result
}

type pcAlias = pc

func fn(pc *pc) value {
	params := block(pc.next())
	code := block(pc.next())

	defaults := blockOfRefinements(pc.vm, params)
	stackSize := len(defaults)

	bind(pc.vm, code, func(sym sym, create bool) bound {
		for i, def := range defaults {
			if def == sym {
				return makeStackBinding(i - stackSize)
			}
		}
		return 0
	})

	yfunc := func(pc *pcAlias) value {
		for i := 0; i < stackSize; i++ {
			pc.vm.push(cell(pc.next()))
		}
		result := pc.vm.exec(code)
		pc.vm.sp -= uint(stackSize)
		return result
	}

	return pc.vm.addProc(yfunc)
}

func either(pc *pc) value {
	cond := pc.next()
	ifTrue := block(pc.next())
	ifFalse := block(pc.next())

	if boolValue(cond) {
		return pc.vm.exec(ifTrue)
	}

	return pc.vm.exec(ifFalse)
}

func BootVM(vm *VM) {
	vm.dictionary.put(vm, vm.getSymbolID("add"), vm.alloc(cell(vm.addProc(add))))
	vm.dictionary.put(vm, vm.getSymbolID("sub"), vm.alloc(cell(vm.addProc(sub))))

	vm.dictionary.put(vm, vm.getSymbolID("gt"), vm.alloc(cell(vm.addProc(gt))))

	vm.dictionary.put(vm, vm.getSymbolID("either"), vm.alloc(cell(vm.addProc(either))))

	vm.dictionary.put(vm, vm.getSymbolID("fn"), vm.alloc(cell(vm.addProc(fn))))
}
