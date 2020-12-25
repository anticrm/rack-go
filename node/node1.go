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

type Deployment struct {
	Domain string
}

type Node1 struct {
	vm *yar.VM
}

type HttpService struct {
	mux *http.ServeMux
}

func NewNode1() *Node1 {

	vm := yar.NewVM(1000, 100)
	yar.BootVM(vm)
	// vm.AddNative("expose", expose)

	service := &HttpService{mux: http.NewServeMux()}
	vm.Services["http"] = service

	code := vm.Parse("calc: fn [x y] [add x y] expose :calc [x y]")
	vm.BindAndExec(code)

	// mux.Handle("/", &exposedFn{})
	// mux.HandleFunc("/posts", func(rw http.ResponseWriter, req *http.Request) {
	// 	rw.Write([]byte("Visit http://bit.ly/just-enough-go to get started"))
	// })
	server := http.Server{Addr: ":8080", Handler: service.mux}
	server.ListenAndServe()

	return &Node1{vm: vm}
}

// func (h *exposedFn) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
// 	rw.Write([]byte("Welcome to the \"Just Enough Go\" blog series!!"))
// }

func (node *Node1) deploy(deployment *Deployment) {

}
