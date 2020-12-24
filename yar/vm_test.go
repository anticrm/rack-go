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

import "testing"

func TestBind(t *testing.T) {
	vm := NewVM(1000, 100)
	vm.dictionary.put(vm, vm.getSymbolID("native"), vm.alloc(cell(vm.addNative(func(vm *VM) Value { return 42 }))))
	code := vm.Parse("native [x y]")
	t.Logf("%s", vm.toString(Value(vm.read(ptr(code)))))
	vm.dump()
	vm.bind(Block(vm.read(ptr(code))))
	t.Log("after bindings")
	t.Logf("%s", vm.toString(Value(vm.read(ptr(code)))))
	vm.dump()
}

func TestAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add 1 2")
	c := Block(vm.read(ptr(code)))
	vm.bind(c)
	result := vm.call(c)
	t.Logf("result: %016x", result)
}

func TestAddAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add add 1 2 3")
	c := Block(vm.read(ptr(code)))
	vm.bind(c)
	result := vm.call(c)
	t.Logf("result: %016x", result)
}

func TestFn(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("x: fn [n] [add n 10] x 5")
	c := Block(vm.read(ptr(code)))
	vm.bind(c)
	t.Logf("%s", vm.toString(Value(c)))
	result := vm.call(c)
	t.Logf("result: %016x", result)
}

func TestSum(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("sum: fn [n] [either gt n 1 [add n sum sub n 1] [n]] sum 100")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestGetWord(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("sum: fn [n] [add n n] x: 5 :x")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestFork(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	vm.AddNative("fork", func(vm *VM) Value {
		fn := Proc(vm.Next())
		clone := vm.Clone()
		stack := []Value{makeInt(42), makeInt(41)}
		fork := clone.Fork(stack, uint(len(stack)))
		return fork.Exec(fn.First())
	})
	code := vm.Parse("sum: fn [x y] [add x y] fork :sum")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestMakeObject(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("o: make-object [a: 1 b: 2]")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestPath(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("o: make-object [a: 42 b: 2] o/a")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestPath2(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("o: make-object [a: 42 b: make-object [c: 55]] o/b/c")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func BenchmarkFib(t *testing.B) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("fib: fn [n] [either gt n 1 [add fib sub n 2 fib sub n 1] [n]] fib 40")
	c := Block(vm.read(ptr(code)))
	vm.bind(c)
	vm.call(c)
}
