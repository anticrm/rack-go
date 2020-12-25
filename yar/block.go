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
	"strings"
)

// BLOCK
//-------------------------
//   FIRST  | LAST | KIND |
//-------------------------

type Block Value
type firstLast obj
type pFirstLast ptr
type blockEntry item
type pBlockEntry ptr

func makeBlock(firstLast pFirstLast) Block { return Block(makeValue(int(firstLast), BlockType)) }
func (b Block) firstLast() pFirstLast      { return pFirstLast(b.Value().Val()) }
func (b Block) Value() Value               { return Value(b) }
func (b Block) First(vm *VM) pBlockEntry   { return firstLast(vm.read(ptr(b.firstLast()))).first() }
func (v Value) Block() Block               { return Block(v) }
func (b Block) Add(vm *VM, value Value)    { b.firstLast().add(vm, value) }

func (b firstLast) first() pBlockEntry { return pBlockEntry(obj(b).val()) }
func (b firstLast) last() pBlockEntry  { return pBlockEntry(obj(b).ptr()) }

func (b blockEntry) next() pBlockEntry { return pBlockEntry(item(b).ptr()) }
func (b blockEntry) pval() ptr         { return ptr(item(b).val()) }

func (b pBlockEntry) Next(vm *VM) pBlockEntry { return blockEntry(vm.read(ptr(b))).next() }
func (b pBlockEntry) pval(vm *VM) ptr         { return blockEntry(vm.read(ptr(b))).pval() }
func (b pBlockEntry) Value(vm *VM) Value      { return Value(vm.read(b.pval(vm))) }

// func (b pBlockEntry) ptr(vm *VM) (ptr, bool) {
// 	ptrval := ptrval(vm.read(ptr(b)))
// 	if ptrval.compressed() {
// 		return 0, false
// 	}
// 	return ptrval.ptr(), true
// }

func blockBind(vm *VM, value Value, factory bindFactory) {
	bind(vm, value.Block(), factory)
}

func makeFirstLast(first pBlockEntry, last pBlockEntry) Value {
	return Value(makeObj(int(first), ptr(last), BlockType))
}

func (vm *VM) AllocBlock() Block {
	fl := pFirstLast(vm.alloc(cell(makeItem(0, 0))))
	return makeBlock(fl)
}

// func (b pBlock) first(vm *VM) pBlockEntry {
// 	return Block(vm.read(ptr(b))).First()
// }

func (b pFirstLast) addEntry(vm *VM, newLast pBlockEntry) {
	fl := firstLast(vm.read(ptr(b)))
	last := fl.last()
	if last != 0 {
		pItem(last).setPtr(vm, ptr(newLast))
		vm.write(ptr(b), cell(makeFirstLast(fl.first(), newLast)))
	} else {
		vm.write(ptr(b), cell(makeFirstLast(newLast, newLast)))
	}
}

func (b pFirstLast) addPtr(vm *VM, ptr ptr) {
	b.addEntry(vm, pBlockEntry(vm.alloc(cell(makeItem(int(ptr), 0)))))
}

func (b pFirstLast) add(vm *VM, value Value) {
	b.addPtr(vm, vm.alloc(cell(value)))
}

func blockToString(vm *VM, b Value) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf(" #%016x ", b))
	result.WriteByte('[')

	block := Block(b)
	for i := block.First(vm); i != 0; i = i.Next(vm) {
		value := i.Value(vm)
		result.WriteString(vm.ToString(value))
		result.WriteByte(' ')
	}

	result.WriteByte(']')
	return result.String()
}
