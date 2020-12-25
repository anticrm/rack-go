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
	nodes := vm.AllocBlock()
	nodes.Add(vm, vm.AllocString("localhost:63001"))
	vm.Dictionary.Put(vm, vm.GetSymbolID("nodes"), nodes.Value())
	return 0
}

func clusterPackage() *yar.Pkg {
	result := yar.NewPackage("cluster")
	result.AddFunc("init", clusterInit)
	return result
}

func loadClusterPackage(vm *yar.VM) {
	mod := vm.AllocDict()
	vm.LoadPackage(clusterPackage(), mod)
	vm.Dictionary.Put(vm, vm.GetSymbolID("cluster"), mod.Value())
}
