// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

var (
	projectID   string
	accessToken string
	ctx         = context.Background()
)

type TestResource struct {
	Name     string
	Zone     string
	DiskName string
}

var testResources = []TestResource{
	TestResource{"test-resource-1", "us-east1-b", "test-disk-2"},
	TestResource{"test-resource-2", "us-east1-b", "test-disk-3"},
	TestResource{"test-resource-3", "us-east1-c", "test-disk-2"},
	TestResource{"test-skip", "us-east1-c", "test-disk-3"},
	TestResource{"another-resource-1", "us-east1-b", "test-disk-4"},
	TestResource{"another-resource-2", "us-east1-b", "test-disk-5"},
}

// TestReaperIntegration creates test instances in GCP, and runs a reaper with a config to test functionality.
func TestReaperIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reaper integration test in short mode")
	}
	err := setup(true)
	if err != nil {
		t.Error(err)
	}
	resources := []*reaperconfig.ResourceConfig{
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b", "us-east1-c"}, "test", "skip", "9 7 * * *"),
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "another", "", "1 * * * *"),
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "another-resource-1", "", "* * * 10 *"),
	}
	reaperConfig := reaper.NewReaperConfig(resources, "TestSchedule", projectID, "UUID")

	reaper := reaper.NewReaper()
	err = reaper.UpdateReaperConfig(reaperConfig)
	if err != nil {
		t.Error(err)
	}
	reaper.GetResources(ctx)

	var expectedWatchedResources = []string{"test-resource-1", "test-resource-2", "test-resource-3", "another-resource-1", "another-resource-2"}
	for _, expectedResource := range expectedWatchedResources {
		resourceIdx := -1
		for idx, watchedResource := range reaper.Watchlist {
			if strings.Compare(watchedResource.Name, expectedResource) == 0 {
				resourceIdx = idx
				break
			}
		}
		if resourceIdx == -1 {
			t.Errorf("Resource %s is not watched by Reaper", expectedResource)
		}
	}

	reaper.FreezeTime(time.Now().AddDate(0, 1, 0))
	reaper.SweepThroughResources(ctx)

	expectedResource := "another-resource-1"
	for _, watchedResource := range reaper.Watchlist {
		if strings.Compare(watchedResource.Name, expectedResource) == 0 {
			return
		}
	}
	t.Errorf("Resource %s not in watchlist", expectedResource)
}

func setup(shouldCreateResources bool) error {
	var err error
	projectID, accessToken, err = ReadConfigFile()
	if err != nil {
		return err
	}
	if shouldCreateResources {
		createTestResources()
	}
	return nil
}

func createTestResources() {
	for _, resource := range testResources {
		createGCEInstance(ctx, resource.Name, resource.Zone, resource.DiskName)
	}
}

func createGCEInstance(ctx context.Context, name, zone, diskName string) {
	endpoint := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances", projectID, zone)

	reqBody := struct {
		MachineType       string `json:"machineType"`
		Name              string `json:"name"`
		NetworkInterfaces []struct {
			Network string `json:"network"`
		} `json:"networkInterfaces"`
		Disks []struct {
			Boot             bool `json:"boot"`
			AutoDelete       bool `json:"autoDelete"`
			InitializeParams struct {
				DiskName    string `json:"diskName"`
				SourceImage string `json:"sourceImage"`
			} `json:"initializeParams"`
			Mode      string `json:"mode"`
			Interface string `json:"interface"`
		} `json:"disks"`
	}{
		Name:        name,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/f1-micro", zone),

		NetworkInterfaces: []struct {
			Network string `json:"network"`
		}{
			{
				Network: fmt.Sprintf("projects/%s/global/networks/default", projectID),
			},
		}, //For simplicity use the default network
		Disks: []struct {
			Boot             bool `json:"boot"`
			AutoDelete       bool `json:"autoDelete"`
			InitializeParams struct {
				DiskName    string `json:"diskName"`
				SourceImage string `json:"sourceImage"`
			} `json:"initializeParams"`
			Mode      string `json:"mode"`
			Interface string `json:"interface"`
		}{
			{
				Boot:       true,
				AutoDelete: false,
				Mode:       "READ_WRITE",
				Interface:  "SCSI",
				InitializeParams: struct {
					DiskName    string `json:"diskName"`
					SourceImage string `json:"sourceImage"`
				}{
					DiskName:    diskName,
					SourceImage: "projects/debian-cloud/global/images/family/debian-9",
				},
			},
		},
	}

	bodyData, err := json.Marshal(reqBody)
	if err != nil {
		log.Println(err.Error())
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyData)))
	if err != nil {
		log.Println(err.Error())
	}
	request.Header.Set(http.CanonicalHeaderKey("authorization"), fmt.Sprintf("Bearer %s", accessToken))
	request.Header.Set(http.CanonicalHeaderKey("content-type"), "application/json")

	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		log.Println(err.Error())
	}

	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Println(string(data))
}
