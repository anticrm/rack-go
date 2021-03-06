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
	"testing"

	"github.com/anticrm/rack/yar"
)

func TestClusterInit(t *testing.T) {
	vm := yar.NewVM(1000, 100)
	yar.BootVM(vm)
	vm.Library.Add(clusterPackage())
	clusterModule(vm)
	code := vm.Parse("cluster/init cluster/docker-service \"redis\" \"redis\" print cluster/nodes")
	result := vm.BindAndExec(code)
	t.Logf("result: %016x", result)
	// vm.Dump()
}
