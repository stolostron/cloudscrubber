
# Internal Cloud Cleanup Tool for AWS, GCP, and Azure

Cloudscrubber is a tool used to track cloud resource usage. This tool runs against AWS, GCP, and Azure accounts and tags resources with expiry dates and generates a list of resources that are past expired. These lists containg the cluster ids needed are passed into our Jenkins destroy jobs to be destroyed appropriately


## Features

- Tag AWS vpcs for managed clusters (ROSA/OSD etc), installer provisioned installations (IPI), Kubernetes clusters (EKS)
- Generate files per region with expired vpcs that were tagged with `awstag`
- Extend AWS expiryTags for clusters still in use

- Tag GCP instances that have openshift provisioned labels currently for installer provisioned installations (IPI), Kubernetes clusters (GKE)
- Generate a list of cluster names with region listed
- Extend GCP labels for clusters still in use

###
- Azure (Future Feature)


## Environment Variables

To run this project, you will need to add the following environment variables

`CLOUD_TASK` (use one of the following)

- awstag
- awsprint
- awsextend
- gcptag
- gcpprint
- gcpextend
- azuretag
- azureprint
- azureextend

### For AWS

`AWS_ACCESS_KEY_ID`

`AWS_SECRET_ACCESS_KEY`

`DAYS` (for awsextend)

`CLUSTER` (for awsextend)

`REGION` (for awsextend)

### FOR GCP

For GCP creds, you will need the filepath to your serviceaccount.json filepath

`GCLOUD_CREDS_FILE_PATH`

`DAYS` (for gcpextend)

`CLUSTER` (for gcpextend)
## Deployment

To deploy this project run in docker (sample)

```bash
  docker build -t <image name> .
```

```bash
  docker run --env AWS_ACCESS_KEY_ID="<key>" --env AWS_SECRET_ACCESS_KEY="<secret>" --env CLOUD_TASK="<cloudtask>"
```

