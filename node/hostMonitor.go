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
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func startHostMonitor(nodeID uint64, nodeName string, cmd chan string) chan bool {

	ticker := time.NewTicker(10 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				fmt.Println("Gathering Host Information...")
				cpuInfo, err := cpu.Info()
				if err != nil {
					fmt.Printf("%v\n", err)
				}
				sendCommand(cmd, []string{"cluster/node-info",
					strconv.Itoa(int(nodeID)), quote(nodeName), strconv.Itoa(int(cpuInfo[0].Cores)), quote(cpuInfo[0].ModelName)})
			}
		}
	}()

	return done
}
