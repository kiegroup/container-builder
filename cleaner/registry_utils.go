/*
 * Copyright 2022 Red Hat, Inc. and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cleaner

import (
	"fmt"
	"net"
	"os"
	"time"

	registryContainer "github.com/heroku/docker-registry-client/registry"
	"github.com/sirupsen/logrus"
)

func IsPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on port %q: %s", port, err)
		return false
	}
	ln.Close()
	return true
}

func GetRegistryConnection(url string, username string, password string) (*registryContainer.Registry, error) {
	registryConn, err := registryContainer.New(url, username, password)
	if err != nil {
		logrus.Error(err, "First Attempt to connect with RegistryContainer")
	}
	// we try ten times if the machine is slow and the registry needs time to start
	if err != nil {
		logrus.Info("Waiting for a correct ping with RegistryContainer")

		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			if registryConn == nil {
				registryConn, _ = registryContainer.New(url, username, password)
			}
			if registryConn != nil {
				if err := registryConn.Ping(); err != nil {
					continue
				}
			}
		}
	}
	return registryConn, err
}

func GetRegistryContainer() (RegistryContainer, error) {
	registryContainerConnection, err := GetRegistryConnection(REGISTRY_CONTAINER_URL, "", "")
	if err != nil {
		logrus.Errorf("Can't connect to the RegistryContainer")
		return RegistryContainer{}, err
	}
	return RegistryContainer{Connection: *registryContainerConnection}, nil
}
