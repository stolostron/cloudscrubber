package clouds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type AzureClient struct {
	AzureClientSecret        *azidentity.ClientSecretCredential
	AzureResourceGroupClient *armresources.ResourceGroupsClient
	AzureVMClient            *armcompute.VirtualMachinesClient
}

type AzureClusters struct {
	IPI   []*armresources.ResourceGroup
	AKS   []*armresources.ResourceGroup
	OTHER []*armresources.ResourceGroup
}

var AzureIgnoreList = []string{
	"DefaultResourceGroup-EUS",
	"NetworkWatcherRG",
	"os4-common",
	"domain",
	"az.red-chesterfield.com",
	"cloud-shell-storage-",
	"jnp-a_group",
	"MC_jnp-a_group_jnp-aro_centralus",
	"MA_defaultazuremonitorworkspace-cus_centralus_managed",
	"MC_clc-aks",
}

func NewAzureClient(tenantID, clientID, clientSecret, subscriptionID string) (*AzureClient, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return &AzureClient{}, err
	}
	regc, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return &AzureClient{}, err
	}
	vmc, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return &AzureClient{}, err
	}
	return &AzureClient{
		AzureClientSecret:        cred,
		AzureResourceGroupClient: regc,
		AzureVMClient:            vmc,
	}, nil
}

// list of all resource groups associated with the subscription
func (az *AzureClient) ListResourceGroup(ctx context.Context) (*armresources.ResourceGroupsClientListResponse, error) {
	rep := az.AzureResourceGroupClient.NewListPager(nil)
	if rep.More() {
		res, err := rep.NextPage(ctx)
		if err != nil {
			return &armresources.ResourceGroupsClientListResponse{}, err
		}
		return &res, nil
	}
	return &armresources.ResourceGroupsClientListResponse{}, fmt.Errorf("failed to get the azure resources")
}

// tag resource group
func (az *AzureClient) TagResourceGroup(resourcegroup string, ctx context.Context) {
	// existing tags
	rg, _ := az.AzureResourceGroupClient.Get(ctx, resourcegroup, nil)
	tag := rg.Tags

	// create expiry tag
	currentTime := time.Now().UTC().Format("2006-01-02")
	expireDate := GetExpiryTag(3, currentTime)
	expiryValue := expireDate
	tag["expirytag"] = &expiryValue

	// create update patch with existing + expiry tag
	update := armresources.ResourceGroupPatchable{
		Tags: tag,
	}
	// call the update
	az.AzureResourceGroupClient.Update(ctx, resourcegroup, update, nil)
}

