package clouds

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
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

type Clusters struct {
	ClusterNameByLabel string
	Instances          []*compute.Instance
	ExpireDate         string
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

	//fmt.Println(expireDate)

	// get current labels and current fingerprint of the instance
	currentLabels := instance.Labels
	currentFingerprint := instance.LabelFingerprint

	// create a map of labels of existing ones
	mergedLabels := make(map[string]string)
	for k, v := range currentLabels {
		mergedLabels[k] = v
	}
	// check if instance has an expiryTag using an if statement
	if _, ok := mergedLabels["expirytag"]; !ok {
		currentTime := time.Now().UTC().Format("2006-01-02")
		expireDate := GetExpiryTag(3, currentTime)
		mergedLabels["expirytag"] = expireDate
	}

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

// Return a map with cluster name and the instances associated with it as values
// in gcp, only instances can have labels so this groups them and associates them with a cluster name the default label provides when deployed
func (gc *GCloudClient) GetClusterListByLabel() map[string]*Clusters {
	// get a list of all zones
	zones, _ := gc.ListZone()

	//create a map with cluster name as a key and struct Clusters as a value
	exist := make(map[string]*Clusters)

	// loop through all the zones
	for _, zone := range zones {
		// for each zone, grab all the instances
		instances, _ := gc.ListInstances(zone)
		// for each instance, check if it has cluster name label given by default when deployed for openshift
		for _, instance := range instances.Items {
			clusterName := getClusterNameFromLabels(instance.Labels)
			// if openshift label exist, check the map to add the cluster name / instance as a key-value pair or add to an existing key-value pair in the map
			if clusterName != "" {
				// create a new entry in the map if cluster name isn't a key
				if _, ok := exist[clusterName]; !ok {
					exist[clusterName] = &Clusters{ClusterNameByLabel: clusterName, Instances: []*compute.Instance{instance}}
					// add instance to the struct in the value of the map entry that already exists with a cluster name
				} else {
					exist[clusterName].Instances = append(exist[clusterName].Instances, instance)
				}
			}
		}
	}
	linkExpiryTagToClusterStruct(exist)
	return exist
}

// get cluster name from label, empty if no labels exist
func getClusterNameFromLabels(label map[string]string) string {
	for k, v := range label {
		if v == "owned" && strings.Contains(k, labelPrefix) {
			name := strings.TrimPrefix(k, labelPrefix)
			return name
		}
	}
	return ""
}

// get expire dates associated with the Cluster struct
func linkExpiryTagToClusterStruct(clusters map[string]*Clusters) {
	for _, v := range clusters {
		if v.ExpireDate == "" {
			for _, instance := range v.Instances {
				if _, ok := instance.Labels["expirytag"]; ok {
					v.ExpireDate = instance.Labels["expirytag"]
				}
			}
		}
	}
}

// return the zone with the api removed
func GetZone(zone string) string {
	prefix := "https://www.googleapis.com/compute/v1/projects/gc-acm-test/zones/"
	if strings.HasPrefix(zone, prefix) {
		return zone[len(prefix):]
	}
	return ""
}
