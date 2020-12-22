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
)

const (
	MapBinding   = iota
	StackBinding = iota
	LastBinding  = iota
)

func makeMapBinding(value ptr) imm {
	return makeImm(int(value), MapBinding)
}

func makeStackBinding(value int) imm {
	return makeImm(value, StackBinding)
}

type sym = uint
type procFunc func(pc *pc) value

type VM struct {
	mem            []cell
	stack          []cell
	top            uint
	sp             uint
	result         value
	dictionary     pDict
	proc           []procFunc
	symbols        map[string]sym
	nextSymbol     uint
	inverseSymbols map[sym]string

	toStringFunc [LastType]func(vm *VM, value value) string
	bindFunc     [LastType]func(vm *VM, ptr ptr, factory bindFactory)
	execFunc     [LastType]func(pc *pc, value value) value
	getBound     [LastBinding]func(bindings imm) value
	setBound     [LastBinding]func(bindings imm, value value)
}

func notImplemented(vm *VM, value value) string {
	return "<not implemented>"
}

func NewVM(memSize int, stackSize int) *VM {
	vm := &VM{
		mem:            make([]cell, memSize),
		top:            0,
		stack:          make([]cell, stackSize),
		sp:             0,
		nextSymbol:     0,
		symbols:        make(map[string]uint),
		inverseSymbols: make(map[uint]string),
	}

	for i := 0; i < LastType; i++ {
		vm.toStringFunc[i] = notImplemented
	}
	vm.toStringFunc[BlockType] = blockToString
	vm.toStringFunc[WordType] = wordToString

	vm.bindFunc[WordType] = wordBind
	vm.bindFunc[SetWordType] = setWordBind
	vm.bindFunc[BlockType] = func(vm *VM, ptr ptr, factory bindFactory) {
		bind(vm, block(vm.read(ptr)), factory)
	}
	vm.bindFunc[IntegerType] = func(vm *VM, ptr ptr, factory bindFactory) {}

	vm.execFunc[WordType] = wordExec
	vm.execFunc[SetWordType] = setWordExec
	vm.execFunc[ProcType] = procExec
	vm.execFunc[BlockType] = func(pc *pc, value value) value { return value }
	vm.execFunc[IntegerType] = func(pc *pc, value value) value { return value }

	vm.getBound[MapBinding] = func(binding imm) value {
		symValPtr := intValue(binding)
		symVal := symval(vm.read(ptr(symValPtr)))
		symValValPtr := symVal.val()
		return value(vm.read(symValValPtr))
	}

	vm.setBound[MapBinding] = func(binding imm, value value) {
		symValPtr := intValue(binding)
		symVal := symval(vm.read(ptr(symValPtr)))
		symValValPtr := symVal.val()
		vm.write(symValValPtr, cell(value))
	}

	vm.getBound[StackBinding] = func(binding imm) value {
		offset := intValue(binding)
		return value(vm.stack[int(vm.sp)+offset])
	}

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

func (vm *VM) push(value cell) {
	vm.stack[vm.sp] = value
	vm.sp++
}

func (vm *VM) pop() cell {
	vm.sp--
	return vm.stack[vm.sp]
}

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

func (vm *VM) toString(value value) string {
	kind := value.kind()
	return vm.toStringFunc[kind](vm, value)
}

// B I N D I N G S

func bind(vm *VM, block block, factory bindFactory) {
	for i := block.first(); i != 0; i = i.next(vm) {
		ptr := i.pval(vm)
		obj := value(vm.read(ptr))
		kind := obj.kind()
		vm.bindFunc[kind](vm, ptr, factory)
	}
}

func (vm *VM) bind(block block) {
	bind(vm, block, func(sym sym, create bool) bound {
		symValPtr := vm.dictionary.find(vm, sym)
		if symValPtr == 0 {
			if create {
				// fmt.Printf("putting symbol %d - %s\n", sym, vm.inverseSymbols[sym])
				vm.dictionary.put(vm, sym, 0)
				symValPtr = vm.dictionary.find(vm, sym) // TODO: fix this garbage
				// fmt.Printf("found %16x\n", symValPtr)
			} else {
				// fmt.Printf("binding not found %d - %s\n", sym, vm.inverseSymbols[sym])
				// panic("can't find binding")
				return 0
			}
		}
		return makeMapBinding(ptr(symValPtr))
	})
}

func (vm *VM) exec(block block) value {
	return newPC(vm, block).exec()
}

// P C

type pc struct {
	pc pBlockEntry
	vm *VM
}

func newPC(vm *VM, block block) *pc {
	first := block.first()
	return &pc{vm: vm, pc: first}
}

func (pc *pc) nextNoInfix() value {
	vm := pc.vm
	entry := blockEntry(vm.read(ptr(pc.pc)))
	value := value(vm.read(entry.pval()))
	pc.pc = entry.next()
	kind := value.kind()
	result := vm.execFunc[kind](pc, value)
	vm.result = result
	return result
}

func (pc *pc) next() value {
	return pc.nextNoInfix()
}

func (pc *pc) exec() value {
	var result value
	for pc.pc != 0 {
		result = pc.next()
	}
	return result
}
