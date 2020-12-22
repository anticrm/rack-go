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

package z

type dict = obj
type pDict ptr
type pDictEntry ptr
type symval item
type pSymval ptr

func makeDict() dict {
	return dict(makeObj(0, 0, MapType))
}

func (vm *VM) allocDict() pDict {
	return pDict(vm.alloc(cell(makeDict())))
}

func (d dict) first() pDictEntry { return pDictEntry(obj(d).ptr()) }

func (p pDictEntry) next(vm *VM) pDictEntry { return pDictEntry(pItem(p).ptr(vm)) }
func (p pDictEntry) symval(vm *VM) pSymval  { return pSymval(pItem(p).val(vm)) }

func (p pSymval) sym(vm *VM) sym           { return sym(pItem(p).ptr(vm)) }
func (p pSymval) setPtr(vm *VM, value ptr) { pPtrval(p).setPtr(vm, value) }

// func (p pSymval) setValue(vm *VM, value value) sym { return pPtrval(p).setValue(vm, value) }

func (pd pDict) put(vm *VM, sym sym, value ptr) {
	d := dict(vm.read(ptr(pd)))
	last := pDictEntry(0)
	for i := d.first(); i != 0; i = i.next(vm) {
		last = i
		sv := i.symval(vm)
		if sv.sym(vm) == sym {
			sv.setPtr(vm, value)
			return
		}
	}

	symval := vm.alloc(cell(makePtrVal(value, ptr(sym))))
	pair := vm.alloc(cell(makeItem(symval, 0)))

	if last != 0 {
		pItem(last).setPtr(vm, ptr(pair))
	} else {
		vm.write(ptr(pd), cell(makeObj(0, ptr(pair), MapType)))
	}
}

func (pd pDict) find(vm *VM, sym sym) pPtrVal {
	d := dict(vm.read(ptr(pd)))
	for i := d.first(); i != 0; i = i.next(vm) {
		sv := i.symval(vm)
		if sv.sym(vm) == sym {
			return sv.
		}
	}
}