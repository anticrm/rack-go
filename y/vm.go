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

package y

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	BlockType    = iota
	MapType      = iota
	SmallIntType = iota
	IntegerType  = iota
	BoolType     = iota
	WordType     = iota
	GetWordType  = iota
	SetWordType  = iota
	QuoteType    = iota
	NativeType   = iota
	ProcType     = iota
	LastType     = iota
)

type cell uint64
type value = cell
type ptr uint

func kind(cell cell) int {
	return int(cell & 0xff)
}

func car(cell cell) ptr {
	return ptr(cell >> 40)
}

func cdr(cell cell) ptr {
	return ptr(cell>>16) & 0xffffff
}

func newCell(car ptr, cdr ptr, kind int) cell {
	return cell(car)<<40 | cell(cdr)<<16 | cell(kind)
}

func data(car ptr, cdr uint) cell {
	return cell(car)<<32 | cell(cdr)
}

func dataCdr(cell cell) uint {
	return uint(cell & 0xffffffff)
}

func dataCar(cell cell) ptr {
	return ptr(cell >> 32)
}

// T O  S T R I N G

func notImplemented(vm *VM, cell cell) string {
	return "<not implemented>"
}

// B L O C K

func canCompress(value value) bool { return value&0x7fffffff == value }

func asEmbedded(value value) uint  { return uint(value<<1 | 0x1) }
func asPointer(ptr ptr) uint       { return uint(ptr) << 1 }
func isEmbedded(data uint) bool    { return data&0x01 == 0x01 }
func extractPtr(data uint) ptr     { return ptr(data >> 1) }
func extractValue(data uint) value { return value(data >> 1) }

func blockEntryAsIs(next ptr, value uint) cell      { return data(next, value) }
func blockEntryPtr(next ptr, value ptr) cell        { return data(next, asPointer(value)) }
func blockEntryEmbedded(next ptr, value value) cell { return data(next, asEmbedded(value)) }

func newBlockEntryPtr(vm *VM, next ptr, ptr ptr) ptr {
	return vm.alloc(blockEntryPtr(next, ptr))
}
func newBlockEntryValue(vm *VM, next ptr, value value) ptr {
	return vm.alloc(blockEntryEmbedded(next, value))
}

func blockEntryNext(cell cell) ptr { return dataCar(cell) }
func blockEntryValue(vm *VM, cell cell) cell {
	data := dataCdr(cell)
	if isEmbedded(data) {
		return extractValue(data)
	}
	return vm.read(extractPtr(data))
}
func blockEntryValueAsIs(cell cell) uint { return dataCdr(cell) }

func newBlock(first ptr, last ptr) cell { return newCell(first, last, BlockType) }
func blockFirst(block cell) ptr         { return car(block) }
func blockLast(block cell) ptr          { return ptr(cdr(block)) }

func blockAddEntry(vm *VM, block ptr, entry ptr) {
	b := vm.read(block)
	first := blockFirst(b)
	last := blockLast(b)
	if last != 0 {
		lastValue := blockEntryValueAsIs(vm.read(last))
		vm.write(last, blockEntryAsIs(entry, lastValue))
	}
	if first == 0 {
		first = entry
	}
	vm.write(block, newBlock(first, entry))
}

func blockAdd(vm *VM, block ptr, value value) {
	if canCompress(value) {
		blockAddEntry(vm, block, newBlockEntryValue(vm, 0, value))
	} else {
		blockAddEntry(vm, block, newBlockEntryPtr(vm, 0, vm.alloc(value)))
	}
}

func blockAddPtr(vm *VM, block ptr, ptr ptr) {
	blockAddEntry(vm, block, newBlockEntryPtr(vm, 0, ptr))
}

func blockToString(vm *VM, block cell) string {
	var result strings.Builder
	result.WriteByte('[')
	entryPtr := blockFirst(block)
	for entryPtr != 0 {
		entry := vm.read(entryPtr)
		value := blockEntryValue(vm, entry)
		result.WriteString(vm.toString(value))
		result.WriteByte(' ')
		entryPtr = blockEntryNext(entry)
	}
	result.WriteByte(']')
	return result.String()
}

// M A P

func newMapEntry(next ptr, symval ptr) cell { return data(next, uint(symval)) }
func mapEntryNext(cell cell) ptr            { return dataCar(cell) }
func mapEntrySymVal(cell cell) ptr          { return ptr(dataCdr(cell)) }

