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

func TestParse1(t *testing.T) {
	x := Parse("add 1 2")
	if (*x[1].(*IntValue) != IntValue{Value: 1}) {
		t.Errorf("x[1] != 1")
	}
}

func TestNative1(t *testing.T) {
	vm := CreateVM()
	vm.Dict.Value["add"] = &NativeValue{Value: func(pc *PC, values []Value) Value {
		return &IntValue{Value: values[0].(*IntValue).Value + values[1].(*IntValue).Value}
	}}
	code := Parse("native [x y] add")
	vm.Bind(code)
	x := vm.Exec(code)
	params := Parse("1 2")
	pc := newPC(vm, params)
	result := x.(*ProcValue).Value(pc)

	if result.(*IntValue).Value != 3 {
		t.Error("result != 3")
	}
}

func TestGetPath(t *testing.T) {
	vm := CreateVM()
	m := make(map[string]Value)
	m["y"] = &IntValue{Value: 42}
	vm.Dict.Value["x"] = &MapValue{Value: m}
	code := Parse(":x/y")
	vm.Bind(code)
	x := vm.Exec(code)

	if x.(*IntValue).Value != 42 {
		t.Error("result != 42")
	}
}

func TestAdd(t *testing.T) {
	vm := CreateVM()
	LoadCore(vm)
	code := Parse("add 1 2")
	vm.Bind(code)
	x := vm.Exec(code)
	if x.(*IntValue).Value != 3 {
		t.Error("result != 3")
	}
}

func TestAddAdd(t *testing.T) {
	vm := CreateVM()
	LoadCore(vm)
	code := Parse("add add 1 2 3")
	vm.Bind(code)
	x := vm.Exec(code)
	if x.(*IntValue).Value != 6 {
		t.Error("result != 6")
	}
}

func TestFn(t *testing.T) {
	vm := CreateVM()
	LoadCore(vm)
	code := Parse("x: fn [n] [add n 10] x 5")
	vm.Bind(code)
	x := vm.Exec(code)
	if x.(*IntValue).Value != 15 {
		t.Errorf("result != 15, %v", x.(*IntValue).Value)
	}
}

func TestSum(t *testing.T) {
	vm := CreateVM()
	LoadCore(vm)
	code := Parse("sum: fn [n] [either gt n 1 [add n sum sub n 1] [n]] sum 100")
	vm.Bind(code)
	x := vm.Exec(code)
	if x.(*IntValue).Value != 5050 {
		t.Errorf("result != 5050, %v", x.(*IntValue).Value)
	}
}

func BenchmarkFib(t *testing.B) {
	vm := CreateVM()
	LoadCore(vm)
	code := Parse("fib: fn [n] [either gt n 1 [add fib sub n 2 fib sub n 1] [n]] fib 40")
	vm.Bind(code)
	x := vm.Exec(code)
	t.Logf("fib %d", x.(*IntValue).Value)
	// if x.(*IntValue).Value != 5050 {
	// 	t.Errorf("result != 5050, %v", x.(*IntValue).Value)
	// }
}