// check if a cluster name is in the slice of the ignore list
func AzureContains(s string, ss []string) bool {
	for _, v := range ss {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

// Tag resource groups not in the ignore list using the azure struct
func (az *AzureClient) TagAzureClusters(clusters []*armresources.ResourceGroup, ctx context.Context) {
	azClusters := GetAzureClustersByType(clusters)

	for _, cluster := range azClusters.IPI {
		if _, ok := cluster.Tags["expirytag"]; !ok {
			az.TagResourceGroup(*cluster.Name, ctx)
			//fmt.Println(*cluster.Name)
		}
	}
	for _, cluster := range azClusters.AKS {
		if _, ok := cluster.Tags["expirytag"]; !ok {
			az.TagResourceGroup(*cluster.Name, ctx)
			//fmt.Println(*cluster.Name)
		}
	}
	for _, cluster := range azClusters.OTHER {
		if _, ok := cluster.Tags["expirytag"]; !ok {
			az.TagResourceGroup(*cluster.Name, ctx)
			//fmt.Println(*cluster.Name)
		}
	}
}

// Finds all IPI and AKS resource groups
func GetAzureClustersByType(clusters []*armresources.ResourceGroup) AzureClusters {
	AzureStruct := AzureClusters{
		IPI:   []*armresources.ResourceGroup{},
		AKS:   []*armresources.ResourceGroup{},
		OTHER: []*armresources.ResourceGroup{},
	}
	aks_nodes := make([]*armresources.ResourceGroup, 0)

	for _, cluster := range clusters {
		// check if resource group isnt in the ignore list
		if AzureContains(*cluster.Name, AzureIgnoreList) {
			continue
		}
		// Iterate over the Tags map and find either IPI clusters or aks node resource groups
		key := getClusterKey(*cluster.Name)
		if _, ok := cluster.Tags[key]; ok {
			AzureStruct.IPI = append(AzureStruct.IPI, cluster)
		} else if _, ok := cluster.Tags["aks-managed-cluster-rg"]; ok {
			//fmt.Println(*cluster.Name)
			aks_nodes = append(aks_nodes, cluster)
		}
	}
	// find the aks resource group that the aks node resource group is managing
	// as that is the resource group used to destroy both resource group associated with an AKS cluster
	for _, aks_node := range aks_nodes {
		for _, cluster := range clusters {
			if *aks_node.Tags["aks-managed-cluster-rg"] == *cluster.Name {
				//fmt.Println(*cluster.Name)
				AzureStruct.AKS = append(AzureStruct.AKS, cluster)
			}
		}
	}

	// create a map that is used to find the difference between all the resourcegroups and IPI + AKS resources groups
	tempMap := make(map[*armresources.ResourceGroup]bool)
	combinedClusters := append(AzureStruct.IPI, AzureStruct.AKS...)
	for _, cluster := range combinedClusters {
		//fmt.Println(*cluster.Name)
		tempMap[cluster] = true
	}

	// use the map to see if it doesn't exist in the current set of all resourcegroups
	// if not, add it to OTHER in the struct
	for _, cluster := range clusters {
		if _, ok := tempMap[cluster]; !ok && !AzureContains(*cluster.Name, AzureIgnoreList) {
			AzureStruct.OTHER = append(AzureStruct.OTHER, cluster)
		}
	}
	return AzureStruct
}

// Map key name used to find IPI clusters
func getClusterKey(resourceGroupName string) string {
	return "kubernetes.io_cluster." + strings.TrimSuffix(resourceGroupName, "-rg")
}

// return a struct with resources groups that are expired
func GetExpiredResourceGroups(clusters []*armresources.ResourceGroup) AzureClusters {
	azureclusters := GetAzureClustersByType(clusters)

	expiredClusters := AzureClusters{}

	for _, cluster := range azureclusters.IPI {
		if _, ok := cluster.Tags["expirytag"]; ok && IsExpired(*cluster.Tags["expirytag"]) {
			//fmt.Println(*cluster.Name)
			expiredClusters.IPI = append(expiredClusters.IPI, cluster)
		}
	}
	for _, cluster := range azureclusters.AKS {
		if _, ok := cluster.Tags["expirytag"]; ok && IsExpired(*cluster.Tags["expirytag"]) {
			//fmt.Println(*cluster.Name)
			expiredClusters.AKS = append(expiredClusters.AKS, cluster)
		}
	}
	for _, cluster := range azureclusters.OTHER {
		if _, ok := cluster.Tags["expirytag"]; ok && IsExpired(*cluster.Tags["expirytag"]) {
			//fmt.Println(*cluster.Name)
			expiredClusters.OTHER = append(expiredClusters.OTHER, cluster)
		}
	}

	return expiredClusters
}

// Print all the resourcegroups that are expired using the expired tag
func PrintExpiredResourceGroups(rg []*armresources.ResourceGroup) {
	expiredClusters := GetExpiredResourceGroups(rg)
	fmt.Println("IPI Clusters")
	for _, cluster := range expiredClusters.IPI {
		fmt.Println(*cluster.Name)
	}
	fmt.Println("\nAKS/Managed Clusters Clusters")
	for _, cluster := range expiredClusters.AKS {
		fmt.Println(*cluster.Name)
	}
	fmt.Println("\nOther/Non-Cluster Resourcegroups")
	for _, cluster := range expiredClusters.OTHER {
		fmt.Println(*cluster.Name)
	}
}

// extend expiry date for an azure resource group
func (az *AzureClient) ExtendAzureCluster(clusterName string, days int, clusters []*armresources.ResourceGroup, ctx context.Context) {
	for _, cluster := range clusters {
		if _, ok := cluster.Tags["expirytag"]; ok && *cluster.Name == clusterName {
			az.extendAzureExpiryDate(*cluster.Name, days, ctx)
		}
	}
}

// helper function to extend expiry tag
func (az *AzureClient) extendAzureExpiryDate(resourcegroup string, days int, ctx context.Context) {
	rg, _ := az.AzureResourceGroupClient.Get(ctx, resourcegroup, nil)
	tag := rg.Tags

	// create expiry tag
	currentTime := time.Now().UTC().Format("2006-01-02")
	expireDate := GetExpiryTag(days, currentTime)
	expiryValue := expireDate
	tag["expirytag"] = &expiryValue

	// create update patch with existing + expiry tag
	update := armresources.ResourceGroupPatchable{
		Tags: tag,
	}
	// call the update
	az.AzureResourceGroupClient.Update(ctx, resourcegroup, update, nil)
}
