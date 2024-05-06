package clouds

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const (
	labelPrefix = "kubernetes-io-cluster-"
	gkeLabel    = "goog-k8s-cluster-name"
)

type GCloudClient struct {
	ComputeService       *compute.Service
	ZoneService          *compute.ZonesService
	InstanceGroupService *compute.InstanceGroupsService
	InstanceService      *compute.InstancesService
	CloudConfig          *google.Credentials
}

// returns a google cloud client using a service account json
func NewGoogleCloudClient(ctx context.Context) (*GCloudClient, error) {
	filePath := os.Getenv("GCLOUD_CREDS_FILE_PATH")
	if filePath == "" {
		return nil, fmt.Errorf("failed to get the gcloud creds file path from environment variable, please check the variable GCLOUD_CREDS_FILE")
	}
	// read from a json file
	gcreds, err := os.ReadFile(os.Getenv("GCLOUD_CREDS_FILE_PATH"))
	if err != nil {
		fmt.Printf("failed to read file from path")
	}
	// call the credentials
	cfg, err := google.CredentialsFromJSON(ctx, gcreds)
	if err != nil {
		fmt.Printf("failed to parse the config due to %v\n", err)
		return &GCloudClient{}, err
	}

	options := []option.ClientOption{
		option.WithCredentialsJSON(gcreds),
	}

	cs, err := compute.NewService(ctx, options...)
	if err != nil {
		return &GCloudClient{}, err
	}

	zs := compute.NewZonesService(cs)
	is := compute.NewInstancesService(cs)
	igs := compute.NewInstanceGroupsService(cs)
	return &GCloudClient{
		ComputeService:       cs,
		ZoneService:          zs,
		InstanceGroupService: igs,
		InstanceService:      is,
		CloudConfig:          cfg,
	}, nil
}

// list all the available zones on the project
func (gc *GCloudClient) ListZone() ([]string, error) {
	var zones []string

	zs, err := gc.ZoneService.List(gc.CloudConfig.ProjectID).Do()
	if err != nil {
		return []string{}, nil
	}
	for _, v := range zs.Items {
		zones = append(zones, v.Name)
	}
	return zones, nil
}

// list instance groups
func (gc *GCloudClient) ListInstanceGroup(zone string) (*compute.InstanceGroupList, error) {
	cs, err := gc.InstanceGroupService.List(gc.CloudConfig.ProjectID, zone).Do()
	if err != nil {
		return &compute.InstanceGroupList{}, err
	}
	return cs, nil
}

// list vm instances
func (gc *GCloudClient) ListInstances(zone string) (*compute.InstanceList, error) {
	is, err := gc.InstanceService.List(gc.CloudConfig.ProjectID, zone).Do()
	if err != nil {
		return &compute.InstanceList{}, err
	}
	return is, nil
}

// function to add a label to existing labels for a single instance
func (gc *GCloudClient) LabelInstance(project string, zone string, instance *compute.Instance) {
	ctx := context.Background()

	currentTime := time.Now().UTC().Format("2006-01-02")
	expireDate := GetExpiryTag(3, currentTime)
	//fmt.Println(expireDate)

	// get current labels and current fingerprint of the instance
	currentLabels := instance.Labels
	currentFingerprint := instance.LabelFingerprint

	// create a map of labels of existing ones
	mergedLabels := make(map[string]string)
	for k, v := range currentLabels {
		mergedLabels[k] = v
	}
	// add the expiry tag label
	mergedLabels["expirytag"] = expireDate

	// create the request
	request := gc.ComputeService.Instances.SetLabels(project, zone, instance.Name, &compute.InstancesSetLabelsRequest{
		Labels:           mergedLabels,
		LabelFingerprint: currentFingerprint,
	})

	requestWithContext := request.Context(ctx)

	// Execute the API request and get the response
	_, err := requestWithContext.Do()
	if err != nil {
		// Handle the error
		log.Fatalf("Failed to set labels for instance: %v", err)
	}

	//fmt.Println("Labels added successfully!")
}
