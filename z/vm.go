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

import (
	"fmt"
)

const (
	MapBinding   = iota
	StackBinding = iota
	LastBinding  = iota
)

type sym = uint
type procFunc func(pc *pc) value

type VM struct {
	mem            []cell
	top            uint
	result         value
	dictionary     pDict
	proc           []procFunc
	symbols        map[string]sym
	nextSymbol     uint
	inverseSymbols map[sym]string

	toStringFunc [LastType]func(vm *VM, cell cell) string
	bindFunc     [LastType]func(vm *VM, ptr ptr, factory bindFactory)
	execFunc     [LastType]func(pc *pc, value value) value
	getBound     [LastBinding]func(bindings imm) value
	setBound     [LastBinding]func(bindings imm, value value)
}

func NewVM(memSize int) *VM {
	vm := &VM{
		mem:            make([]cell, memSize),
		top:            0,
		nextSymbol:     0,
		symbols:        make(map[string]uint),
		inverseSymbols: make(map[uint]string),
	}

	vm.getBound[MapBinding] = getMapBinding

	vm.dictionary = vm.allocDict()

	return vm
}

func (vm *VM) alloc(cell cell) ptr {
	vm.top++
	vm.mem[vm.top] = cell
	return ptr(vm.top)
}

func (vm *VM) read(ptr ptr) cell        { return vm.mem[ptr] }
func (vm *VM) write(ptr ptr, cell cell) { vm.mem[ptr] = cell }

func (vm *VM) dump() {
	for i := 0; i <= int(vm.top); i++ {
		fmt.Printf("%016x\n", vm.mem[i])
	}
}

func (vm *VM) getSymbolID(sym string) uint {
	id, ok := vm.symbols[sym]
	if !ok {
		vm.nextSymbol++
		id = vm.nextSymbol
		vm.symbols[sym] = id
		vm.inverseSymbols[id] = sym
	}
	return id
}

func (vm *VM) addProc(f procFunc) value {
	id := len(vm.proc)
	vm.proc = append(vm.proc, f)
	return makeProc(id)
}

// B I N D I N G S

func bind(vm *VM, block block, factory bindFactory) {
	for i := block.first(); i != 0; i = i.next(vm) {
		ptr, ok := i.ptr(vm)
		if ok {
			obj := value(vm.read(ptr))
			kind := obj.kind()
			vm.bindFunc[kind](vm, ptr, factory)
		}
	}
}

func (vm *VM) bind(block block) {
	bind(vm, block, func(sym sym, create bool) bound {
		symValPtr := vm.dictionary.find(vm, sym)
		if symValPtr == 0 {
			if create {
				// fmt.Printf("putting symbol %d - %s\n", sym, vm.inverseSymbols[sym])
				mapPut(vm, vm.dictionary, sym, vm.alloc(0))
				symValPtr = mapFind(vm, vm.dictionary, sym) // TODO: fix this garbage
				// fmt.Printf("found %16x\n", symValPtr)
			} else {
				// fmt.Printf("binding not found %d - %s\n", sym, vm.inverseSymbols[sym])
				// panic("can't find binding")
				return 0
			}
		}
		return newMapBinding(symValPtr)
	})
}

func getMapBinding(binding imm) value {
	// symValPtr := mapBindingPtr(binding)
	// symVal := vm.read(symValPtr)
	// symValValPtr := symValVal(symVal)
	// value := vm.read(symValValPtr)
	// return value
	return 0
}

// P C

type pc struct {
	pc pBlockEntry
	vm *VM
}

func (pc *pc) nextNoInfix() value {
	entry := blockEntry(pc.vm.read(ptr(pc.pc)))
	value := entry.value(pc.vm)
	pc.pc = entry.next()
	kind := value.kind()
	result := pc.vm.execFunc[kind](pc, value)
	pc.vm.result = result
	return result
}

func (pc *pc) next() value {
	return pc.nextNoInfix()
}
