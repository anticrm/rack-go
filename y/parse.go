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

package y

import (
	"fmt"
	"strings"
)

type codeStack []Code

func (s *codeStack) push(code Code) {
	*s = append(*s, code)
}

func (s *codeStack) pop() Code {
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

func (vm *VM) Parse(s string) Code {
	var stack codeStack
	result := vm.alloc(newBlock(0, 0))
	i := 0

	for i < len(s) {
		switch s[i] {
		case ' ', '\n':
			i++
		case ']':
			i++
			code := result
			result = stack.pop()
			blockAddPtr(vm, result, code)
		case '[':
			i++
			stack.push(result)
			result = vm.alloc(newBlock(0, 0))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			val := 0
			for i < len(s) && isDigit(s[i]) {
				val = val*10 + int(s[i]-'0')
				i++
			}
			blockAdd(vm, result, newInteger(val))
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
					// var path []string
					// path = append(path, ident)
					// i++
					// for i < len(s) {
					// 	ident := readIdent(s, &i)
					// 	path = append(path, ident)
					// 	if i >= len(s) || s[i] != '/' {
					// 		break
					// 	}
					// 	i++
					// }
					// if kind == GetWordKind {
					// 	result = append(result, &GetPathValue{Path: path})
					// } else {
					// 	panic("path not implemented")
					// }
					// break
					panic("not implemented")
				}
			}

			switch kind {
			case WordType:
				blockAdd(vm, result, newWord(vm, ident))
			case SetWordType:
				blockAdd(vm, result, newSetWord(vm, ident))
			default:
				fmt.Printf("kind %v", kind)
				panic("not implemented other words")
			}
		}
	}
	return result
}
