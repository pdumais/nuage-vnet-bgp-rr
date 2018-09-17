package main

import (
    "fmt"
    "context"
    "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
    "github.com/Azure/go-autorest/autorest/azure/auth"
 //   "github.com/Azure/go-autorest/autorest/to"
)

type AzureSubnet struct {
    address string
    subnet uint32
}

type Azure struct {
    vnetClient network.VirtualNetworksClient
    subnetsClient network.SubnetsClient
    ctx context.Context
}

func NewAzure(subscriptionId string, clientId string, clientSecret string, tenantId string, context context.Context) (*Azure){
    creds := auth.NewClientCredentialsConfig(clientId, clientSecret, tenantId)
    auth, err:= creds.Authorizer()
    if (err != nil) {
        fmt.Printf("Cannot authenticate with Azure\n")
        return nil
    }

    azure := new(Azure)
    azure.ctx = context
    azure.vnetClient = network.NewVirtualNetworksClient(subscriptionId)
    azure.vnetClient.Authorizer = auth

    azure.subnetsClient = network.NewSubnetsClient(subscriptionId)
    azure.subnetsClient.Authorizer = auth

    return azure
}

func (self *Azure) ChangeUplink(address string) {
    fmt.Printf("Will make a change in Azure\n")
    //TODO
}

func (self *Azure) GetSubnets(vnet string) ([]*AzureSubnet) {
    var nets []*AzureSubnet
    self.subnetsClient.ListComplete(self.ctx, "", vnet)
    //TODO
    return nets
}
