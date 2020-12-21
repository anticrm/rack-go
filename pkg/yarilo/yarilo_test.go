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

package yarilo

import (
	"testing"
)

func TestParse1(t *testing.T) {
	x := Parse("add 1 2")
	if (x[1] != Integer{Val: 1}) {
		t.Errorf("x[1] != 1")
	}
}

func TestNative1(t *testing.T) {
	vm := CreateVM()
	vm.Dict["add"] = NativeFunc{F: func(pc *PC, values []Value) Value {
		return Integer{Val: values[0].(Integer).Val + values[1].(Integer).Val}
	}}
	code := Parse("native [x y] add")
	vm.Bind(code)
	x := vm.Exec(code).(Proc)
	params := Parse("1 2")
	pc := NewPC(vm, params)
	result := x.F(pc)

	t.Logf("%+v", result)
}
