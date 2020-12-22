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
	"log"
	"net/http"
)

type Deployment struct {
	Domain string
}

type Node struct {
}

type exposedFn struct{}

func NewNode() *Node {
	mux := http.NewServeMux()
	mux.Handle("/", &exposedFn{})
	mux.HandleFunc("/posts", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("Visit http://bit.ly/just-enough-go to get started"))
	})
	server := http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())

	return &Node{}
}

func (h *exposedFn) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Welcome to the \"Just Enough Go\" blog series!!"))
}

func (node *Node) deploy(deployment *Deployment) {

}
