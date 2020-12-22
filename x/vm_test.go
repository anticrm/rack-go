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

package x

import "testing"

func TestBind(t *testing.T) {
	vm := NewVM(1000, 100)
	vm.dictionary.put(vm, vm.getSymbolID("native"), vm.alloc(cell(vm.addProc(func(pc *pc) value { return 42 }))))
	code := vm.Parse("native [x y]")
	t.Logf("%s", vm.toString(value(vm.read(ptr(code)))))
	vm.dump()
	vm.bind(block(vm.read(ptr(code))))
	t.Log("after bindings")
	t.Logf("%s", vm.toString(value(vm.read(ptr(code)))))
	vm.dump()
}

func TestAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add 1 2")
	c := block(vm.read(ptr(code)))
	vm.bind(c)
	result := vm.exec(c)
	t.Logf("result: %016x", result)
}

func TestAddAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add add 1 2 3")
	c := block(vm.read(ptr(code)))
	vm.bind(c)
	result := vm.exec(c)
	t.Logf("result: %016x", result)
}

func TestFn(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("x: fn [n] [add n 10] x 5")
	c := block(vm.read(ptr(code)))
	vm.bind(c)
	t.Logf("%s", vm.toString(value(c)))
	result := vm.exec(c)
	t.Logf("result: %016x", result)
}

func TestSum(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("sum: fn [n] [either gt n 1 [add n sum sub n 1] [n]] sum 100")
	c := block(vm.read(ptr(code)))
	vm.bind(c)
	t.Logf("%s", vm.toString(value(c)))
	result := vm.exec(c)
	t.Logf("result: %016x", result)
}

func BenchmarkFib(t *testing.B) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("fib: fn [n] [either gt n 1 [add fib sub n 2 fib sub n 1] [n]] fib 40")
	c := block(vm.read(ptr(code)))
	vm.bind(c)
	vm.exec(c)
}