func newSymVal(sym sym, val ptr) cell { return data(val, sym) }
func symValSym(symval cell) sym       { return dataCdr(symval) }
func symValVal(symval cell) ptr       { return dataCar(symval) }

func newMap(first ptr, last ptr) cell { return newCell(first, last, MapType) }
func mapFirst(block cell) ptr         { return car(block) }
func mapLast(block cell) ptr          { return ptr(cdr(block)) }

func mapPut(vm *VM, m ptr, sym sym, value ptr) {
	_map := vm.read(m)
	first := mapFirst(_map)
	last := mapLast(_map)

	entryPtr := first
	for entryPtr != 0 {
		entry := vm.read(entryPtr)
		symValPtr := mapEntrySymVal(entry)
		symval := vm.read(symValPtr)
		_sym := symValSym(symval)
		if _sym == sym {
			vm.write(symValPtr, newSymVal(_sym, value))
			break
		}
		entryPtr = mapEntryNext(entry)
	}

	newSymVal := vm.alloc(newSymVal(sym, value))
	newEntry := vm.alloc(newMapEntry(0, newSymVal))
	if last != 0 {
		entry := vm.read(last)
		symval := mapEntrySymVal(entry)
		vm.write(last, newMapEntry(newEntry, symval))
	}
	if first == 0 {
		first = newEntry
	}
	vm.write(m, newMap(first, newEntry))
}

func mapFind(vm *VM, m ptr, sym sym) ptr {
	_map := vm.read(m)
	first := mapFirst(_map)

	entryPtr := first
	for entryPtr != 0 {
		entry := vm.read(entryPtr)
		symValPtr := mapEntrySymVal(entry)
		symval := vm.read(symValPtr)
		_sym := symValSym(symval)
		if _sym == sym {
			return symValPtr
		}
		entryPtr = mapEntryNext(entry)
	}
	return 0
}

func mapToString(vm *VM, block cell) string {
	var result strings.Builder
	result.WriteByte('[')
	entryPtr := mapFirst(block)
	for entryPtr != 0 {
		entry := vm.read(entryPtr)
		symval := vm.read(mapEntrySymVal(entry))
		result.WriteString(vm.inverseSymbols[symValSym(symval)])
		result.WriteByte(':')
		result.WriteString(vm.toString(vm.read(symValVal(symval))))
		result.WriteByte(' ')
		entryPtr = blockEntryNext(entry)
	}
	result.WriteByte(']')
	return result.String()
}

// I N T E G E R

func newInteger(value int) cell {
	return cell(value)<<16 | cell(IntegerType)
}

func intValue(value value) int { return int(value >> 16) }

func intToString(vm *VM, cell cell) string {
	i := intValue(cell)
	return strconv.Itoa(i)
}

func newBoolean(value bool) cell {
	v := 0
	if value {
		v = 1
	}
	return cell(v)<<16 | cell(BoolType)
}

func boolValue(value value) bool { return value>>16 != 0 }

func boolToString(vm *VM, cell cell) string {
	b := boolValue(cell)
	if b {
		return "true"
	}
	return "false"
}

// W O R D

func newWord(vm *VM, sym string) cell {
	return newCell(ptr(vm.getSymbolID(sym)), 0, WordType)
}

func word(sym sym, bindings ptr) cell { return newCell(ptr(sym), bindings, WordType) }

func wordSym(word value) sym      { return sym(car(word)) }
func wordBindings(word value) ptr { return cdr(word) }

func wordBind(vm *VM, ptr ptr, factory bindFactory) {
	_word := vm.read(ptr)
	sym := wordSym(_word)
	bindings := factory(sym, false)
	if bindings != 0 {
		vm.write(ptr, word(sym, vm.alloc(bindings)))
	}
}

func wordExec(pc *pc, value value) value {
	bindings := pc.vm.read(wordBindings(value))
	if bindings == 0 {
		fmt.Printf("word %s\n", pc.vm.inverseSymbols[wordSym(value)])
		panic("word not bound")
	}
	bindingKind := bindingsKind(bindings)
	bound := pc.vm.getBound[bindingKind](bindings)
	return pc.vm.execFunc[kind(bound)](pc, bound)
}

