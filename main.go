package main

import (
	"fmt"
	"os"
	"time"

	"github.com/stolostron/cloudscrubber/pkg/clouds"
	"k8s.io/klog"
)

var (
	awsRegion = []string{
		// "us-east-1",
		// "us-east-2",
		"us-west-1",
		// "us-west-2",
	}
)

func main() {
	// Get env variable to determine what task to run
	cloudTask := os.Getenv("cloudtask")

	switch cloudTask {
	case "awstag":
		fmt.Println("test aws tag")
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
		fmt.Println("test aws print")
	}
}
