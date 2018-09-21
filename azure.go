package main

import (
    "log"
    "context"
    "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
    "github.com/Azure/go-autorest/autorest/azure/auth"
 //   "github.com/Azure/go-autorest/autorest/to"
)

type AzureSubnet struct {
    prefix string
    subnet uint32
    name string
}

type Azure struct {
    vnetClient network.VirtualNetworksClient
    subnetsClient network.SubnetsClient
    routesClient network.RoutesClient
    ctx context.Context
    rgroup string
    vnet string
    routetable string
    nsgs []string
}

func NewAzure(subscriptionId string, clientId string, clientSecret string, tenantId string, rgroup string, vnet string, routetable string, context context.Context, nsgs []string) (*Azure){
    creds := auth.NewClientCredentialsConfig(clientId, clientSecret, tenantId)
    auth, err:= creds.Authorizer()
    if (err != nil) {
        log.Printf("Cannot authenticate with Azure\n")
        //TODO: terminate app
        return nil
    }

    azure := new(Azure)
    azure.vnet = vnet
    azure.rgroup = rgroup
    azure.routetable = routetable
    azure.ctx = context
    azure.vnetClient = network.NewVirtualNetworksClient(subscriptionId)
    azure.vnetClient.Authorizer = auth
    azure.nsgs = nsgs

    azure.subnetsClient = network.NewSubnetsClient(subscriptionId)
    azure.subnetsClient.Authorizer = auth

    azure.routesClient = network.NewRoutesClient(subscriptionId)
    azure.routesClient.Authorizer = auth

    return azure
}

func (self *Azure) ChangeUplink(address string) {
    log.Printf("Will make a change in Azure\n")
    for list, err := self.routesClient.ListComplete(self.ctx, self.rgroup, self.routetable); list.NotDone(); err = list.Next() {
        if err != nil {
            log.Printf("Could not find any routes in Azure\n")
            return
        }
        r := list.Value()
        for _,nsg := range self.nsgs {
            if *r.NextHopIPAddress == nsg {
                log.Printf("Will change next hop for route %s\n",*r.Name)
                *r.NextHopIPAddress = address
                self.routesClient.CreateOrUpdate(self.ctx, self.rgroup, self.routetable, *r.Name, r)
            }
        }
    }
}

func (self *Azure) GetSubnets() ([]*AzureSubnet) {
    var nets []*AzureSubnet

    for list, err := self.subnetsClient.ListComplete(self.ctx, self.rgroup, self.vnet); list.NotDone(); err = list.Next() {
        if err != nil {
            log.Printf("Could not find any subnets in Azure\n")
            return nets
        }

        as := new(AzureSubnet)
        as.prefix = *list.Value().SubnetPropertiesFormat.AddressPrefix
        as.name = *list.Value().Name
        nets = append(nets,as)
    }
    return nets
}
