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
	"net"
	"net/url"
	"strconv"

	"github.com/anticrm/rack/docker"
	"github.com/anticrm/rack/http"
)

func GetFreeAddr() (*net.TCPAddr, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr), nil
}

func main() {
	fmt.Print("rack node (c) 2020 anticrm folks.\n")

	server := http.NewServer()

	for i := 0; i < 4; i++ {
		addr, err := GetFreeAddr()
		if err != nil {
			panic(err)
		}
		go docker.Run("anticrm/scrn:5", addr.Port)

		url, err := url.Parse("http://localhost:" + strconv.Itoa(addr.Port))
		if err != nil {
			panic(err)
		}

		local := http.NewBackend(url)
		server.AddBackend(local)
	}

	server.Start()
}
