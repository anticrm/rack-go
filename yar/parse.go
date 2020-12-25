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
	"strings"
)

type codeStack []Block

func (s *codeStack) push(code Block) {
	*s = append(*s, code)
}

func (s *codeStack) pop() Block {
	index := len(*s) - 1
	element := (*s)[index]
	*s = (*s)[:index]
	return element
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func readIdent(s string, i *int) string {
	var ident strings.Builder
	for *i < len(s) && strings.IndexByte(" \n[](){}:;/", s[*i]) == -1 {
		ident.WriteByte(s[*i])
		*i = (*i + 1)
	}
	return ident.String()
}

func (vm *VM) Parse(s string) Block {
	var stack codeStack
	result := vm.AllocBlock()
	i := 0

	for i < len(s) {
		switch s[i] {
		case ' ', '\n', '\r', '\t':
			i++
		case ']':
			i++
			code := result
			result = stack.pop()
			result.Add(vm, code.Value())
		case '[':
			i++
			stack.push(result)
			result = vm.AllocBlock()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			val := 0
			for i < len(s) && isDigit(s[i]) {
				val = val*10 + int(s[i]-'0')
				i++
			}
			result.Add(vm, MakeInt(val).Value())

		case '"':
			var builder strings.Builder
			i++
			for i < len(s) && s[i] != '"' {
				builder.WriteByte(s[i])
				i++
			}
			result.Add(vm, vm.AllocString(builder.String()).Value())
			i++
			break

		default:
			kind := WordType
			c := s[i]
			if c == '\'' {
				kind = QuoteType
				i++
			} else if c == ':' {
				kind = GetWordType
				i++
			}

			ident := readIdent(s, &i)

			if i < len(s) {
				if s[i] == ':' {
					kind = SetWordType
					i++
				} else if s[i] == '/' {
					var path path

					if kind == GetWordType {
						path = vm.AllocGetPath()
					} else if kind == WordType {
						path = vm.AllocPath()
					} else {
						panic("path not implemented")
					}

					path.Add(vm, vm.GetSymbolID(ident))
					i++
					for i < len(s) {
						ident = readIdent(s, &i)
						path.Add(vm, vm.GetSymbolID(ident))
						if i >= len(s) || s[i] != '/' {
							break
						}
						i++
					}
					result.Add(vm, path.Value())
					break
				}
			}

			switch kind {
			case WordType:
				result.Add(vm, vm.AllocWord(vm.GetSymbolID(ident)).Value())
			case GetWordType:
				result.Add(vm, vm.AllocGetWord(vm.GetSymbolID(ident)).Value())
			case SetWordType:
				result.Add(vm, vm.AllocSetWord(vm.GetSymbolID(ident)).Value())
			default:
				fmt.Printf("kind %v", kind)
				panic("not implemented other words")
			}
		}
	}
	return result
}
