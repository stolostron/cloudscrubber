package main

import (
	clouds "cloudmodule/pkg/clouds"
	"fmt"

	"k8s.io/klog"
)

var (
	awsRegion = []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
)

func main() {
	// create aws client for each region

	ac, err := clouds.NewAWSClient("us-east-1")
	if err != nil {
		klog.Errorf("failed %v\n", err)
	}
	vpcMap := ac.MapVpcIdsWithCreationTime()
	//newTime := clouds.GetExpiryTag("vpc-063bb605484654c0f", vpcMap["vpc-063bb605484654c0f"])
	//fmt.Println(newTime)

	fmt.Println(len(vpcMap))
}