func wordToString(vm *VM, value value) string {
	var result strings.Builder
	result.WriteString(vm.inverseSymbols[wordSym(value)])
	result.WriteString("(")
	result.WriteString(strconv.Itoa(int(vm.read(wordBindings(value)))))
	result.WriteString(")")
	return result.String()
}

// S E T W O R D

func newSetWord(vm *VM, sym string) cell {
	return newCell(ptr(vm.getSymbolID(sym)), 0, SetWordType)
}

func setWord(sym sym, bindings ptr) cell { return newCell(ptr(sym), bindings, SetWordType) }

func setWordBind(vm *VM, ptr ptr, factory bindFactory) {
	_word := vm.read(ptr)
	sym := wordSym(_word)
	// fmt.Printf("create bindings %s\n", vm.inverseSymbols[sym])
	bindings := factory(sym, true)
	vm.write(ptr, setWord(sym, vm.alloc(bindings)))
}

func setWordExec(pc *pc, value value) value {
	bindings := pc.vm.read(wordBindings(value))
	if bindings == 0 {
		panic("word not bound")
	}
	result := pc.next()
	bindingKind := bindingsKind(bindings)
	pc.vm.setBound[bindingKind](bindings, result)
	return result
}

// Code - pointer to code block
type Code = ptr
type sym = uint
type bound = cell

type bindFactory func(sym sym, create bool) bound

type nativeFunc func()
type procFunc func(pc *pc) value

// VM - VM
type VM struct {
	_Mem           []cell
	_Stack         []cell
	sp             int
	top            ptr
	result         cell
	dictionary     ptr
	symbols        map[string]sym
	nextSymbol     uint
	inverseSymbols map[sym]string
	native         []nativeFunc
	proc           []procFunc
	toStringFunc   [LastType]func(vm *VM, cell cell) string
	bindFunc       [LastType]func(vm *VM, ptr ptr, factory bindFactory)
	execFunc       [LastType]func(pc *pc, value value) value
	getBound       [LastBinding]func(bindings cell) value
	setBound       [LastBinding]func(bindings cell, value value)
}

func (vm *VM) read(ptr ptr) cell {
	return vm._Mem[ptr]
}

func (vm *VM) write(ptr ptr, value value) {
	if ptr == 0 {
		panic("null pointer assignment")
	}
	vm._Mem[ptr] = value
}

func (vm *VM) push(value value) {
	vm._Stack[vm.sp] = value
	vm.sp++
}

func (vm *VM) pop() value {
	vm.sp--
	return vm._Stack[vm.sp]
}

// NewVM - NewVM
func NewVM(memSize int, stackSize int) *VM {
	vm := &VM{
		_Mem:           make([]cell, memSize),
		_Stack:         make([]cell, stackSize),
		sp:             0,
		top:            1,
		nextSymbol:     1,
		symbols:        make(map[string]uint),
		inverseSymbols: make(map[uint]string),
	}
	for i := 0; i < LastType; i++ {
		vm.toStringFunc[i] = notImplemented
	}
	vm.toStringFunc[BlockType] = blockToString
	vm.toStringFunc[IntegerType] = intToString
	vm.toStringFunc[WordType] = wordToString

	vm.bindFunc[WordType] = wordBind
	vm.bindFunc[SetWordType] = setWordBind
	vm.bindFunc[BlockType] = func(vm *VM, ptr ptr, factory bindFactory) {
		bind(vm, vm.read(ptr), factory)
	}
	vm.bindFunc[IntegerType] = func(vm *VM, ptr ptr, factory bindFactory) {}

	vm.execFunc[WordType] = wordExec
	vm.execFunc[SetWordType] = setWordExec
	vm.execFunc[ProcType] = procExec
	vm.execFunc[BlockType] = func(pc *pc, value value) value { return value }
	vm.execFunc[IntegerType] = func(pc *pc, value value) value { return value }

	vm.getBound[MapBinding] = func(binding cell) value {
		symValPtr := mapBindingPtr(binding)
		symVal := vm.read(symValPtr)
		symValValPtr := symValVal(symVal)
		value := vm.read(symValValPtr)
		return value
	}

	vm.getBound[StackBinding] = func(binding cell) value {
		offset := stackBindingOffset(binding)
		return vm._Stack[vm.sp-offset]
	}

	vm.setBound[MapBinding] = func(binding cell, value value) {
		symValPtr := mapBindingPtr(binding)
		symVal := vm.read(symValPtr)
		symValValPtr := symValVal(symVal)
		vm.write(symValValPtr, value)
	}

	vm.dictionary = vm.alloc(newMap(0, 0))
	return vm
}

