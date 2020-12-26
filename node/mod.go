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
	services: []
	init: fn [] [
		append nodes make-object [addr: "localhost:63001" cpus: 2 docker-procs: []]
		append nodes make-object [addr: "localhost:63002" cpus: 2 docker-procs: []]
	]
	docker-service: fn [_image _port] [
		print image
		foreach node nodes [
			repeat cpu node/cpus [
				append node/docker-procs make-object [image: _image port: _port]
			]
		]
	]
	node-info: fn [nodeID nodeName cores cpuModelName /local node] [
		node: get in nodes nodeName
		if unset? node [node: set in nodes nodeName make-object []]
		node/cores: cores
		node/cpuModelName: cpuModelName
	]
]
`

func clusterModule(vm *yar.VM) yar.Value {
	code := vm.Parse(clusterY)
	return vm.BindAndExec(code)
}
