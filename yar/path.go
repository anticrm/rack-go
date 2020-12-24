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

type pathEntry obj
type pPathEntry ptr

func (b pathEntry) next() pPathEntry { return pPathEntry(obj(b).ptr()) }
func (b pathEntry) sym() sym         { return sym(obj(b).val()) }

func (b pPathEntry) Next(vm *VM) pPathEntry { return pathEntry(vm.read(ptr(b))).next() }
func (b pPathEntry) sym(vm *VM) sym         { return pathEntry(vm.read(ptr(b))).sym() }

type pathBuilder struct {
	vm    *VM
	first pPathEntry
	last  pPathEntry
}

func (b pathBuilder) add(sym sym) {
	vm := b.vm
	next := pPathEntry(vm.alloc(cell(makeObj(int(sym), 0, PathType))))
	if b.last == 0 {
		b.first = next
		b.last = next
	} else {
		cur := pPathEntry(vm.read(ptr(b.last)))
		vm.write(ptr(b.last), cell(makeObj(int(cur.sym(vm)), ptr(next), PathType)))
		b.last = next
	}
}

func pathToString(vm *VM, p pathEntry) string {
	var result strings.Builder

	result.WriteString(vm.InverseSymbols[p.sym()])
	i := p.next()

	for i != 0 {
		result.WriteByte('/')
		result.WriteString(vm.InverseSymbols[i.sym(vm)])
		i = i.Next(vm)
	}

	return result.String()
}