func (vm *VM) alloc(cell cell) ptr {
	result := vm.top
	vm._Mem[vm.top] = cell
	vm.top++
	return result
}

func (vm *VM) getSymbolID(sym string) uint {
	id, ok := vm.symbols[sym]
	if !ok {
		id = vm.nextSymbol
		vm.nextSymbol++
		vm.symbols[sym] = id
		vm.inverseSymbols[id] = sym
	}
	return id
}

func (vm *VM) toString(cell cell) string {
	kind := kind(cell)
	return vm.toStringFunc[kind](vm, cell)
}

const (
	MapBinding   = iota
	StackBinding = iota
	LastBinding  = iota
)

func newMapBinding(symValPtr ptr) cell {
	result := cell(symValPtr<<16 | MapBinding) //  data(symValPtr, MapBinding)
	kind := bindingsKind(result)
	if kind != MapBinding {
		panic("conversion issues")
	}
	p := mapBindingPtr(result)
	if p != symValPtr {
		panic("conversion issues 3")
	}
	return result
}
func mapBindingPtr(cell cell) ptr {
	return ptr(cell >> 16)
}

func bindingsKind(value value) int { return int(value & 0xffff) }

func newStackBindins(offset int) cell {
	result := cell(-offset<<16 | StackBinding)
	kind := bindingsKind(result)
	if kind != StackBinding {
		panic("conversion issues")
	}
	o := stackBindingOffset(result)
	if o != -offset {
		fmt.Printf("O: %d\n", o)
		panic("conversion issues 22")
	}
	return result
}

func stackBindingOffset(bindings cell) int { return int(bindings >> 16) }

func bind(vm *VM, block cell, factory bindFactory) {
	entryPtr := blockFirst(block)
	for entryPtr != 0 {
		entry := vm.read(entryPtr)
		data := blockEntryValueAsIs(entry)
		if !isEmbedded(data) {
			ptr := extractPtr(data)
			kind := kind(vm.read(ptr))
			vm.bindFunc[kind](vm, ptr, factory)
		} else {
			// fmt.Printf("skip bind %16x", extractValue(data))
		}
		entryPtr = blockEntryNext(entry)
	}
}

func (vm *VM) Bind(block cell) {
	bind(vm, block, func(sym sym, create bool) bound {
		symValPtr := mapFind(vm, vm.dictionary, sym)
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

func newNative(value int) cell {
	return cell(value)<<16 | cell(NativeType)
}

func nativeToString(vm *VM, cell cell) string {
	i := int(cell >> 16)
	return fmt.Sprintf("<native #%02x>", i)
}

func (vm *VM) addNative(f nativeFunc) cell {
	id := len(vm.native)
	vm.native = append(vm.native, f)
	return newNative(id)
}

func newProc(value int) cell {
	return cell(value)<<16 | cell(ProcType)
}

func getProcFunc(vm *VM, value value) procFunc {
	i := int(value >> 16)
	return vm.proc[i]
}

func procToString(vm *VM, cell cell) string {
	i := int(cell >> 16)
	return fmt.Sprintf("<proc #%02x>", i)
}

func (vm *VM) addProc(f procFunc) cell {
	id := len(vm.proc)
	vm.proc = append(vm.proc, f)
	return newProc(id)
}

func procExec(pc *pc, value value) value {
	i := int(value >> 16)
	f := pc.vm.proc[i]
	return f(pc)
}

// P C

type pc struct {
	pc ptr
	vm *VM
}

type pcAlias = pc

func newPC(vm *VM, block cell) *pc {
	first := blockFirst(block)
	return &pc{vm: vm, pc: first}
}

func (pc *pc) nextNoInfix() value {
	entry := pc.vm.read(pc.pc)
	value := blockEntryValue(pc.vm, entry)
	pc.pc = blockEntryNext(entry)
	kind := kind(value)
	result := pc.vm.execFunc[kind](pc, value)
	pc.vm.result = result
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

func (vm *VM) Exec(block value) value {
	return newPC(vm, block).exec()
}
