
# Internal Cloud Cleanup Tool for AWS, GCP, and Azure

Cloudscrubber is a tool used to track cloud resource usage. This tool runs against AWS, GCP, and Azure accounts and tags resources with expiry dates and generates a list of resources that are past expired. These lists containg the cluster ids needed are passed into our Jenkins destroy jobs to be destroyed appropriately


## Features

- Tag AWS vpcs for managed clusters (ROSA/OSD etc), installer provisioned installations (IPI), Kubernetes clusters (EKS)
- Announce AWS vpcs and generate files per region with expired vpcs
- Extend AWS expiryTags for clusters still in use (in development)

###
- GCP (Future Feature)
- Azure (Future Feature)


## Environment Variables

To run this project, you will need to add the following environment variables

### For AWS

`AWS_ACCESS_KEY_ID`

`AWS_SECRET_ACCESS_KEY`

`CLOUD_TASK` (use one of the following)
- awstag
- awsprint
- awsextend

To use awsextend (add the following environment variables)

`DAYS`

`CLUSTER`

`REGION`


## Deployment

To deploy this project run in docker (sample)

```bash
  docker build -t <image name> .
```

```bash
  docker run --env AWS_ACCESS_KEY_ID="<key>" --env AWS_SECRET_ACCESS_KEY="<secret>" --env CLOUD_TASK="<cloudtask>"
```

