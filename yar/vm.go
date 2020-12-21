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

const (
	IntKind     = iota
	BoolKind    = iota
	WordKind    = iota
	GetWordKind = iota
	SetWordKind = iota
	Quote       = iota
	ProcKind    = iota
	NativeKind  = iota
	BlockKind   = iota
	MapKind     = iota
	PathKind    = iota
	GetPathKind = iota
)

type bound interface {
	get(sym string) Value
	set(sym string, value Value)
}

type bindFactory func(sym string) bound

type Value interface {
	Kind() int
	bind(factory bindFactory)
	exec(pc *PC) Value
}

type Code []Value

func bind(code Code, factory bindFactory) {
	for _, item := range code {
		item.bind(factory)
	}
}

// W O R D S

type Word struct {
	bound bound
	sym   string
}

func (w *Word) Kind() int { return WordKind }

func (w *Word) bind(factory bindFactory) {
	bound := factory(w.sym)
	if bound != nil {
		w.bound = bound
	}
}

func (w *Word) exec(pc *PC) Value {
	if w.bound == nil {
		panic("word not bound")
	}
	value := w.bound.get(w.sym)
	if value == nil {
		panic("nothing when read")
	}
	return value.exec(pc)
}

type SetWord struct {
	bound bound
	sym   string
}

func (w *SetWord) Kind() int { return SetWordKind }

func (w *SetWord) bind(factory bindFactory) {
	bound := factory(w.sym)
	if bound != nil {
		w.bound = bound
	}
}

func (w *SetWord) exec(pc *PC) Value {
	if w.bound == nil {
		panic("word not bound")
	}
	result := pc.next()
	w.bound.set(w.sym, result)
	return result
}

// T Y P E S

type IntValue struct {
	Value int
}

func (i *IntValue) Kind() int                { return IntKind }
func (i *IntValue) bind(factory bindFactory) {}
func (i *IntValue) exec(pc *PC) Value        { return i }

type BoolValue struct {
	Value bool
}

func (i *BoolValue) Kind() int                { return BoolKind }
func (i *BoolValue) bind(factory bindFactory) {}
func (i *BoolValue) exec(pc *PC) Value        { return i }

type ProcValue struct {
	Value func(*PC) Value
}

func (i *ProcValue) Kind() int                { return ProcKind }
func (i *ProcValue) bind(factory bindFactory) {}
func (i *ProcValue) exec(pc *PC) Value        { return i.Value(pc) }

type NativeValue struct {
	Value func(*PC, []Value) Value
}

func (i *NativeValue) Kind() int                { return NativeKind }
func (i *NativeValue) bind(factory bindFactory) {}
func (i *NativeValue) exec(pc *PC) Value        { return i }

type BlockValue struct {
	Value Code
}

func (i *BlockValue) Kind() int                { return BlockKind }
func (i *BlockValue) bind(factory bindFactory) { bind(i.Value, factory) }
func (i *BlockValue) exec(pc *PC) Value        { return i }

type MapValue struct {
	Value map[string]Value
}

func (i *MapValue) Kind() int                { return MapKind }
func (i *MapValue) bind(factory bindFactory) {}
func (i *MapValue) exec(pc *PC) Value        { return i }

func (i *MapValue) get(sym string) Value        { return i.Value[sym] }
func (i *MapValue) set(sym string, value Value) { i.Value[sym] = value }

type GetPathValue struct {
	bound bound
	Path  []string
}

func (i *GetPathValue) Kind() int { return GetPathKind }
func (i *GetPathValue) bind(factory bindFactory) {
	bound := factory(i.Path[0])
	if bound != nil {
		i.bound = bound
	}
}
func (path *GetPathValue) exec(pc *PC) Value {
	if path.bound == nil {
		panic("path not bound")
	}
	val := path.bound.get(path.Path[0])
	for i := 1; i < len(path.Path); i++ {
		val = val.(bound).get(path.Path[i])
	}
	return val
}

// V M

type VM struct {
	result Value
	Dict   MapValue
	stack  []Value
}

func (vm *VM) Exec(code Code) Value {
	return newPC(vm, code).exec()
}

func (vm *VM) Bind(code Code) {
	bind(code, func(sym string) bound {
		return &vm.Dict
	})
}

type PC struct {
	code Code
	pc   int
	vm   *VM
}

func newPC(vm *VM, code Code) *PC {
	return &PC{code: code, vm: vm, pc: 0}
}

func (pc *PC) nextNoInfix() Value {
	item := pc.code[pc.pc]
	pc.pc++
	result := item.exec(pc)
	pc.vm.result = result
	return result
}

func (pc *PC) next() Value {
	result := pc.nextNoInfix()
	return result
}

func (pc *PC) exec() Value {
	var result Value
	for pc.pc < len(pc.code) {
		result = pc.next()
	}
	return result
}
