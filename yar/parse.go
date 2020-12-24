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

type codeStack []pBlock

func (s *codeStack) push(code pBlock) {
	*s = append(*s, code)
}

func (s *codeStack) pop() pBlock {
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

func (vm *VM) Parse(s string) pBlock {
	var stack codeStack
	result := vm.allocBlock()
	i := 0

	for i < len(s) {
		switch s[i] {
		case ' ', '\n':
			i++
		case ']':
			i++
			code := result
			result = stack.pop()
			result.add(vm, ptr(code))
		case '[':
			i++
			stack.push(result)
			result = vm.allocBlock()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			val := 0
			for i < len(s) && isDigit(s[i]) {
				val = val*10 + int(s[i]-'0')
				i++
			}
			result.add(vm, vm.alloc(cell(makeInt(val))))
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
					builder := pathBuilder{vm: vm}
					builder.add(vm.getSymbolID(ident))
					i++
					for i < len(s) {
						ident = readIdent(s, &i)
						builder.add(vm.getSymbolID(ident))
						if i >= len(s) || s[i] != '/' {
							break
						}
						i++
					}
					if kind == GetWordType {
						result.add(vm, ptr(builder.first))
					} else {
						panic("path not implemented")
					}
					break
				}
			}

			switch kind {
			case WordType:
				result.add(vm, vm.alloc(cell(vm.makeWord(ident))))
			case GetWordType:
				result.add(vm, vm.alloc(cell(vm.makeGetWord(ident))))
			case SetWordType:
				result.add(vm, vm.alloc(cell(vm.makeSetWord(ident))))
			default:
				fmt.Printf("kind %v", kind)
				panic("not implemented other words")
			}
		}
	}
	return result
}
