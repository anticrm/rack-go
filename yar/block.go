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

import "strings"

// BLOCK
//-------------------------
//   FIRST  | LAST | KIND |
//-------------------------

type Block obj
type pBlock ptr
type blockEntry item
type pBlockEntry ptr

func (b Block) First() pBlockEntry { return pBlockEntry(obj(b).val()) }
func (b Block) last() pBlockEntry  { return pBlockEntry(obj(b).ptr()) }

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

func makeBlock(first pBlockEntry, last pBlockEntry) Value {
	return Value(makeObj(int(first), ptr(last), BlockType))
}

func (vm *VM) allocBlock() pBlock {
	return pBlock(vm.alloc(cell(makeBlock(0, 0))))
}

// func (b pBlock) first(vm *VM) pBlockEntry {
// 	return Block(vm.read(ptr(b))).First()
// }

func (b pBlock) addEntry(vm *VM, newLast pBlockEntry) {
	block := Block(vm.read(ptr(b)))
	last := block.last()
	if last != 0 {
		pItem(last).setPtr(vm, ptr(newLast))
		vm.write(ptr(b), cell(makeBlock(block.First(), newLast)))
	} else {
		vm.write(ptr(b), cell(makeBlock(newLast, newLast)))
	}
}

// func (b pBlock) add(vm *VM, value value) {
// 	b.addEntry(vm, pBlockEntry(vm.allocPtrVal(value, 0)))
// }

func (b pBlock) add(vm *VM, ptr ptr) {
	b.addEntry(vm, pBlockEntry(vm.alloc(cell(makeItem(int(ptr), 0)))))
}

func blockToString(vm *VM, b Value) string {
	var result strings.Builder
	result.WriteByte('[')

	block := Block(b)
	for i := block.First(); i != 0; i = i.Next(vm) {
		value := i.Value(vm)
		result.WriteString(vm.toString(value))
		result.WriteByte(' ')
	}

	result.WriteByte(']')
	return result.String()
}
