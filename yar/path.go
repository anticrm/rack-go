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
	"strings"
)

type path obj
type pathEntry obj
type pPathEntry ptr

func (p path) bindings() pBindings { return pBindings(obj(p).val()) }
func (p path) first() pPathEntry   { return pPathEntry(obj(p).ptr()) }

func _makePath(bindings pBindings, first pPathEntry, kind int) path {
	return path(makeObj(int(bindings), ptr(first), kind))
}

func pathBind(vm *VM, ptr ptr, factory bindFactory) {
	p := path(vm.read(ptr))
	fmt.Printf("path %016x\n", p)
	symPtr := p.first()
	sym := symPtr.sym(vm)
	bindings := factory(sym, false)
	fmt.Printf("binding path %x %d\n", bindings, sym)
	if bindings != 0 {
		pb := pBindings(vm.alloc(cell(bindings)))
		path := _makePath(pb, symPtr, Value(p).Kind())
		vm.write(ptr, cell(path))
	}
}

func getPathExec(vm *VM, val Value) Value {
	p := path(val)
	bindings := Value(vm.read(p.bindings()))
	if bindings == 0 {
		// fmt.Printf("%016x, %d, %s\n", val, p.sym(), vm.InverseSymbols[w.Sym()])
		panic("getpath not bound")
	}
	bindingKind := bindings.Kind()
	bound := dict(vm.getBound[bindingKind](bindings))
	fmt.Printf("bound %x\n", bound)

	symPtr := p.first()
	i := symPtr.Next(vm)

	for i != 0 {
		sym := i.sym(vm)
		// dict := pDict(bound.val())
		// fmt.Printf("@ %x\n", dict)
		// fmt.Printf("dict %s\n", vm.toString(Value(vm.read(ptr(dict)))))
		sv := symval(vm.read(ptr(bound.find(vm, sym))))
		bound = dict(vm.read(sv.val()))
		i = i.Next(vm)
	}

	// if val.Kind() == GetWordType {
	// 	fmt.Printf(" // %s\n", vm.toString(bound))
	// }
	return Value(bound)
}

func (b pathEntry) next() pPathEntry { return pPathEntry(obj(b).ptr()) }
func (b pathEntry) sym() sym         { return sym(obj(b).val()) }

func (b pPathEntry) Next(vm *VM) pPathEntry { return pathEntry(vm.read(ptr(b))).next() }
func (b pPathEntry) sym(vm *VM) sym         { return pathEntry(vm.read(ptr(b))).sym() }

type pathBuilder struct {
	first pPathEntry
	last  pPathEntry
}

func (b *pathBuilder) add(vm *VM, sym sym) {
	next := pPathEntry(vm.alloc(cell(makeObj(int(sym), 0, PathType))))
	if b.last == 0 {
		b.first = next
		b.last = next
	} else {
		cur := pathEntry(vm.read(ptr(b.last)))
		vm.write(ptr(b.last), cell(makeObj(int(cur.sym()), ptr(next), PathType)))
		b.last = next
	}
}

func (b pathBuilder) get(kind int) path {
	return path(makeObj(0, ptr(b.first), kind))
}

func pathToString(vm *VM, p pathEntry) string {
	var result strings.Builder

	result.WriteString(vm.InverseSymbols[p.sym()])
	i := p.next()

	for i != 0 {
		result.WriteByte('/')
		result.WriteString(vm.InverseSymbols[i.sym(vm)])
		i = i.Next(vm)
	}

	return result.String()
}
