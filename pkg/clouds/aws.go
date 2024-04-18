package clouds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"k8s.io/klog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
)

type AWSClient struct {
	AWSEC2Client *ec2.EC2
	AWSEKSClient *eks.EKS
}

type taggedVpc struct {
	vpcId     string
	expiryTag string
}

type VpcType struct {
	Rosa []string
	Ipi  []string
	Eks  []string
}

func NewAWSClient(region string) (*AWSClient, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(region),
		},
	)
	if err != nil {
		return nil, err
	}

	// Create ec2 service client
	ec2svc := ec2.New(sess)
	// Create the eks service client
	ekssvc := eks.New(sess)
	return &AWSClient{
		AWSEC2Client: ec2svc,
		AWSEKSClient: ekssvc,
	}, nil
}

func (ac *AWSClient) GetVpcs() ([]*ec2.Vpc, error) {
	result, err := ac.AWSEC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return result.Vpcs, nil
}

func (ac *AWSClient) GetInstances() ([]*ec2.Reservation, error) {
	result, err := ac.AWSEC2Client.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}
	return result.Reservations, nil
}

// Tag vpc instances with expiry tag
func TagVpcInstance(region string, vpcId string, creationTime string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		klog.Errorf("Error loading AWS config:")
	}

	clientResource := resourcegroupstaggingapi.NewFromConfig(cfg)

	// create an arn for vpcs
	// region := "us-east-2"
	// vpcId := "vpc-0cbefbf003d21894b"
	vpcARN := "arn:aws:ec2:" + region + ":" + "902449478968:vpc/" + vpcId

	fmt.Println(vpcARN)

	//expireDate := "testValue"
	// func to create expiry tag
	expireDate := GetExpiryTag(vpcId, creationTime)

	// used to tag a vpc
	input := resourcegroupstaggingapi.TagResourcesInput{
		ResourceARNList: []string{
			vpcARN,
		},
		Tags: map[string]string{
			"expiryTag": expireDate,
		},
	}

	output, err := clientResource.TagResources(context.TODO(), &input)
	if err != nil {
		fmt.Printf("Error tagging vpc")
	}
	fmt.Println(output)
}

// Takes creationTime and creates expiryTag with 3 days
func GetExpiryTag(vpcId string, creationTime string) string {
	date, err := time.Parse("2006-01-02", creationTime)
	if err != nil {
		klog.Errorf("failed getting expiryTag from vpc or creationTime")
	}

	// Add two days to the date
	newDate := date.Add(3 * 24 * time.Hour)

	// Format the new date as a string
	newDateString := newDate.Format("2006-01-02")

	return newDateString
}

// get list of vpcId from list of vpcs in a region
func (ac *AWSClient) GetVpcArn() (vpcId []string) {
	vpcs, err := ac.GetVpcs()
	if err != nil {
		klog.Errorf("failed getting list of vpc")
	}
	// create slice to get the list of VpcIds
	vpcIdList := []string{}
	for _, vpc := range vpcs {
		vpcIdList = append(vpcIdList, *vpc.VpcId)
		//fmt.Println(*vpc.VpcId)
		//fmt.Println()
	}
	return vpcIdList
}

// Get vpc ids without expiry tag
func (ac *AWSClient) GetVpcArnWithoutExpiryTag() (vpcId []string) {
	vpcs, err := ac.GetVpcs()
	if err != nil {
		klog.Errorf("failed getting list of vpc")
	}
	// create slice to get the list of VpcIds
	vpcIdList := []string{}
	tagPresent := false
	for _, vpc := range vpcs {
		for _, tag := range vpc.Tags {
			if *tag.Key == "expiryTag" {
				tagPresent = true
			}
		}
		// expireTag wasnt there, else it was found and set to false for next vpc
		if !tagPresent {
			vpcIdList = append(vpcIdList, *vpc.VpcId)
		} else {
			tagPresent = false
		}
		//fmt.Println(*vpc.VpcId)
		//fmt.Println()
	}
	return vpcIdList
}

