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

// W O R D
//-------------------------
//   BINDING | SYM | KIND |
//-------------------------

type Word = obj
type pWord = pItem

// type Binding imm
// type pBindings = ptr
type bindFactory func(sym sym, create bool) Binding

func (v Value) Word() Word  { return Word(v) }
func (w Word) Value() Value { return Value(w) }

func _makeWord(sym sym, bindings pBinding, kind int) Word {
	return makeObj(int(bindings), ptr(sym), kind)
}

func (vm *VM) AllocWord(sym sym) Word {
	bindings := pBinding(vm.alloc(0))
	return _makeWord(sym, bindings, WordType)
}

func (vm *VM) AllocGetWord(sym sym) Word {
	bindings := pBinding(vm.alloc(0))
	return _makeWord(sym, bindings, GetWordType)
}

func (vm *VM) AllocSetWord(sym sym) Word {
	bindings := pBinding(vm.alloc(0))
	return _makeWord(sym, bindings, SetWordType)
}

func (vm *VM) AllocQuoteWord(sym sym) Word {
	bindings := pBinding(vm.alloc(0))
	return _makeWord(sym, bindings, QuoteType)
}

func (w Word) Sym() sym           { return sym(obj(w).ptr()) }
func (w Word) bindings() pBinding { return pBinding(obj(w).val()) }

func wordBind(vm *VM, value Value, factory bindFactory) {
	w := value.Word()
	sym := w.Sym()
	bindings := factory(sym, false)
	if bindings != 0 {
		vm.write(ptr(w.bindings()), cell(bindings))
	}
}

func setWordBind(vm *VM, value Value, factory bindFactory) {
	w := value.Word()
	sym := w.Sym()
	bindings := factory(sym, true)
	vm.write(ptr(w.bindings()), cell(bindings))
}

func wordExec(vm *VM, val Value) Value {
	bound := getWordExec(vm, val)
	return vm.execFunc[bound.Kind()](vm, bound)
}

func getWordExec(vm *VM, val Value) Value {
	w := Word(val)
	bindings := Binding(vm.read(ptr(w.bindings())))
	if bindings == 0 {
		fmt.Printf("%016x, %d, %s\n", val, w.Sym(), vm.InverseSymbols[w.Sym()])
		return MakeError(1).Value()
	}
	bindingKind := bindings.Kind()
	bound := vm.getBound[bindingKind](bindings)
	return bound
}

func setWordExec(vm *VM, val Value) Value {
	w := Word(val)
	bindings := Binding(vm.read(ptr(w.bindings())))
	if bindings == 0 {
		panic("setword not bound")
	}
	result := vm.Next()
	bindingKind := bindings.Kind()
	vm.setBound[bindingKind](bindings, result)
	return result
}

func wordToString(vm *VM, value Value) string {
	var result strings.Builder
	w := Word(value)
	result.WriteString(vm.InverseSymbols[w.Sym()])
	result.WriteString("(")
	result.WriteString(strconv.Itoa(int(vm.read(ptr(w.bindings())))))
	result.WriteString(")")
	return result.String()
}
