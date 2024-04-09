package clouds

import (
	"context"
	"fmt"

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

func (ac *AWSClient) GetVpc() ([]*ec2.Vpc, error) {
	result, err := ac.AWSEC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return result.Vpcs, nil
}

func (ac *AWSClient) GetInstance() ([]*ec2.Reservation, error) {
	result, err := ac.AWSEC2Client.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}
	return result.Reservations, nil
}

func TagVpcInstance(region string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		klog.Errorf("Error loading AWS config:")
	}

	clientResource := resourcegroupstaggingapi.NewFromConfig(cfg)

	// create an arn for vpcs
	//region := "us-east-2"
	vpc := "vpc-0cbefbf003d21894b"
	vpcARN := "arn:aws:ec2:" + region + ":" + "902449478968:vpc/" + vpc

	fmt.Println(vpcARN)

	value := "testValue"

	input := resourcegroupstaggingapi.TagResourcesInput{
		ResourceARNList: []string{
			vpcARN,
		},
		Tags: map[string]string{
			"QEName": value,
		},
	}

	output, err := clientResource.TagResources(context.TODO(), &input)
	if err != nil {
		fmt.Printf("Error tagging vpc")
	}
	fmt.Println(output)

	//res, _ := ac.GetInstance()

	// loop all instances in us-east-2
	/*
		for _, v := range res {
			//t.Log(v)
			for _, i := range v.Instances {
				//t.Log(*i.IamInstanceProfile.Arn)
				// t.Log(i)
				// break
			}
			//break
		}
	*/
}

func GetVpcArn() {

}
