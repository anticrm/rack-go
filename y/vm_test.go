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
	"testing"
)

func TestParse1(t *testing.T) {
	vm := NewVM(1000, 100)
	x := vm.Parse("add 1 2")
	vm.dump()
	t.Logf("pointer %d", x)
	t.Logf("%s", vm.toString(vm.read(x)))
}

func TestAddNative(t *testing.T) {
	vm := NewVM(1000, 100)
	mapPut(vm, vm.dictionary, vm.getSymbolID("native"), vm.alloc(vm.addNative(func() {})))
	vm.dump()
}

func TestBind(t *testing.T) {
	vm := NewVM(1000, 100)
	mapPut(vm, vm.dictionary, vm.getSymbolID("native"), vm.alloc(vm.addNative(func() {})))
	code := vm.Parse("native [x y]")
	t.Logf("%s", vm.toString(vm.read(code)))
	vm.dump()
	vm.Bind(vm.read(code))
	t.Log("after bindings")
	vm.dump()
}

func TestProc(t *testing.T) {
	vm := NewVM(1000, 100)
	mapPut(vm, vm.dictionary, vm.getSymbolID("one"), vm.alloc(vm.addProc(func(pc *pc) value {
		return newInteger(1)
	})))
	code := vm.Parse("one")
	t.Logf("%s", vm.toString(vm.read(code)))
	vm.dump()
	vm.Bind(vm.read(code))
	t.Log("after bindings")
	vm.dump()
	result := vm.Exec(vm.read(code))
	t.Logf("result: %016x", result)
}

func TestAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add 1 2")
	vm.Bind(vm.read(code))
	result := vm.Exec(vm.read(code))
	t.Logf("result: %016x", result)
}

func TestAddAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add add 1 2 3")
	vm.Bind(vm.read(code))
	result := vm.Exec(vm.read(code))
	t.Logf("result: %016x", result)
}

func TestFn(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("x: fn [n] [add n 10] x 5")
	vm.Bind(vm.read(code))
	t.Logf("%s\n", vm.toString(vm.read(code)))
	result := vm.Exec(vm.read(code))
	t.Logf("result: %016x", result)
}

func TestMap(t *testing.T) {
	vm := NewVM(1000, 100)
	mapPut(vm, vm.dictionary, 123, 0)
	x := mapFind(vm, vm.dictionary, 123)
	t.Logf("%d\n", x)
	mapPut(vm, vm.dictionary, 124, 0)
	y := mapFind(vm, vm.dictionary, 124)
	t.Logf("%d\n", y)
}

func TestSum(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("sum: fn [n] [either gt n 1 [add n sum sub n 1] [n]] sum 100")
	vm.Bind(vm.read(code))
	t.Logf("%s\n", vm.toString(vm.read(code)))
	result := vm.Exec(vm.read(code))
	t.Logf("result: %016x", result)
}

func BenchmarkFib(t *testing.B) {
	vm := NewVM(1000, 10000)
	BootVM(vm)
	code := vm.Parse("fib: fn [n] [either gt n 1 [add fib sub n 2 fib sub n 1] [n]] fib 40")
	vm.Bind(vm.read(code))
	t.Logf("%s\n", vm.toString(vm.read(code)))
	result := vm.Exec(vm.read(code))
	t.Logf("result: %016x", result)
}
