package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/stolostron/cloudscrubber/pkg/clouds"
	"k8s.io/klog"
)

var (
	ctx       = context.Background()
	awsRegion = []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
)

func main() {
	// Get env variable to determine what task to run
	cloudTask := os.Getenv("CLOUD_TASK")

	switch cloudTask {
	case "awstag":
		fmt.Println("Tagging aws vpcs without expiryTags")
		// create aws client for each region
		for _, region := range awsRegion {
			ac, err := clouds.NewAWSClient(region)
			if err != nil {
				klog.Errorf("failed to create aws client %v\n", err)
			}
			// get list of vpcs that dont have expiry tags
			vpcsWithNoTags := ac.GetVpcArnWithoutExpiryTag()

			// get current date
			currentTime := time.Now().UTC().Format("2006-01-02")

			for _, vpcId := range vpcsWithNoTags {
				clouds.TagVpcInstance(region, vpcId, currentTime)
			}
		}
	case "awsprint":
		fmt.Println("Outputting aws vpcs that are expired")
		for _, region := range awsRegion {
			ac, err := clouds.NewAWSClient(region)
			if err != nil {
				klog.Errorf("failed to create aws client %v\n", err)
			}

			vpcs := ac.GetVpcTypesThatAreExpired()
			//fmt.Println(vpcs)

			file, err := os.Create("cloudoutput/" + region + "-aws.txt")
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			// Redirect standard output to the file
			os.Stdout = file
			clouds.GenerateFiles(region, vpcs)
		}
	case "awsextend":
		daysExtended := os.Getenv("DAYS")
		clusterName := os.Getenv("CLUSTER")
		region := os.Getenv("REGION")

		days, _ := strconv.Atoi(daysExtended)

		ac, err := clouds.NewAWSClient(region)
		if err != nil {
			klog.Errorf("failed to create aws client %v\n", err)
		}
		ac.ExtendExpiryTag(region, clusterName, days)
	case "gcptag":
		gc, err := clouds.NewGoogleCloudClient(ctx)
		if err != nil {
			klog.Errorf("failed to create the cloud client due to: %v\n", err)
		}
		if err != nil {
			klog.Errorf("failed to get zones due to: %v\n", err)
		}
		project := gc.CloudConfig.ProjectID

		clusters := gc.GetClusterListByLabel()
		for _, cluster := range clusters {
			for _, instance := range cluster.Instances {
				gc.LabelInstance(project, clouds.GetZone(instance.Zone), instance)
			}
		}
	case "gcpprint":

	}
	run()
}

// export GCLOUD_CREDS_FILE_PATH=~/Desktop/Cloud/osServiceAccount.json
func run() {
	gc, err := clouds.NewGoogleCloudClient(ctx)
	if err != nil {
		klog.Errorf("failed to create the cloud client due to: %v\n", err)
	}

	clusters := gc.GetClusterListByLabel()
	for k, v := range clusters {
		fmt.Println(k, v.ExpireDate)
	}
}
