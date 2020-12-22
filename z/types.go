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

type ptr uint
type cell int64

const (
	BlockType   = iota
	WordType    = iota
	GetWordType = iota
	SetWordType = iota
	QuoteType   = iota
	MapType     = iota
	IntegerType = 0x80
	ProcType    = iota
	LastType    = iota
)

// There are 4 types of cell layouts in the VM: value, obj, item, and ptrval

// VALUE
//-------------------------
//  X X X X X X X  | KIND |
//-------------------------

type value cell

// func (v value) val() int        { return int(v >> 32) }
// func (v value) ptr() ptr        { return ptr(v&0xffffffff) >> 8 }
func (v value) kind() int       { return int(v & 0xff) }
func (v value) immutable() bool { return v.kind()&0x80 != 0 }

// func makeValue(val int, ptr ptr, kind int) value {
// 	return value(cell(val)<<32 | cell(ptr)<<8 | cell(kind))
// }

// OBJ
//-------------------------
//  VAL     | PTR  | KIND |
//-------------------------

type obj = value

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

func makeItem(val ptr, ptr ptr) item { return item(val<<32 | ptr) }

func (i item) ptr() ptr { return ptr(i & 0xffffffff) }
func (i item) val() ptr { return ptr(i >> 32) }

func (e pItem) ptr(vm *VM) ptr { return item(vm.read(ptr(e))).ptr() }
func (e pItem) val(vm *VM) ptr { return item(vm.read(ptr(e))).val() }

func (e pItem) setPtr(vm *VM, next ptr) {
	entryValue := vm.read(ptr(e))
	newValue := (uint64(entryValue) & 0xffffffff80000000) | uint64(next)
	vm.write(ptr(e), cell(newValue))
}

// PTRVAL
//--------------------------------------
// S |  PTR / VAL | C | X X X X X X X  |
//--------------------------------------

type ptrval cell
type pPtrval ptr

func (i ptrval) compressed() bool { return int(i>>32)&1 == 1 }
func (i ptrval) ptr() ptr         { return ptr(i >> 33) }

func (e pPtrval) value(vm *VM) value {
	return ptrval(vm.read(ptr(e))).value(vm)
}

func (e pPtrval) setPtr(vm *VM, value ptr) {
	cur := vm.read(ptr(e))
	newVal := cell(_ptrval(value)) | (cur & 0xffffffff)
	vm.write(ptr(e), newVal)
}

func (e pPtrval) setValue(vm *VM, value value) {
	cur := vm.read(ptr(e))
	if canCompress(value) {
		newVal := cell(_compressedPtrval(value)) | (cur & 0xffffffff)
		vm.write(ptr(e), newVal)
	} else {
		if value.immutable() {
			ptr := ptrval(cur).ptr()
			vm.write(ptr, cell(value))
		} else {

		}
	}
}

func (i ptrval) value(vm *VM) value {
	if i.compressed() {
		return value(i >> 33)
	}
	return value(vm.read(ptr(i >> 33)))
}

func canCompress(value value) bool { return value.immutable() && uint64(value)&0x7fffffff00000000 == 0 }

func _compressedPtrval(value value) ptrval { return ptrval(value<<33 | 0x100000000) }
func _ptrval(ptr ptr) ptrval               { return ptrval(ptr << 33) }

func makePtrVal(value ptr, rest ptr) ptrval { return ptrval(value<<33 | rest) }

func (vm *VM) allocPtrVal(value value, rest ptr) pPtrval {
	if canCompress(value) {
		return pPtrval(vm.alloc(cell(_compressedPtrval(value)) | cell(rest)))
	}
	return pPtrval(vm.alloc(cell(_ptrval(vm.alloc(cell(value)))) | cell(rest)))
}

// P R I M I T I V E S

type imm = value

func makeImm(value int, kind int) imm {
	return imm(value<<8 | kind)
}

func makeInt(value int) imm {
	return makeImm(value, IntegerType)
}

func intValue(value imm) int {
	return int(value >> 8)
}

func makeProc(value int) imm {
	return makeImm(value, ProcType)
}
