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
	LastType    = iota
)

// There are 4 types of cell layouts in the VM: value, obj, item, and ptrval

// VALUE
//-------------------------
//  X X X X X X X  | KIND |
//-------------------------

type Value cell

// func (v value) val() int        { return int(v >> 32) }
// func (v value) ptr() ptr        { return ptr(v&0xffffffff) >> 8 }
func (v Value) Kind() int { return int(v & 0xff) }

// func (v value) immutable() bool { return v.kind()&0x80 != 0 }

// func makeValue(val int, ptr ptr, kind int) value {
// 	return value(cell(val)<<32 | cell(ptr)<<8 | cell(kind))
// }

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

type imm = Value

func makeImm(value int, kind int) imm {
	return imm(value<<8 | kind)
}

func makeInt(value int) imm {
	return makeImm(value, IntegerType)
}

func makeBool(value bool) imm {
	if value {
		return makeImm(1, BooleanType)
	}
	return makeImm(0, BooleanType)
}

func intValue(value imm) int {
	return int(value >> 8)
}

func boolValue(value imm) bool {
	return intValue(value) != 0
}

func makeNative(value int) imm {
	return makeImm(value, NativeType)
}

func nativeExec(vm *VM, value Value) Value {
	i := intValue(value)
	f := vm.proc[i]
	return f(vm)
}

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
