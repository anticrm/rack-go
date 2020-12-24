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
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
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
	procNames      []string
	symbols        map[string]sym
	nextSymbol     uint
	InverseSymbols map[sym]string
	Services       map[string]interface{}

	toStringFunc [LastType]func(vm *VM, value Value) string
	bindFunc     []func(vm *VM, ptr ptr, factory bindFactory)
	execFunc     []func(vm *VM, value Value) Value
	getBound     [LastBinding]func(bindings imm) Value
	setBound     [LastBinding]func(bindings imm, value Value)
}

func notImplemented(vm *VM, value Value) string {
	return fmt.Sprintf("<not implemented:%d>", value.Kind())
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

	vm.dictionary = vm.allocDict()
	vm.initBindings()

	return vm
}

func (vm *VM) initBindings() {

	vm.execFunc = execFunc
	vm.bindFunc = bindFunc

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

type Library interface {
	getFunction(name string) procFunc
}

func (vm *VM) addNativeFunc(name string, f procFunc) {
	vm.procNames = append(vm.procNames, name)
	vm.dictionary.put(vm, vm.getSymbolID(name), vm.alloc(cell(vm.addNative(f))))
}

func (vm *VM) AddNative(name string, lib Library) {
	vm.procNames = append(vm.procNames, name)
	vm.dictionary.put(vm, vm.getSymbolID(name), vm.alloc(cell(vm.addNative(lib.getFunction(name)))))
}

func (vm *VM) Push(value Value) {
	vm.push(value)
}

func (vm *VM) Show() {
	fmt.Printf("Stack pointer: %d\n", vm.sp)
}

type SerialVM struct {
	Top        uint
	Dictionary uint
	Mem        []cell
	Symbols    map[string]sym
	ProcNames  []string
}

func (vm *VM) Save() []byte {
	var result bytes.Buffer

	svm := &SerialVM{Top: vm.top, Dictionary: uint(vm.dictionary), Mem: vm.mem, Symbols: vm.symbols, ProcNames: vm.procNames}

	enc := gob.NewEncoder(&result)
	err := enc.Encode(svm)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	return result.Bytes()
}

func LoadVM(data []byte, stackSize int, lib Library) *VM {
	reader := bytes.NewReader(data)
	var svm SerialVM

	dec := gob.NewDecoder(reader)
	err := dec.Decode(&svm)
	if err != nil {
		log.Fatal("decode error 1:", err)
	}

	vm := &VM{top: svm.Top, mem: svm.Mem, dictionary: pDict(svm.Dictionary), symbols: svm.Symbols, procNames: svm.ProcNames}

	vm.InverseSymbols = make(map[sym]string)
	for k, v := range vm.symbols {
		vm.InverseSymbols[v] = k
	}
	vm.nextSymbol = uint(len(vm.symbols))

	for _, n := range vm.procNames {
		vm.proc = append(vm.proc, lib.getFunction(n))
	}

	vm.stack = make([]Value, stackSize)
	vm.initBindings()

	return vm
}

// pc             pBlockEntry
// mem            []cell
// stack          []Value
// top            uint
// sp             uint
// result         Value
// readOnly       bool
// dictionary     pDict
// proc           []procFunc
// procNames      []string
// symbols        map[string]sym
// nextSymbol     uint
// InverseSymbols map[sym]string
// Services       map[string]interface{}
