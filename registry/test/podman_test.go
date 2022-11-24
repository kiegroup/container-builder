/*
Copyright 2022.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryWithPodman(t *testing.T) {
	connectionLocal, err := GetPodmanConnection()
	if err != nil {
		fmt.Errorf("%s \n", err)
		assert.FailNow(t, "Connection refused")
	}
	p := Podman{connection: connectionLocal}
	defer p.RemoveRegistryContainerAndImage()
	assert.Truef(t, p.StartRegistry(), "Registry not started")
	assert.Truef(t, p.IsRegistryImagePresent(), "Registry image not present")
	assert.Truef(t, p.IsRegistryRunning(), "Registry container not running")
	assert.Truef(t, p.StopRegistry(), "Registry not stopped")

}
