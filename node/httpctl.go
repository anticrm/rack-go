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
	"fmt"
	"net/http"
)

type controlHandler struct {
	cmd chan string
}

func startCtl(cmd chan string) {
	mux := http.NewServeMux()
	mux.Handle("/do", &controlHandler{cmd: cmd})
	server := http.Server{Addr: ":8080", Handler: mux}
	go server.ListenAndServe()
}

func (h *controlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	fmt.Printf("running command: %s\n", cmd)
	h.cmd <- cmd
	fmt.Fprintln(w, "Kewl!")
}
