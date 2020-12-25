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

import "strconv"

type ptr uint
type cell int64

const (
	BlockType   = iota
	WordType    = iota
	GetWordType = iota
	SetWordType = iota
	QuoteType   = iota
	MapType     = iota
	IntegerType = iota
	BooleanType = iota
	NativeType  = iota
	ProcType    = iota
	PathType    = iota
	GetPathType = iota
	StringType  = iota
	LastType    = iota
)

// There are 4 types of cell layouts in the VM: value, obj, item, and ptrval

// VALUE
//------------------------------
//       V A L U E      | KIND |
//------------------------------

// Value represent any value within VM
type Value cell

// Kind return value type
func (v Value) Kind() int { return int(v & 0xff) }

// Val returns embedded value, interpretation depends on value type
func (v Value) Val() int                { return int(v >> 8) }
func makeValue(val int, kind int) Value { return Value(val)<<8 | Value(kind) }

// BINDING
//------------------------------
//       V A L U E      | KIND |
//------------------------------

// Value represent any value within VM
type Binding cell
type pBinding ptr

// Kind return value type
func (b Binding) Kind() int { return int(b & 0xff) }

// Val returns embedded value, interpretation depends on value type
func (b Binding) Val() int                  { return int(b >> 8) }
func makeBinding(val int, kind int) Binding { return Binding(val)<<8 | Binding(kind) }

// OBJ
//-------------------------
//  VAL     | PTR  | KIND |
//-------------------------

type obj Value

func (v obj) val() int { return int(v >> 32) }
func (v obj) ptr() ptr { return ptr(v&0xffffffff) >> 8 }

func makeObj(val int, ptr ptr, kind int) obj {
	return obj(cell(val)<<32 | cell(ptr)<<8 | cell(kind))
}

// ITEM
//-----------------------------
// V A L       |       P T R  |
//-----------------------------

type item cell
type pItem ptr

func makeItem(val int, ptr ptr) item { return item(val<<32) | item(ptr) }

func (i item) val() int { return int(i >> 32) }
func (i item) ptr() ptr { return ptr(i & 0xffffffff) }

func (e pItem) ptr(vm *VM) ptr { return item(vm.read(ptr(e))).ptr() }
func (e pItem) val(vm *VM) int { return item(vm.read(ptr(e))).val() }

func (e pItem) setPtr(vm *VM, next ptr) {
	entryValue := vm.read(ptr(e))
	newValue := (uint64(entryValue) & 0xffffffff00000000) | uint64(next)
	vm.write(ptr(e), cell(newValue))
}

// P R I M I T I V E S

type imm Value

func makeImm(value int, kind int) imm {
	return imm(value<<8 | kind)
}

func (i imm) Val() int {
	return int(i >> 8)
}

type integer Value

func MakeInt(value int) integer {
	return integer(makeValue(value, IntegerType))
}
func (i integer) Value() Value { return Value(i) }

func intToString(vm *VM, b Value) string {
	return strconv.Itoa(b.Val())
}

type boolean Value

func MakeBool(value bool) boolean {
	if value {
		return boolean(makeValue(1, BooleanType))
	}
	return boolean(makeValue(0, BooleanType))
}
func (b boolean) Value() Value { return Value(b) }
func (b boolean) Val() bool    { return b.Value().Val() != 0 }
func (v Value) Bool() boolean  { return boolean(v) }

type native Value

func makeNative(value int) native {
	return native(makeValue(value, NativeType))
}

func (v Value) native() native { return native(v) }
func (n native) Value() Value  { return Value(n) }

func nativeExec(vm *VM, value Value) Value {
	i := value.Val()
	f := vm.proc[i]
	return f(vm)
}

///

///

type Proc obj

func makeProc(stackSize int, code ptr) Value {
	return Value(makeObj(stackSize, code, ProcType))
}

func (p Proc) StackSize() int     { return obj(p).val() }
func (p Proc) First() pBlockEntry { return pBlockEntry(obj(p).ptr()) }

func procExec(vm *VM, value Value) Value {
	p := Proc(value)
	stackSize := p.StackSize()
	code := p.First()
	for i := 0; i < stackSize; i++ {
		vm.push(vm.Next())
	}
	result := vm.Exec(code)
	vm.sp -= uint(stackSize)
	return result
}

///

// V M T

func identity(vm *VM, value Value) Value                          { return value }
func execNotImplemented(vm *VM, value Value) Value                { panic("not implemented") }
func bindNotImplemented(vm *VM, value Value, factory bindFactory) { panic("not implemented") }
func bindNothing(vm *VM, value Value, factory bindFactory)        {}

var (
	execFunc = []func(vm *VM, value Value) Value{
		identity,
		wordExec,
		getWordExec,
		setWordExec,
		execNotImplemented,
		identity, // map
		identity, // int
		identity, // bool
		nativeExec,
		procExec,
		pathExec,
		getPathExec,
		identity,
	}

	bindFunc = []func(vm *VM, value Value, factory bindFactory){
		blockBind,
		wordBind,
		wordBind,
		setWordBind,
		bindNotImplemented,
		bindNothing, // map
		bindNothing,
		bindNothing,
		bindNothing,
		bindNothing,
		pathBind,
		pathBind,
		bindNothing,
	}
)
