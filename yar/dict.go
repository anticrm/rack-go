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

import (
	"fmt"
	"strconv"
	"strings"
)

type dict Value
type dictFirst cell
type pDictFirst ptr
type pDictEntry ptr
type symval item
type pSymval ptr

func makeDict(first pDictFirst) dict {
	return dict(makeValue(int(first), MapType))
}

func (vm *VM) AllocDict() dict {
	return makeDict(pDictFirst(vm.alloc(0)))
}

func (d dict) Put(vm *VM, sym sym, value Value) pSymval { return d.dictFirst().put(vm, sym, value) }
func (d dict) Find(vm *VM, sym sym) pSymval             { return d.dictFirst().find(vm, sym) }
func (d dict) Value() Value                             { return Value(d) }
func (d dict) dictFirst() pDictFirst                    { return pDictFirst(d.Value().Val()) }
func (v Value) Dict() dict                              { return dict(v) }

func (d dictFirst) first() pDictEntry { return pDictEntry(obj(d).ptr()) }

func (p pDictEntry) next(vm *VM) pDictEntry { return pDictEntry(pItem(p).ptr(vm)) }
func (p pDictEntry) symval(vm *VM) pSymval  { return pSymval(pItem(p).val(vm)) }

func (p pSymval) sym(vm *VM) sym           { return sym(pItem(p).ptr(vm)) }
func (p pSymval) val(vm *VM) int           { return pItem(p).val(vm) }
func (p pSymval) setPtr(vm *VM, value ptr) { pItem(p).setPtr(vm, value) }

func makeSymval(sym sym, value ptr) symval { return symval(makeItem(int(value), ptr(sym))) }

func (s symval) sym() sym { return sym(item(s).ptr()) }
func (s symval) val() ptr { return ptr(item(s).val()) }

// func (p pSymval) setValue(vm *VM, value value) sym { return pPtrval(p).setValue(vm, value) }

func (pd pDictFirst) put(vm *VM, sym sym, value Value) pSymval {
	p := vm.alloc(cell(value))
	d := dictFirst(vm.read(ptr(pd)))
	last := pDictEntry(0)
	for i := d.first(); i != 0; i = i.next(vm) {
		last = i
		sv := i.symval(vm)
		if sv.sym(vm) == sym {
			sv.setPtr(vm, p)
			return sv
		}
	}

	symval := vm.alloc(cell(makeSymval(sym, p)))
	pair := vm.alloc(cell(makeItem(int(symval), 0)))

	if last != 0 {
		pItem(last).setPtr(vm, ptr(pair))
	} else {
		vm.write(ptr(pd), cell(makeObj(0, ptr(pair), MapType)))
	}

	return pSymval(symval)
}

func (pd pDictFirst) find(vm *VM, sym sym) pSymval {
	d := dictFirst(vm.read(ptr(pd)))
	for i := d.first(); i != 0; i = i.next(vm) {
		sv := i.symval(vm)
		if sv.sym(vm) == sym {
			return sv
		}
	}
	return 0
}

func dictToString(vm *VM, b Value) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf(" #%016x ", b))
	result.WriteByte('[')

	dict := b.Dict()
	first := dict.dictFirst()
	d := dictFirst(vm.read(ptr(first)))
	for i := d.first(); i != 0; i = i.next(vm) {
		sv := i.symval(vm)
		result.WriteString(symvalToString(vm, sv))
		// result.WriteString(vm.InverseSymbols[sv.sym(vm)])
		// result.WriteString(": ")
		// val := vm.read(ptr(sv.val(vm)))
		// result.WriteString(vm.toString(Value(val)))
	}

	result.WriteByte(']')
	return result.String()
}

func symvalToString(vm *VM, sv pSymval) string {
	val := vm.read(ptr(sv.val(vm)))
	sym := sv.sym(vm)
	return vm.InverseSymbols[sym] + "(" + strconv.Itoa(int(sym)) + ")" + ": " + vm.ToString(Value(val))
}
