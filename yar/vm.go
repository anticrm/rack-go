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
type procFunc func(vm *VM) Value

type VM struct {
	pc             pBlockEntry
	mem            []cell
	stack          []Value
	top            uint
	sp             uint
	result         Value
	readOnly       bool
	dictionary     pDict
	proc           []procFunc
	symbols        map[string]sym
	nextSymbol     uint
	InverseSymbols map[sym]string
	Services       map[string]interface{}

	toStringFunc [LastType]func(vm *VM, value Value) string
	bindFunc     [LastType]func(vm *VM, ptr ptr, factory bindFactory)
	execFunc     [LastType]func(vm *VM, value Value) Value
	getBound     [LastBinding]func(bindings imm) Value
	setBound     [LastBinding]func(bindings imm, value Value)
}

func notImplemented(vm *VM, value Value) string {
	return "<not implemented>"
}

func NewVM(memSize int, stackSize int) *VM {
	vm := &VM{
		mem:            make([]cell, memSize),
		top:            0,
		stack:          make([]Value, stackSize),
		sp:             0,
		nextSymbol:     0,
		symbols:        make(map[string]uint),
		InverseSymbols: make(map[sym]string),
		Services:       make(map[string]interface{}),
	}

	for i := 0; i < LastType; i++ {
		vm.toStringFunc[i] = notImplemented
	}
	vm.toStringFunc[BlockType] = blockToString
	vm.toStringFunc[WordType] = wordToString

	vm.bindFunc[WordType] = wordBind
	vm.bindFunc[GetWordType] = wordBind
	vm.bindFunc[SetWordType] = setWordBind
	vm.bindFunc[BlockType] = func(vm *VM, ptr ptr, factory bindFactory) {
		bind(vm, Block(vm.read(ptr)), factory)
	}
	vm.bindFunc[IntegerType] = func(vm *VM, ptr ptr, factory bindFactory) {}

	vm.execFunc[WordType] = wordExec
	vm.execFunc[GetWordType] = getWordExec
	vm.execFunc[SetWordType] = setWordExec
	vm.execFunc[NativeType] = nativeExec
	vm.execFunc[ProcType] = procExec
	vm.execFunc[BlockType] = func(vm *VM, value Value) Value { return value }
	vm.execFunc[IntegerType] = func(vm *VM, value Value) Value { return value }

	vm.dictionary = vm.allocDict()

	vm.initBindings()

	return vm
}

func (vm *VM) initBindings() {
	vm.getBound[MapBinding] = func(binding imm) Value {
		symValPtr := intValue(binding)
		symVal := symval(vm.read(ptr(symValPtr)))
		symValValPtr := symVal.val()
		return Value(vm.read(symValValPtr))
	}

	vm.setBound[MapBinding] = func(binding imm, value Value) {
		symValPtr := ptr(intValue(binding))
		symVal := symval(vm.read(symValPtr))
		p := vm.alloc(cell(value))
		vm.write(symValPtr, cell(makeSymval(symVal.sym(), p)))
	}

	vm.getBound[StackBinding] = func(binding imm) Value {
		offset := intValue(binding)
		return Value(vm.stack[int(vm.sp)+offset])
	}
}

func (vm *VM) Clone() *VM {
	vm.readOnly = true
	return vm
}

func (vm *VM) Fork(stack []Value, sp uint) *VM {
	fork := *vm
	fork.stack = stack
	fork.sp = sp
	fork.initBindings()
	return &fork
}

func (vm *VM) alloc(cell cell) ptr {
	if vm.readOnly {
		panic("alloc in read only mode")
	}
	vm.top++
	vm.mem[vm.top] = cell
	return ptr(vm.top)
}

func (vm *VM) read(ptr ptr) cell { return vm.mem[ptr] }
func (vm *VM) write(ptr ptr, cell cell) {
	if vm.readOnly {
		panic("write in read only mode")
	}
	if ptr == 0 {
		panic("null pointer assignment")
	}
	vm.mem[ptr] = cell
}

func (vm *VM) push(value Value) {
	vm.stack[vm.sp] = value
	vm.sp++
}

func (vm *VM) pop() Value {
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
		vm.InverseSymbols[id] = sym
	}
	return id
}

func (vm *VM) addNative(f procFunc) Value {
	id := len(vm.proc)
	vm.proc = append(vm.proc, f)
	return makeNative(id)
}

func (vm *VM) toString(value Value) string {
	kind := value.Kind()
	return vm.toStringFunc[kind](vm, value)
}

// B I N D I N G S

func bind(vm *VM, block Block, factory bindFactory) {
	for i := block.First(); i != 0; i = i.Next(vm) {
		ptr := i.pval(vm)
		obj := Value(vm.read(ptr))
		kind := obj.Kind()
		vm.bindFunc[kind](vm, ptr, factory)
	}
}

func (vm *VM) bind(block Block) {
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

func (vm *VM) call(block Block) Value {
	pc := vm.pc
	vm.pc = block.First()
	var result Value
	for vm.pc != 0 {
		result = vm.Next()
	}
	vm.pc = pc
	return result
}

func (vm *VM) Exec(first pBlockEntry) Value {
	pc := vm.pc
	vm.pc = first
	var result Value
	for vm.pc != 0 {
		result = vm.Next()
	}
	vm.pc = pc
	return result
}

func (vm *VM) BindAndExec(code pBlock) Value {
	block := Block(vm.read(ptr(code)))
	vm.bind(block)
	return vm.call(block)
}

func (vm *VM) nextNoInfix() Value {
	entry := blockEntry(vm.read(ptr(vm.pc)))
	value := Value(vm.read(entry.pval()))
	vm.pc = entry.next()
	kind := value.Kind()
	result := vm.execFunc[kind](vm, value)
	vm.result = result
	return result
}

func (vm *VM) Next() Value {
	return vm.nextNoInfix()
}

func (vm *VM) AddNative(name string, f procFunc) {
	vm.dictionary.put(vm, vm.getSymbolID(name), vm.alloc(cell(vm.addNative(f))))
}

func (vm *VM) Push(value Value) {
	vm.push(value)
}

func (vm *VM) Show() {
	fmt.Printf("Stack pointer: %d\n", vm.sp)
}
