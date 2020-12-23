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
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/anticrm/rack/node"
	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Print("anticrm rack (c) copyright 2020, 2021 Anticrm Project Contributors. All rights reserved.\n")
	configFile := flag.String("f", "rack.yml", "Config file")
	nodeAddr := flag.String("addr", "", "This node address")
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Can't read config file #%v ", err)
	}

	conf := &node.ClusterConfig{}

	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	cluster := node.NewCluster(conf)
	cluster.Start(*nodeAddr)
}
