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

import (
	"fmt"
	"strconv"
	"strings"
)

// W O R D
//-------------------------
//   BINDING | SYM | KIND |
//-------------------------

type word = obj
type pWord = pItem
type bound = imm
type pBindings = ptr
type bindFactory func(sym sym, create bool) bound

func _makeWord(sym sym, bindings pBindings, kind int) word {
	return makeObj(int(bindings), ptr(sym), kind)
}

func (vm *VM) makeWord(sym string) word {
	return _makeWord(vm.getSymbolID(sym), 0, WordType)
}

func (vm *VM) makeSetWord(sym string) word {
	return _makeWord(vm.getSymbolID(sym), 0, SetWordType)
}

func (w word) sym() sym            { return sym(obj(w).ptr()) }
func (w word) bindings() pBindings { return pBindings(obj(w).val()) }

func wordBind(vm *VM, ptr ptr, factory bindFactory) {
	w := word(vm.read(ptr))
	sym := w.sym()
	bindings := factory(sym, false)
	if bindings != 0 {
		vm.write(ptr, cell(_makeWord(sym, vm.alloc(cell(bindings)), WordType)))
	}
}

func setWordBind(vm *VM, ptr ptr, factory bindFactory) {
	w := word(vm.read(ptr))
	sym := w.sym()
	// fmt.Printf("create bindings %s\n", vm.inverseSymbols[sym])
	bindings := factory(sym, true)
	vm.write(ptr, cell(_makeWord(sym, vm.alloc(cell(bindings)), SetWordType)))
}

func wordExec(pc *pc, val value) value {
	w := val
	bindings := value(pc.vm.read(w.bindings()))
	if bindings == 0 {
		fmt.Printf("word %s\n", pc.vm.inverseSymbols[w.sym()])
		panic("word not bound")
	}
	bindingKind := bindings.kind()
	bound := pc.vm.getBound[bindingKind](bindings)
	return pc.vm.execFunc[bound.kind()](pc, bound)
}

func setWordExec(pc *pc, val value) value {
	w := word(val)
	bindings := value(pc.vm.read(w.bindings()))
	if bindings == 0 {
		panic("word not bound")
	}
	result := pc.next()
	bindingKind := bindings.kind()
	pc.vm.setBound[bindingKind](bindings, result)
	return result
}

func wordToString(vm *VM, value value) string {
	var result strings.Builder
	w := word(value)
	result.WriteString(vm.inverseSymbols[w.sym()])
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
