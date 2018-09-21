package main

import (
    "log"
    "context"
    "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
    "github.com/Azure/go-autorest/autorest/azure/auth"
    "strings"
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
    nsgs []string
    routeTables map[string]string
}

func NewAzure(subscriptionId string, clientId string, clientSecret string, tenantId string, rgroup string, vnet string, context context.Context, nsgs []string) (*Azure){
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

    for rt,_ := range self.routeTables {
        routes := self.getNsgRoutes(rt)
        for _,r := range routes {
            log.Printf("Will change next hop for route %s\n",*r.Name)
            *r.NextHopIPAddress = address
            self.routesClient.CreateOrUpdate(self.ctx, self.rgroup, rt, *r.Name, r)
        }
    }
}

func (self *Azure) getNsgRoutes(name string) ([]network.Route){
    var routeList []network.Route

    for list, err := self.routesClient.ListComplete(self.ctx, self.rgroup, name); list.NotDone(); err = list.Next() {
        if (err != nil) {
            return routeList
        }
        r := list.Value()
        for _,nsg := range self.nsgs {
            if *r.NextHopIPAddress == nsg {
                routeList = append(routeList,r)
            }
        }
    }
    return routeList
}

func (self *Azure) GetSubnets() ([]*AzureSubnet) {
    var nets []*AzureSubnet
    self.routeTables = make(map[string]string)

    for list, err := self.subnetsClient.ListComplete(self.ctx, self.rgroup, self.vnet); list.NotDone(); err = list.Next() {
        if err != nil {
            log.Printf("Could not find any subnets in Azure\n")
            return nets
        }

        // we are not interested in subjects that are not attached to a route table that contains at least one route with an NSG as 
        // its next hop. Those subnets won't be advertised and they will (obviously) not be modified when an NSG changes status.
        rt := list.Value().SubnetPropertiesFormat.RouteTable
        if (rt == nil) {
            log.Printf("Found subnet %s but will ignore it because it has no route table\n",*list.Value().SubnetPropertiesFormat.AddressPrefix)
            continue
        }

        tmp := strings.Split(*list.Value().SubnetPropertiesFormat.RouteTable.ID,"/")
        rtname := tmp[len(tmp)-1]
        routes := self.getNsgRoutes(rtname)
        if (len(routes) == 0) {
            log.Printf("Found subnet %s but will ignore it because it has no route to an nsg\n",*list.Value().SubnetPropertiesFormat.AddressPrefix)
            continue
        }

        as := new(AzureSubnet)
        as.prefix = *list.Value().SubnetPropertiesFormat.AddressPrefix
        as.name = *list.Value().Name
        log.Printf("Found subnet %s\n",as.prefix)

        // We keep these route table names for later when we will want to change next hops. Because we know these routes
        // have an NSG as a next hop, it was checked earlier
        self.routeTables[rtname] = rtname
        nets = append(nets,as)
    }
    return nets
}
