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
	"testing"
)

func TestBlock(t *testing.T) {
	vm := NewVM(1000, 100)
	block := vm.AllocBlock()
	block.Add(vm, MakeInt(1).Value())
	block.Add(vm, MakeInt(2).Value())
	block.Add(vm, vm.AllocBlock().Value())
	block.Add(vm, vm.AllocBlock().Value())
	vm.dump()

	// i := block.first(vm)
	// if i.value(vm) != 0x180 {
	// 	t.Error("i.value(vm) != 0x180\n")
	// }
	// i = i.next(vm)
	// if i.value(vm) != 0x280 {
	// 	t.Error("i.value(vm) != 0x280\n")
	// }
	// i = i.next(vm)
	// if i.value(vm) != 0 {
	// 	t.Error("i.value(vm) != 0x0 /1\n")
	// }
	// i = i.next(vm)
	// if i.value(vm) != 0 {
	// 	t.Error("i.value(vm) != 0x0 /2\n")
	// }
	// for i := block.first(vm); i != 0; i = i.next(vm) {
	// 	fmt.Printf("%016x\n", i.value(vm))
	// }
}
