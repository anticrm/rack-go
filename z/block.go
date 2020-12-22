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

// BLOCK
//-------------------------
//   FIRST  | LAST | KIND |
//-------------------------

type block obj
type pBlock ptr
type blockEntry item
type pBlockEntry ptr

func (b block) first() pBlockEntry { return pBlockEntry(obj(b).val()) }
func (b block) last() pBlockEntry  { return pBlockEntry(obj(b).ptr()) }

func (b blockEntry) next() pBlockEntry  { return pBlockEntry(item(b).ptr()) }
func (b blockEntry) value(vm *VM) value { return ptrval(b).value(vm) }

func (b pBlockEntry) next(vm *VM) pBlockEntry { return blockEntry(vm.read(ptr(b))).next() }
func (b pBlockEntry) value(vm *VM) value      { return blockEntry(vm.read(ptr(b))).value(vm) }

func (b pBlockEntry) ptr(vm *VM) (ptr, bool) {
	ptrval := ptrval(vm.read(ptr(b)))
	if ptrval.compressed() {
		return 0, false
	}
	return ptrval.ptr(), true
}

func makeBlock(first pBlockEntry, last pBlockEntry) value {
	return makeObj(int(first), ptr(last), BlockType)
}

func (vm *VM) allocBlock() pBlock {
	return pBlock(vm.alloc(cell(makeBlock(0, 0))))
}

func (b pBlock) first(vm *VM) pBlockEntry {
	return block(vm.read(ptr(b))).first()
}

func (b pBlock) addEntry(vm *VM, newLast pBlockEntry) {
	block := block(vm.read(ptr(b)))
	last := block.last()
	if last != 0 {
		pItem(last).setPtr(vm, ptr(newLast))
		vm.write(ptr(b), cell(makeBlock(block.first(), newLast)))
	} else {
		vm.write(ptr(b), cell(makeBlock(newLast, newLast)))
	}
}

func (b pBlock) add(vm *VM, value value) {
	b.addEntry(vm, pBlockEntry(vm.allocPtrVal(value, 0)))
}

func (b pBlock) addPtr(vm *VM, ptr ptr) {
	b.addEntry(vm, pBlockEntry(vm.alloc(cell(_ptrval(ptr)))))
}
