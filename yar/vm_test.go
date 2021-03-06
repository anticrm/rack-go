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
	"testing"
)

func TestBind(t *testing.T) {
	vm := NewVM(1000, 100)
	vm.Dictionary.Put(vm, vm.GetSymbolID("native"), vm.addNative(func(vm *VM) Value { return 42 }))
	code := vm.Parse("native [x y]")
	t.Logf("%s", vm.ToString(code.Value()))
	vm.Dump()
	vm.bind(code)
	t.Log("after bindings")
	t.Logf("%s", vm.ToString(code.Value()))
	vm.Dump()
}

func TestAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add 1 2")
	vm.bind(code)
	result := vm.call(code)
	t.Logf("result: %016x", result)
}

func TestAddAdd(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("add add 1 2 3")
	vm.bind(code)
	result := vm.call(code)
	t.Logf("result: %016x", result)
}

func TestFn(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("x: fn [n] [add n 10] x 5")
	vm.bind(code)
	t.Logf("%s", vm.ToString(code.Value()))
	result := vm.call(code)
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
	// vm := NewVM(1000, 100)
	// BootVM(vm)
	// vm.addNativeFunc("fork", func(vm *VM) Value {
	// 	fn := Proc(vm.Next())
	// 	clone := vm.Clone()
	// 	stack := []Value{makeInt(42), makeInt(41)}
	// 	fork := clone.Fork(stack, uint(len(stack)))
	// 	return fork.Exec(fn.First())
	// })
	// code := vm.Parse("sum: fn [x y] [add x y] fork :sum")
	// result := vm.BindAndExec(code)
	// t.Logf("result: %016x", result)
}

func TestMakeObject(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("o: make-object [a: 1 b: 2 c: add 5 5] o/c")
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

func TestStrings(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("o: [\"a\" \"b\"]")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestForeach(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("foreach val [\"a\" \"b\"] [print val]")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestAppend(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("print append [\"a\" \"b\"] \"c\"")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestError(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("unknown 5")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestIn(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("x: make-object [a: 41 b: 2] get in x 'a")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestSetPath(t *testing.T) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("x: make-object [a: 41 b: 2] x/a: 256 print x/a")
	fmt.Println(vm.ToString(code.Value()))
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
}

func TestSave(t *testing.T) {
	// vm := NewVM(1000, 100)
	// BootVM(vm)
	// code := vm.Parse("sum: fn [n] [add n n] sum 5")
	// data := vm.Save()
	// lib := &pkg{lib: coreLibrary}
	// vm2 := LoadVM(data, 100, lib)
	// //BootVM(vm2)
	// fmt.Printf("%+v\n\n", vm)
	// fmt.Printf("%+v\n", vm2)
	// result := vm2.BindAndExec(code)
	// t.Logf("result: %016x", result)
}

func BenchmarkFib(t *testing.B) {
	vm := NewVM(1000, 100)
	BootVM(vm)
	code := vm.Parse("fib: fn [n] [either gt n 1 [add fib sub n 2 fib sub n 1] [n]] fib 40")
	vm.BindAndExec(code)
}
