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
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/logger"
	"github.com/lni/goutils/syncutil"
)

const clusterID = uint64(128)

type NodeConfig struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

type ClusterConfig struct {
	Nodes []NodeConfig `yaml:"nodes"`
}

type Cluster struct {
	config *ClusterConfig
}

func NewCluster(config *ClusterConfig) *Cluster {
	return &Cluster{config: config}
}

func (c *Cluster) Start(nodeAddr string) {
	fmt.Printf("cluster nodes:\n")
	var nodeID uint64
	var nodeName string
	initialMembers := make(map[uint64]string)
	if true {
		for i, v := range c.config.Nodes {
			id := uint64(i + 1)
			initialMembers[id] = v.Addr
			if v.Addr == nodeAddr {
				nodeName = v.Name
				nodeID = id
			}
			fmt.Printf(" - %s\n", v)
		}
	}

	if nodeID == 0 {
		log.Fatalf("this node is not a part of cluster (%s)", nodeAddr)
	}

	fmt.Fprintf(os.Stdout, "node name: %s, address: %s\n", nodeName, nodeAddr)

	// change the log verbosity
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)

	// See GoDoc for all available options
	rc := config.Config{
		// ClusterID and NodeID of the raft node
		NodeID:    nodeID,
		ClusterID: clusterID,
		// In this example, we assume the end-to-end round trip time (RTT) between
		// NodeHost instances (on different machines, VMs or containers) are 200
		// millisecond, it is set in the RTTMillisecond field of the
		// config.NodeHostConfig instance below.
		// ElectionRTT is set to 10 in this example, it determines that the node
		// should start an election if there is no heartbeat from the leader for
		// 10 * RTT time intervals.
		ElectionRTT: 10,
		// HeartbeatRTT is set to 1 in this example, it determines that when the
		// node is a leader, it should broadcast heartbeat messages to its followers
		// every such 1 * RTT time interval.
		HeartbeatRTT: 1,
		CheckQuorum:  true,
		// SnapshotEntries determines how often should we take a snapshot of the
		// replicated state machine, it is set to 10 her which means a snapshot
		// will be captured for every 10 applied proposals (writes).
		// In your real world application, it should be set to much higher values
		// You need to determine a suitable value based on how much space you are
		// willing use on Raft Logs, how fast can you capture a snapshot of your
		// replicated state machine, how often such snapshot is going to be used
		// etc.
		SnapshotEntries: 10,
		// Once a snapshot is captured and saved, how many Raft entries already
		// covered by the new snapshot should be kept. This is useful when some
		// followers are just a little bit left behind, with such overhead Raft
		// entries, the leaders can send them regular entries rather than the full
		// snapshot image.
		CompactionOverhead: 5,
	}

	datadir := filepath.Join(
		".rack",
		fmt.Sprintf("node%d", nodeID))
	// config for the nodehost
	// See GoDoc for all available options
	// by default, insecure transport is used, you can choose to use Mutual TLS
	// Authentication to authenticate both servers and clients. To use Mutual
	// TLS Authentication, set the MutualTLS field in NodeHostConfig to true, set
	// the CAFile, CertFile and KeyFile fields to point to the path of your CA
	// file, certificate and key files.
	nhc := config.NodeHostConfig{
		// WALDir is the directory to store the WAL of all Raft Logs. It is
		// recommended to use Enterprise SSDs with good fsync() performance
		// to get the best performance. A few SSDs we tested or known to work very
		// well
		// Recommended SATA SSDs -
		// Intel S3700, Intel S3710, Micron 500DC
		// Other SATA enterprise class SSDs with power loss protection
		// Recommended NVME SSDs -
		// Most enterprise NVME currently available on the market.
		// SSD to avoid -
		// Consumer class SSDs, no matter whether they are SATA or NVME based, as
		// they usually have very poor fsync() performance.
		//
		// You can use the pg_test_fsync tool shipped with PostgreSQL to test the
		// fsync performance of your WAL disk. It is recommended to use SSDs with
		// fsync latency of well below 1 millisecond.
		//
		// Note that this is only for storing the WAL of Raft Logs, it is size is
		// usually pretty small, 64GB per NodeHost is usually more than enough.
		//
		// If you just have one disk in your system, just set WALDir and NodeHostDir
		// to the same location.
		WALDir: datadir,
		// NodeHostDir is where everything else is stored.
		NodeHostDir: datadir,
		// RTTMillisecond is the average round trip time between NodeHosts (usually
		// on two machines/vms), it is in millisecond. Such RTT includes the
		// processing delays caused by NodeHosts, not just the network delay between
		// two NodeHost instances.
		RTTMillisecond: 200,
		// RaftAddress is used to identify the NodeHost instance
		RaftAddress: nodeAddr,
	}

	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	if err := nh.StartCluster(initialMembers, false, NewStateMachine, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}

	cmdChannel := make(chan string)
	raftStopper := syncutil.NewStopper()

	raftStopper.RunWorker(func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(os.Stdout, "synchronizing views...\n")
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				result, err := nh.SyncRead(ctx, clusterID, []byte{})
				cancel()
				if err == nil {
					var count uint64
					count = binary.LittleEndian.Uint64(result.([]byte))
					fmt.Fprintf(os.Stdout, "count: %d\n", count)
				}
			case <-raftStopper.ShouldStop():
				return
			}
		}
	})

	raftStopper.RunWorker(func() {
		cs := nh.GetNoOPSession(clusterID)
		for {
			select {
			case cmd := <-cmdChannel:
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_, err := nh.SyncPropose(ctx, cs, []byte(cmd))
				cancel()
				if err != nil {
					fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
				}
			case <-raftStopper.ShouldStop():
				return
			}
		}
	})

	startHostMonitor(nodeID, nodeName, cmdChannel)
	startCtl(cmdChannel)

	raftStopper.Wait()
}
