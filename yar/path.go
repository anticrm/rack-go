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
	"strings"
)

type path obj

// type pathEntry obj
// type pPathEntry ptr

func (p path) bindings() pBinding    { return pBinding(obj(p).val()) }
func (p path) firstLast() pFirstLast { return pFirstLast(obj(p).ptr()) }
func (p path) Add(vm *VM, sym sym)   { p.firstLast().addPtr(vm, ptr(sym)) }
func (p path) Value() Value          { return Value(p) }

func (v Value) Path() path { return path(v) }

func _makePath(bindings pBinding, firstLast pFirstLast, kind int) path {
	return path(makeObj(int(bindings), ptr(firstLast), kind))
}

func (vm *VM) AllocPath() path {
	bindings := pBinding(vm.alloc(0))
	firstlast := pFirstLast(vm.alloc(0))
	return _makePath(bindings, firstlast, PathType)
}

func (vm *VM) AllocGetPath() path {
	bindings := pBinding(vm.alloc(0))
	firstlast := pFirstLast(vm.alloc(0))
	return _makePath(bindings, firstlast, GetPathType)
}

func pathBind(vm *VM, value Value, factory bindFactory) {
	p := value.Path()
	fl := firstLast(vm.read(ptr(p.firstLast())))
	first := fl.first()
	sym := sym(first.pval(vm))
	bindings := factory(sym, false)
	if bindings != 0 {
		vm.write(ptr(p.bindings()), cell(bindings))
	}
}

func getPathExec(vm *VM, val Value) Value {
	p := val.Path()
	bindings := Binding(vm.read(ptr(p.bindings())))
	if bindings == 0 {
		// fmt.Printf("%016x, %d, %s\n", val, p.sym(), vm.InverseSymbols[w.Sym()])
		panic("getpath not bound")
	}
	bindingKind := bindings.Kind()
	bound := vm.getBound[bindingKind](bindings).Dict()
	// fmt.Printf("bound %s\n", vm.toString(bound.Value()))

	fl := firstLast(vm.read(ptr(p.firstLast())))
	first := fl.first()
	i := first.Next(vm)

	for i != 0 {
		sym := sym(i.pval(vm))
		// fmt.Printf("looking for sym: %s\n", vm.InverseSymbols[sym])
		psv := bound.Find(vm, sym)
		sv := symval(vm.read(ptr(psv)))
		// fmt.Printf("BOUND: %016x\n", psv)
		bound = dict(vm.read(sv.val()))
		i = i.Next(vm)
	}

	// if val.Kind() == GetWordType {
	// 	fmt.Printf(" // %s\n", vm.toString(bound))
	// }
	return Value(bound)
}

func pathExec(vm *VM, val Value) Value {
	bound := getPathExec(vm, val)
	return vm.execFunc[bound.Kind()](vm, bound)
	// return getPathExec(vm, val)
}

// func (b pathEntry) next() pPathEntry { return pPathEntry(obj(b).ptr()) }
// func (b pathEntry) sym() sym         { return sym(obj(b).val()) }

// func (b pPathEntry) Next(vm *VM) pPathEntry { return pathEntry(vm.read(ptr(b))).next() }
// func (b pPathEntry) sym(vm *VM) sym         { return pathEntry(vm.read(ptr(b))).sym() }

// type pathBuilder struct {
// 	first pPathEntry
// 	last  pPathEntry
// }

// func (b *pathBuilder) add(vm *VM, sym sym) {
// 	next := pPathEntry(vm.alloc(cell(makeObj(int(sym), 0, PathType))))
// 	if b.last == 0 {
// 		b.first = next
// 		b.last = next
// 	} else {
// 		cur := pathEntry(vm.read(ptr(b.last)))
// 		vm.write(ptr(b.last), cell(makeObj(int(cur.sym()), ptr(next), PathType)))
// 		b.last = next
// 	}
// }

// func (b pathBuilder) get(kind int) path {
// 	return path(makeObj(0, ptr(b.first), kind))
// }

func pathToString(vm *VM, p path) string {
	var result strings.Builder

	fl := firstLast(vm.read(ptr(p.firstLast())))
	i := fl.first()

	for i != 0 {
		result.WriteByte('/')
		result.WriteString(vm.InverseSymbols[sym(i.pval(vm))])
		i = i.Next(vm)
	}

	return result.String()
}
