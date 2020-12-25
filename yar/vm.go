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
	"strings"
)

const (
	MapBinding   = iota
	StackBinding = iota
	LastBinding  = iota
)

func makeMapBinding(value ptr) Binding {
	return Binding(makeImm(int(value), MapBinding))
}

func MakeStackBinding(value int) Binding {
	return Binding(makeImm(value, StackBinding))
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
	Dictionary     dict
	proc           []procFunc
	procNames      []string
	symbols        map[string]sym
	nextSymbol     uint
	InverseSymbols map[sym]string
	strings        map[uint]string
	nextString     uint
	Library        Library
	Services       map[string]interface{}

	toStringFunc [LastType]func(vm *VM, value Value) string
	bindFunc     []func(vm *VM, value Value, factory bindFactory)
	execFunc     []func(vm *VM, value Value) Value
	getBound     [LastBinding]func(bindings Binding) Value
	setBound     [LastBinding]func(bindings Binding, value Value)
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
		nextString:     0,
		strings:        make(map[uint]string),
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
	vm.toStringFunc[MapType] = dictToString
	vm.toStringFunc[IntegerType] = intToString

	vm.Dictionary = vm.AllocDict()
	vm.initBindings()

	loadNative := vm.addNative(loadNative)
	sym := sym(vm.GetSymbolID("load-native"))
	vm.Dictionary.Put(vm, sym, loadNative)
	vm.procNames = append(vm.procNames, "boot/load-native")

	return vm
}

func (vm *VM) initBindings() {

	vm.execFunc = execFunc
	vm.bindFunc = bindFunc

	vm.getBound[MapBinding] = func(binding Binding) Value {
		symValPtr := binding.Val()
		symVal := symval(vm.read(ptr(symValPtr)))
		symValValPtr := symVal.val()
		return Value(vm.read(symValValPtr))
	}

	vm.setBound[MapBinding] = func(binding Binding, value Value) {
		symValPtr := ptr(binding.Val())
		symVal := symval(vm.read(symValPtr))
		p := vm.alloc(cell(value))
		vm.write(symValPtr, cell(makeSymval(symVal.sym(), p)))
	}

	vm.getBound[StackBinding] = func(binding Binding) Value {
		offset := binding.Val()
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

func (vm *VM) GetSymbolID(sym string) uint {
	id, ok := vm.symbols[sym]
	if !ok {
		vm.nextSymbol++
		id = vm.nextSymbol
		vm.symbols[sym] = id
		vm.InverseSymbols[id] = sym
	}
	return id
}

func loadNative(vm *VM) Value {
	name := vm.Next().String().String(vm)
	f := vm.Library.getFunction(name)
	return vm.addNative(f)
}

func (vm *VM) addNative(f procFunc) Value {
	id := len(vm.proc)
	vm.proc = append(vm.proc, f)
	return makeNative(id).Value()
}

func (vm *VM) toString(value Value) string {
	kind := value.Kind()
	return vm.toStringFunc[kind](vm, value)
}

// B I N D I N G S

func bind(vm *VM, block Block, factory bindFactory) {
	for i := block.First(vm); i != 0; i = i.Next(vm) {
		value := i.Value(vm)
		vm.bindFunc[value.Kind()](vm, value, factory)
	}
}

func (vm *VM) bind(block Block) {
	bind(vm, block, func(sym sym, create bool) Binding {
		symValPtr := vm.Dictionary.Find(vm, sym)
		if symValPtr == 0 {
			if create {
				// fmt.Printf("putting symbol %d - %s\n", sym, vm.inverseSymbols[sym])
				vm.Dictionary.Put(vm, sym, 0)
				symValPtr = vm.Dictionary.Find(vm, sym) // TODO: fix this garbage
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
	vm.pc = block.First(vm)
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

func (vm *VM) BindAndExec(block Block) Value {
	// block := Block(vm.read(ptr(code)))
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

func (vm *VM) ReadNext() Value {
	entry := blockEntry(vm.read(ptr(vm.pc)))
	value := Value(vm.read(entry.pval()))
	vm.pc = entry.next()
	vm.result = value
	return value
}

func (vm *VM) Next() Value {
	return vm.nextNoInfix()
}

type Pkg struct {
	name string
	fn   map[string]procFunc
}

type Library struct {
	packages []*Pkg
}

func (l *Library) Add(pkg *Pkg) {
	l.packages = append(l.packages, pkg)
}

func (l *Library) getFunction(name string) procFunc {
	s := strings.Split(name, "/")
	for _, p := range l.packages {
		if p.name == s[0] {
			return p.fn[s[1]]
		}
	}
	panic("function not found")
}

func NewPackage(name string) *Pkg {
	return &Pkg{name: name, fn: make(map[string]procFunc)}
}

func (p *Pkg) AddFunc(name string, fn procFunc) {
	p.fn[name] = fn
}

// func (vm *VM) addNativeFunc(name string, f procFunc) {
// 	vm.procNames = append(vm.procNames, name)
// 	vm.dictionary.Put(vm, vm.getSymbolID(name), vm.alloc(cell(vm.addNative(f))))
// }

func (vm *VM) LoadPackage(pkg *Pkg, dict dict) {
	for name, fn := range pkg.fn {
		native := vm.addNative(fn)
		sym := sym(vm.GetSymbolID(name))
		dict.Put(vm, sym, native)
		vm.procNames = append(vm.procNames, pkg.name+"/"+name)
	}
}

// func (vm *VM) AddNative(name string, lib Library) {
// 	vm.procNames = append(vm.procNames, name)
// 	vm.dictionary.Put(vm, vm.getSymbolID(name), vm.alloc(cell(vm.addNative(lib.getFunction(name)))))
// }

func (vm *VM) Push(value Value) {
	vm.push(value)
}

func (vm *VM) Show() {
	fmt.Printf("Stack pointer: %d\n", vm.sp)
}

type SerialVM struct {
	Top        uint
	Dictionary dict
	Mem        []cell
	Symbols    map[string]sym
	ProcNames  []string
}

func (vm *VM) Save() []byte {
	var result bytes.Buffer

	svm := &SerialVM{Top: vm.top, Dictionary: vm.Dictionary, Mem: vm.mem, Symbols: vm.symbols, ProcNames: vm.procNames}

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

	vm := &VM{top: svm.Top, mem: svm.Mem, Dictionary: svm.Dictionary, symbols: svm.Symbols, procNames: svm.ProcNames}

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
