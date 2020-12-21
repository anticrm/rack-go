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

const (
	IntValue    = iota
	WordValue   = iota
	ProcValue   = iota
	BlockValue  = iota
	NativeValue = iota
)

type Value interface {
	kind() int
}

type YFunc func(pc *PC) Value

type Proc struct {
	F YFunc
}

func (p Proc) kind() int {
	return ProcValue
}

const (
	Norm    = iota
	GetWord = iota
	SetWord = iota
	Quote   = iota
)

type Bound interface {
	get(sym string) Value
	set(sym string, value Value)
}

type BindFactory func(sym string) Bound

type CodeItem interface {
	Value
	bind(factory BindFactory)
	exec(pc *PC) Value
}

type Code []CodeItem

func Bind(code Code, factory BindFactory) {
	for _, item := range code {
		item.bind(factory)
	}
}

func checkReturn(result Value, pc *PC) Value {
	if result.kind() == ProcValue {
		return result.(Proc).F(pc)
	}
	return result
}

type Word struct {
	_kind int
	bound Bound
	sym   string
	infix bool
}

func (w *Word) kind() int { return WordValue }

func (w *Word) bind(factory BindFactory) {
	bound := factory(w.sym)
	if bound != nil {
		w.bound = bound
	}
}

func (w *Word) exec(pc *PC) Value {
	if w.bound == nil {
		panic("word not bound")
	}
	switch w._kind {
	case Norm:
		f := w.bound.get(w.sym)
		if f == nil {
			panic("nothing when read")
		}
		return checkReturn(f, pc)
	case SetWord:
		x := pc.next()
		w.bound.set(w.sym, x)
		return x
	case GetWord:
		return w.bound.get(w.sym)
	default:
		panic("not implemented")
	}
}

type Path struct {
	kind  int
	bound Bound
	path  []string
}

type Brackets struct {
	code Code
}

type Integer struct {
	Val int
}

func (c Integer) kind() int { return IntValue }

func (c Integer) bind(factory BindFactory) {
}

func (c Integer) exec(pc *PC) Value {
	return c
}

type Block struct {
	code Code
}

func (b *Block) kind() int { return BlockValue }

func (b *Block) bind(factory BindFactory) {
	Bind(b.code, factory)
}

func (b *Block) exec(pc *PC) Value {
	return b
}

type dictType map[string]Value

type VM struct {
	result Value
	Dict   dictType
}

func (dict dictType) get(sym string) Value {
	return dict[sym]
}

func (dict dictType) set(sym string, value Value) {
	dict[sym] = value
}

func (vm *VM) Exec(code Code) Value {
	return NewPC(vm, code).exec()
}

func (vm *VM) Bind(code Code) {
	Bind(code, func(sym string) Bound {
		return vm.Dict
	})
}

type PC struct {
	code Code
	pc   int
	vm   *VM
}

func NewPC(vm *VM, code Code) *PC {
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
