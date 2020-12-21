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

package main

import (
	"fmt"
	"os"

	"github.com/anticrm/rack/pkg/yarilo"
)

func main() {
	input := os.Args[1]

	vm := yarilo.CreateVM()
	vm.Dict["add"] = yarilo.NativeFunc{F: func(pc *yarilo.PC, values []yarilo.Value) yarilo.Value {
		return yarilo.Integer{Val: values[0].(yarilo.Integer).Val + values[1].(yarilo.Integer).Val}
	}}
	code := yarilo.Parse(input)
	vm.Bind(code)
	x := vm.Exec(code).(yarilo.Proc)
	params := yarilo.Parse("1 2")
	pc := yarilo.NewPC(vm, params)
	result := x.F(pc)

	fmt.Printf("%+v", result)
}
