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
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/anticrm/rack/yar"
	sm "github.com/lni/dragonboat/v3/statemachine"
)

// StateMachine is the IStateMachine implementation used
type StateMachine struct {
	ClusterID uint64
	NodeID    uint64
	VM        *yar.VM
}

func NewStateMachine(clusterID uint64, nodeID uint64) sm.IStateMachine {
	sm := &StateMachine{
		ClusterID: clusterID,
		NodeID:    nodeID,
		VM:        yar.NewVM(65536, 100),
	}
	yar.BootVM(sm.VM)
	sm.VM.Library.Add(clusterPackage())
	clusterModule(sm.VM)
	return sm
}

// Lookup performs local lookup on the StateMachine instance. In this example,
// we always return the Count value as a little endian binary encoded byte
// slice.
func (s *StateMachine) Lookup(query interface{}) (interface{}, error) {
	result := make([]byte, 8)
	binary.LittleEndian.PutUint64(result, 0)
	return result, nil
}

// Update updates the object using the specified committed raft entry.
func (s *StateMachine) Update(data []byte) (sm.Result, error) {
	fmt.Printf("NodeID: %04x\n", s.NodeID)
	fmt.Printf("> %s\n", string(data))
	code := s.VM.Parse(string(data))
	result := s.VM.BindAndExec(code)
	fmt.Printf("%s\n", s.VM.ToString(result))
	return sm.Result{Value: uint64(len(data))}, nil
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *StateMachine) SaveSnapshot(w io.Writer,
	fc sm.ISnapshotFileCollection, done <-chan struct{}) error {
	// as shown above, the only state that can be saved is the Count variable
	// there is no external file in this IStateMachine example, we thus leave
	// the fc untouched
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0)
	_, err := w.Write(data)
	return err
}

// RecoverFromSnapshot recovers the state using the provided snapshot.
func (s *StateMachine) RecoverFromSnapshot(r io.Reader,
	files []sm.SnapshotFile,
	done <-chan struct{}) error {
	// restore the Count variable, that is the only state we maintain in this
	// example, the input files is expected to be empty
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	binary.LittleEndian.Uint64(data)
	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *StateMachine) Close() error { return nil }

// GetHash returns a uint64 representing the current object state.
func (s *StateMachine) GetHash() (uint64, error) {
	// the only state we have is that Count variable. that uint64 value pretty much
	// represents the state of this IStateMachine
	return 0, nil
}
