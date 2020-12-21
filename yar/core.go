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

func add(pc *PC, values []Value) Value {
	return &IntValue{Value: values[0].(*IntValue).Value + values[1].(*IntValue).Value}
}

func sub(pc *PC, values []Value) Value {
	return &IntValue{Value: values[0].(*IntValue).Value - values[1].(*IntValue).Value}
}

func gt(pc *PC, values []Value) Value {
	return &BoolValue{Value: values[0].(*IntValue).Value > values[1].(*IntValue).Value}
}

type stackFrame struct {
	vm      *VM
	offsets map[string]int
}

func (s *stackFrame) get(sym string) Value {
	return s.vm.stack[len(s.vm.stack)+s.offsets[sym]]
}

func (s *stackFrame) set(sym string, value Value) {
	s.vm.stack[len(s.vm.stack)+s.offsets[sym]] = value
}

func fn(pc *PC, values []Value) Value {
	params := values[0].(*BlockValue)
	code := values[1].(*BlockValue)
	ref := blockOfRefinements(params.Value)
	if ref["local"] == nil {
		ref["local"] = make([]Value, 0)
	}

	defaults := len(ref["default"])
	locals := len(ref["local"])
	alternatives := make([]string, 0)
	for k := range ref {
		if k != "default" && k != "local" {
			alternatives = append(alternatives, k)
		}
	}
	stackSize := defaults + locals + len(alternatives)
	offsets := make(map[string]int)
	for i, item := range ref["default"] {
		offsets[item.(*Word).sym] = i - stackSize
	}
	for i, item := range ref["local"] {
		offsets[item.(*Word).sym] = i + defaults - stackSize
	}
	for i, item := range alternatives {
		offsets[item] = i + defaults + locals - stackSize
		len := len(ref[item])
		for i, val := range ref[item] {
			offsets[val.(*Word).sym] = i - stackSize - len
		}
	}

	frame := &stackFrame{vm: pc.vm, offsets: offsets}

	bind(code.Value, func(sym string) bound {
		if _, ok := offsets[sym]; ok {
			return frame
		}
		return nil
	})

	create := func(alt int) *ProcValue {
		// altStackSize := 0

		return &ProcValue{Value: func(pc *PC) Value {
			altBase := len(pc.vm.stack)
			// base := altBase + altStackSize

			for i := 0; i < defaults; i++ {
				pc.vm.stack = append(pc.vm.stack, pc.next())
			}
			for i := 0; i < locals; i++ {
				pc.vm.stack = append(pc.vm.stack, &IntValue{Value: 0})
			}
			result := pc.vm.Exec(code.Value)
			pc.vm.stack = pc.vm.stack[:altBase]
			return result
		}}
	}

	return create(-1)
}

func either(pc *PC, values []Value) Value {
	cond := values[0].(*BoolValue)
	if cond.Value {
		return pc.vm.Exec(values[1].(*BlockValue).Value)
	}
	return pc.vm.Exec(values[2].(*BlockValue).Value)
}

const y = `
add: native [x y] :core/add
sub: native [x y] :core/sub

gt: native [x y] :core/gt

fn: native [params code] :core/fn

either: native [cond ifTrue ifFalse] :core/either
`

func LoadCore(vm *VM) {
	natives := &MapValue{Value: make(map[string]Value)}
	natives.Value["add"] = &NativeValue{Value: add}
	natives.Value["sub"] = &NativeValue{Value: sub}
	natives.Value["gt"] = &NativeValue{Value: gt}
	natives.Value["fn"] = &NativeValue{Value: fn}
	natives.Value["either"] = &NativeValue{Value: either}

	vm.Dict.Value["core"] = natives
	code := Parse(y)
	vm.Bind(code)
	vm.Exec(code)
}
