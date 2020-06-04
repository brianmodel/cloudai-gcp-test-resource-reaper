package clients

import (
	"context"
	"time"

	reaperpb "github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	gce "google.golang.org/api/compute/v1"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
)

func (client *GCEClient) Auth() error {
	ctx := context.Background()
	authedClient, err := gce.NewService(ctx)
	if err != nil {
		return err
	}
	client.Client = authedClient
	return nil
}

// Can either get all resources and filter elsewhere, or filter in here (latter is more efficient)
func (client *GCEClient) GetResources(projectID string, config reaperpb.ResourceConfig) ([]resources.Resource, error) {
	var instances []resources.Resource
	zones := config.GetZones()
	for _, zone := range zones {
		zoneInstancesCall := client.Client.Instances.List(projectID, zone)
		instancesInZone, err := zoneInstancesCall.Do()
		if err != nil {
			return nil, err
		}
		for _, instance := range instancesInZone.Items {
			timeCreated, _ := time.Parse(time.RFC3339, instance.CreationTimestamp)
			parsedResource := resources.NewResource(instance.Name, zone, timeCreated, reaperpb.ResourceType_GCE_VM)
			if resources.ShouldAddResourceToWatchlist(parsedResource, config.GetNameFilter(), config.GetSkipFilter()) {
				instances = append(instances, parsedResource)
			}
		}
	}
	return instances, nil
}

func (client *GCEClient) DeleteResource(projectID string, resource resources.Resource) error {
	client.Client.Instances.Delete(projectID, resource.Zone, resource.Name)
}

// withTransport
// https://pkg.go.dev/google.golang.org/api/option/?tab=doc
// ExampleTest
// Hermetic??