// return a struct with vpcs separated by its type
func (ac *AWSClient) GetVpcTypesThatAreExpired() VpcType {
	// clusters := VpcType{}
	// vpcs, _ := ac.GetVpcs()
	// expiredVpcs := ac.GetVpcArnWithExpiryTag()
	// for _, expiredVpc := range expiredVpcs {
	// 	for _, vpc := range vpcs {
	// 		if expiredVpc.vpcId == *vpc.VpcId {
	// 			for _, tag := range vpc.Tags {
	// 				if strings.Contains(*tag.Key, "red-hat-managed") {
	// 					clusters.rosa = append(clusters.rosa, getClusterName(vpc.Tags))
	// 				} else if strings.Contains(*tag.Key, "kubernetes.io/cluster/") {
	// 					clusters.ipi = append(clusters.ipi, getClusterName(vpc.Tags))
	// 				} else {
	// 					clusters.eks = append(clusters.eks, getClusterName(vpc.Tags))
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// return clusters
	clusters := VpcType{}
	vpcs, _ := ac.GetVpcs()
	expiredVpcs := ac.GetVpcArnWithExpiryTag()

	// Create a map to track processed VPC IDs
	processed := make(map[string]bool)

	for _, expiredVpc := range expiredVpcs {
		// Skip if the VPC ID has already been processed
		if processed[expiredVpc.vpcId] {
			continue
		}

		for _, vpc := range vpcs {
			if expiredVpc.vpcId == *vpc.VpcId {
				for _, tag := range vpc.Tags {
					if strings.Contains(*tag.Key, "red-hat-managed") {
						clusters.Rosa = append(clusters.Rosa, getClusterName(vpc.Tags))
						break
					} else if strings.Contains(*tag.Key, "kubernetes.io/cluster/") {
						clusters.Ipi = append(clusters.Ipi, getClusterName(vpc.Tags))
						break
					} else if strings.Contains(*tag.Key, "alpha.eksctl.io/cluster-name") {
						clusters.Eks = append(clusters.Eks, getClusterName(vpc.Tags))
						break
					}
				}

				// Mark the VPC ID as processed
				processed[expiredVpc.vpcId] = true

				// Break out of the inner loop since the VPC has been processed
				break
			}
		}
	}
	return clusters
}

// helper function to get the value from the tag key as cluster name
func getClusterName(tags []*ec2.Tag) string {
	var clusterName string
	for _, tag := range tags {
		if *tag.Key == "Name" {
			clusterName = *tag.Value
		}
	}
	return clusterName
}

// returns a struct with vpcs that have expired tags
func (ac *AWSClient) GetVpcArnWithExpiryTag() []taggedVpc {
	vpcs, err := ac.GetVpcs()
	if err != nil {
		klog.Errorf("failed getting list of vpc")
	}
	expiredVpcs := []taggedVpc{}
	for _, vpc := range vpcs {
		for _, tag := range vpc.Tags {
			if *tag.Key == "expiryTag" && IsExpired(*tag.Value) {
				//vpcIdList = append(vpcIdList, *vpc.VpcId)
				expiredVpcs = append(expiredVpcs, taggedVpc{vpcId: *vpc.VpcId, expiryTag: *tag.Value})
			}
		}
	}
	return expiredVpcs
}

// check if date is expired from tag
func IsExpired(tagDate string) bool {
	currentDate := time.Now().Format("2006-01-02")
	compareString := tagDate
	return currentDate > compareString
}

// map vpcIds with a creationTime from list of instances
func (ac *AWSClient) MapVpcIdsWithCreationTime() map[string]string {
	instanceList, _ := ac.GetInstances()
	//map creation time value with vpcId key
	vpcMap := make(map[string]string)
	//vpcIds := ac.GetVpcArn()
	vpcIds := ac.GetVpcArnWithoutExpiryTag()
	timeFormat := "2006-01-02"
	// iterate through list of vpcs on a region
	for _, vpcID := range vpcIds {
		// iterate through list of instances on a region (note multiple instances can have the same vpc field)
		for _, v := range instanceList {
			// v is a struct and v.instance is also a struct with instance info
			// iterate through the vpc field for i and add a creationTime value for each vpc and make a map
			for _, i := range v.Instances {
				if i.VpcId != nil && *i.VpcId == vpcID {
					// Check if the VPC ID is already in the map
					if _, ok := vpcMap[vpcID]; !ok {
						// Add the VPC ID to the map if not there
						vpcMap[vpcID] = i.LaunchTime.Format(timeFormat)
					}
				}
			}
		}
	}

	for k, v := range vpcMap {
		fmt.Printf("Key: %s, Value: %s\n", k, v)
	}
	//fmt.Println(len(vpcMap))
	//fmt.Println(len(vpcIds))
	//fmt.Println(vpcMap)
	return vpcMap
}

// Create func to get vpcs without instances / creationTime
