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

package manager

import (
	"strings"
	"testing"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

// type OperationType int

// const (
// 	Add    OperationType = 0
// 	Delete OperationType = 1
// 	Quit   OperationType = 2
// 	Run    OperationType = 3
// )

// type MonitorReapersTestCase struct {
// 	Type   OperationType
// 	Reaper *reaper.Reaper
// }

// var monitorReapersTestCases = []MonitorReapersTestCase{
// 	MonitorReapersTestCase{Add, createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_1"))},
// 	MonitorReapersTestCase{Delete, nil},
// 	MonitorReapersTestCase{Quit, nil},
// 	MonitorReapersTestCase{Run, nil},
// }

// func TestMonitorReapers(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)

// 	wg.Done()
// }

type HandleDeleteTestCase struct {
	UUID            string
	ExpectedReapers []*reaper.Reaper
	Expected        bool
}

var handleDeleteTestCases = []HandleDeleteTestCase{
	HandleDeleteTestCase{"UUID_4", append(getTestReapers()[:3], getTestReapers()[4]), true},
	HandleDeleteTestCase{"UUID_1", getTestReapers()[1:], true},
	HandleDeleteTestCase{"UUID_3", append(getTestReapers()[:2], getTestReapers()[3:]...), true},
	HandleDeleteTestCase{"UUID_6", getTestReapers(), false},
	HandleDeleteTestCase{"", getTestReapers(), false},
}

func TestHandleDeleteReaper(t *testing.T) {
	testManager := &ReaperManager{}
	for _, testCase := range handleDeleteTestCases {
		testManager.Reapers = getTestReapers()
		result := testManager.handleDeleteReaper(testCase.UUID)
		if result != testCase.Expected {
			t.Errorf("Error in handleDelete: expected %v, got %v", testCase.Expected, result)
		}
		if !areReaperListsEqual(testCase.ExpectedReapers, testManager.Reapers) {
			t.Error("Reaper deletion not handled correctly my manager")
		}
	}
}

type GetReaperTestCase struct {
	UUID     string
	Expected *reaper.Reaper
}

var getReaperTestCases = []GetReaperTestCase{
	GetReaperTestCase{"UUID_1", createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_1"))},
	GetReaperTestCase{"UUID_2", createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_2"))},
	GetReaperTestCase{"UUID_5", createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_5"))},
	GetReaperTestCase{"UUID_7", nil},
	GetReaperTestCase{"", nil},
}

func TestGetReaper(t *testing.T) {
	testManager := &ReaperManager{}
	testManager.Reapers = getTestReapers()
	for _, testCase := range getReaperTestCases {
		result := testManager.GetReaper(testCase.UUID)
		if result == nil && testCase.Expected != nil {
			t.Errorf("GetReaper could not find reaper that exists with UUID %s", testCase.UUID)
		}
		if result != nil && testCase.Expected == nil {
			t.Errorf("GetReaper found reaper that does not exist with UUID %s", testCase.UUID)
		}
		if result != nil && testCase.Expected != nil && result.UUID != testCase.Expected.UUID {
			t.Errorf("Expected UUID %s, found UUID %s", testCase.Expected.UUID, testCase.UUID)
		}
	}
}

func getTestReapers() []*reaper.Reaper {
	return []*reaper.Reaper{
		createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_1")),
		createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_2")),
		createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_3")),
		createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_4")),
		createTestReaper(reaper.NewReaperConfig(nil, "* * * * *", "testProject", "UUID_5")),
	}
}

func createTestReaper(config *reaperconfig.ReaperConfig) *reaper.Reaper {
	reaper := reaper.NewReaper()
	reaper.UpdateReaperConfig(config)
	return reaper
}

func areReaperListsEqual(reapersA, reapersB []*reaper.Reaper) bool {
	if len(reapersA) != len(reapersB) {
		return false
	}
	for idx := range reapersA {
		if strings.Compare(reapersA[idx].UUID, reapersB[idx].UUID) != 0 {
			return false
		}
	}
	return true
}
