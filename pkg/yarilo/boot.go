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

type NativeFunc struct {
	F func(*PC, []Value) Value
}

func (nf NativeFunc) kind() int {
	return NativeValue
}

func blockOfRefinements(code Code) map[string]Code {
	result := make(map[string]Code)
	current := "default"
	result[current] = make(Code, 0)
	for _, item := range code {
		result[current] = append(result[current], item)
	}
	return result
}

func native(pc *PC) Value {
	params := pc.next()
	impl := pc.next().(NativeFunc)

	ref := blockOfRefinements(params.(*Block).code)

	defaults := len(ref["default"])
	var alternatives []string

	for k := range ref {
		if k != "default" {
			alternatives = append(alternatives, k)
		}
	}

	create := func(alt int) Proc {
		altStackSize := 0
		f := impl
		if alt >= 0 {
			altStackSize = len(ref[alternatives[alt]])
		}

		yfunc := func(pc *PC) Value {
			var values []Value
			for i := 0; i < defaults; i++ {
				values = append(values, pc.next())
			}
			for i := 0; i < altStackSize; i++ {
				values = append(values, pc.next())
			}
			return f.F(pc, values)
		}

		return Proc{F: yfunc}
	}

	if len(alternatives) == 0 {
		return create(-1)
	}

	panic("not implemented")
}

func CreateVM() *VM {
	vm := &VM{Dict: make(map[string]Value)}
	vm.Dict["native"] = Proc{F: native}
	return vm
}
