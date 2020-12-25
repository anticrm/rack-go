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

import "github.com/anticrm/rack/yar"

func clusterInit(vm *yar.VM) yar.Value {
	// nodes := vm.Dictionary.Find(vm, vm.GetSymbolID("cluster"))
	// nodes.Add(vm, vm.AllocString("localhost:63001").Value())
	// vm.Dictionary.Put(vm, vm.GetSymbolID("nodes"), nodes.Value())
	return 0
}

func clusterPackage() *yar.Pkg {
	result := yar.NewPackage("cluster")
	result.AddFunc("init", clusterInit)
	return result
}

const clusterY = `
cluster: make-object [
	nodes: []
	init: fn [] [append nodes 1]
	docker-service: fn [] [print nodes]
]
`

func clusterModule(vm *yar.VM) yar.Value {
	code := vm.Parse(clusterY)
	return vm.BindAndExec(code)
}
