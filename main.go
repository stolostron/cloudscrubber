package main

import (
	"cloudmodule/pkg/clouds"

	"k8s.io/klog"
)

func main() {
	ac, err := clouds.NewAWSClient("us-east-1")
	if err != nil {
		klog.Errorf("failed %v\n", err)
	}
	ac.MapVpcIdsWithCreationTime()
}
