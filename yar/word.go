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
type bound = imm
type pBindings = ptr
type bindFactory func(sym sym, create bool) bound

func _makeWord(sym sym, bindings pBindings, kind int) Word {
	return makeObj(int(bindings), ptr(sym), kind)
}

func (vm *VM) makeWord(sym string) Word {
	return _makeWord(vm.getSymbolID(sym), 0, WordType)
}

func (vm *VM) makeGetWord(sym string) Word {
	return _makeWord(vm.getSymbolID(sym), 0, GetWordType)
}

func (vm *VM) makeSetWord(sym string) Word {
	return _makeWord(vm.getSymbolID(sym), 0, SetWordType)
}

func (w Word) Sym() sym            { return sym(obj(w).ptr()) }
func (w Word) bindings() pBindings { return pBindings(obj(w).val()) }

func wordBind(vm *VM, ptr ptr, factory bindFactory) {
	w := Word(vm.read(ptr))
	sym := w.Sym()
	bindings := factory(sym, false)
	if bindings != 0 {
		vm.write(ptr, cell(_makeWord(sym, vm.alloc(cell(bindings)), Value(w).Kind())))
	}
}

func setWordBind(vm *VM, ptr ptr, factory bindFactory) {
	w := Word(vm.read(ptr))
	sym := w.Sym()
	bindings := factory(sym, true)
	vm.write(ptr, cell(_makeWord(sym, vm.alloc(cell(bindings)), SetWordType)))
}

func wordExec(vm *VM, val Value) Value {
	// w := Word(val)
	// bindings := Value(vm.read(w.bindings()))
	// if bindings == 0 {
	// 	panic("word not bound")
	// }
	// bindingKind := bindings.Kind()
	// bound := vm.getBound[bindingKind](bindings)
	// return vm.execFunc[bound.Kind()](vm, bound)
	bound := getWordExec(vm, val)
	return vm.execFunc[bound.Kind()](vm, bound)
}

func getWordExec(vm *VM, val Value) Value {
	w := Word(val)
	bindings := Value(vm.read(w.bindings()))
	if bindings == 0 {
		fmt.Printf("%016x, %d, %s\n", val, w.Sym(), vm.InverseSymbols[w.Sym()])
		panic("getword not bound")
	}
	bindingKind := bindings.Kind()
	bound := vm.getBound[bindingKind](bindings)
	// if val.Kind() == GetWordType {
	// 	fmt.Printf(" // %s\n", vm.toString(bound))
	// }
	return bound
}

func setWordExec(vm *VM, val Value) Value {
	w := Word(val)
	bindings := Value(vm.read(w.bindings()))
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
	result.WriteString(strconv.Itoa(int(vm.read(w.bindings()))))
	result.WriteString(")")
	return result.String()
}

// func makeWord(sym sym, bindings ptr, kind int) word {
// 	return word(uint64(bindings)<<32 | uint64(sym)<<8 | uint64(kind))
// }

// func (vm *VM) makeWord(sym string) word {
// 	return makeWord(vm.getSymbolID(sym), 0, WordType)
// }

// func (w pWord) bind(vm *VM, factory bindFactory) {
// 	_word := vm.read(w)
// 	sym := _word.sym()
// 	bindings := factory(sym, false)
// 	if bindings != 0 {
// 		pBindings := _word.bindings()
// 		if pBindings != 0 {
// 			vm.write(pBindings, bindings)
// 		} else {
// 			vm.write(w, makeWord(sym, vm.alloc(bindings), _word.kind()))
// 		}
// 	}
// }

// func wordExec(pc *pc, word word) value {
// 	bindings := pc.vm.read(word.bindings())
// 	if bindings == 0 {
// 		fmt.Printf("word %s\n", pc.vm.inverseSymbols[word.sym()])
// 		panic("word not bound")
// 	}
// 	bindingKind := bindings.kind()
// 	bound := pc.vm.getBound[bindingKind](bindings)
// 	return pc.vm.execFunc[kind(bound)](pc, bound)
// }

// func wordToString(vm *VM, value value) string {
// 	var result strings.Builder
// 	result.WriteString(vm.inverseSymbols[wordSym(value)])
// 	result.WriteString("(")
// 	result.WriteString(strconv.Itoa(int(vm.read(wordBindings(value)))))
// 	result.WriteString(")")
// 	return result.String()
// }
