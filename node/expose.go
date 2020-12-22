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

package node

import (
	"net/http"

	"github.com/anticrm/rack/yar"
)

func expose(pc *yar.PC) yar.Value {
	vm := pc.VM
	fn := pc.Next()
	params := yar.Block(pc.Next())

	var extractors []func(r *http.Request) yar.Value

	for i := params.First(); i != 0; i = i.Next(vm) {
		value := i.Value(vm)
		switch value.Kind() {
		case yar.WordType:
			word := yar.Word(value)
			symbol := vm.InverseSymbols[word.Sym()]
			extractor := func(r *http.Request) yar.Value {
				val := r.URL.Query()[symbol]
				return yar.Value(len(val))
			}
			extractors = append(extractors, extractor)
			clone := vm.Clone()
			handler := func(w http.ResponseWriter, r *http.Request) {
				stack := make([]yar.Value, 100)
				for i, e := range extractors {
					stack[i] = e(r)
				}
				fork := clone.Fork(stack, len(extractors))
			}
		default:
			panic("unsupported kind")
		}
	}
}
