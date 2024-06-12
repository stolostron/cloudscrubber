
# Internal Cloud Cleanup Tool for AWS, GCP, and Azure

Cloudscrubber is a tool used to track cloud resource usage. This tool runs against AWS, GCP, and Azure accounts and tags resources with expiry dates and generates a list of resources that are past expired. These lists containg the cluster ids needed are passed into our Jenkins destroy jobs to be destroyed appropriately


## Features

- Tag AWS vpcs (Kubernetes, IPI, managed) with expiry tags using `awstag`
- Generate files per region with expired vpcs using `awsprint`
- Ability to extend dates for expiry tags using `awsextend`

###
- Tag GCP instances (Kubernetes, IPI) with expiry tags using `gcptag`
- Generate a list of cluster names with region listed using `gcpprint`
- Ability to extend dates for expiry tags using `gcpextend`

###
- Tag Azure resource groups (Kubernetes, IPI, managed) with expiry tags using `azuretag`
- Generate a list of resourcegroups that have expired tags using `azureprint`
- Ability to extend dates for expiry tags using `azureextend`


## Environment Variables

To run this project, you will need to add the following environment variables along with credentials for the cloud provider you are using.

`CLOUD_TASK`

### For AWS

`AWS_ACCESS_KEY_ID`

`AWS_SECRET_ACCESS_KEY`

`DAYS` (for awsextend)

`CLUSTER` (for awsextend, pass the infra id like `testcluster-12345`)

`REGION` (for awsextend, region should be an aws region like `us-east-1`)

### For GCP

For GCP creds, you will need the filepath to your serviceaccount.json filepath

`GCLOUD_CREDS_FILE_PATH`

`DAYS` (for gcpextend)

`CLUSTER` (for gcpextend, pass the infra id like `testcluster-12345`)

### For Azure

`TENANT_ID`

`CLIENT_ID`

`SUBSCRIPTION_ID`

`CLIENT_SECRET`

`DAYS` (for azureextend)

`CLUSTER` (for azureextend, pass the full resourcegroup name like `testcluster-12345-rg`)
## Deployment

To deploy this project run in docker (sample)

```bash
  docker build -t <image name> .
```

```bash
  docker run --env AWS_ACCESS_KEY_ID="<key>" --env AWS_SECRET_ACCESS_KEY="<secret>" --env CLOUD_TASK="<cloudtask>"
```

Example to tag aws vpcs with expiry tags

```bash
  docker run --env AWS_ACCESS_KEY_ID="paste_access_key_here" --env AWS_SECRET_ACCESS_KEY="paste_secret_key_here" --env CLOUD_TASK="awstag"
```