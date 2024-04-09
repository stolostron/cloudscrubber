package main

import (
	"cloudmodule/pkg/clouds"
	"fmt"

	"k8s.io/klog"
)

func main() {
	ac, err := clouds.NewAWSClient("us-east-1")
	if err != nil {
		klog.Errorf("failed %v\n", err)
	}
	vpcs, err := ac.GetVpc()
	if err != nil {
		klog.Errorf("failed getting vpc")
	}

	for _, vpc := range vpcs {
		//t.Log(vpc)
		for _, tag := range vpc.Tags {
			fmt.Println(*tag.Key)
			fmt.Println(*tag.Value)
		}
		fmt.Println()
	}
}
