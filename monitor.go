package main

import (
    "log"
    "time"
    "strconv"
    "strings"
    "net"
    api "github.com/osrg/gobgp/api"
    "github.com/golang/protobuf/ptypes"
    "github.com/golang/protobuf/ptypes/any"
)

func getNsgSubnetGateway(ctx *SessionManagerContext, nsgip string) (string,string) {
    for _, sub := range ctx.azure.GetSubnets() {
        _,ipnet,_ := net.ParseCIDR(sub.prefix)
        ip := net.ParseIP(nsgip)
        if ipnet.Contains(ip) {
            netaddr := strings.Split(sub.prefix,"/")[0]
            b:=strings.Split(netaddr,".")
            inc,_ := strconv.Atoi(b[3])
            gw := b[0]+"."+b[1]+"."+b[2]+"."+strconv.Itoa(inc+1)
            return gw,sub.prefix
        }
    }
    return "",""
}

func Monitor(ctx *SessionManagerContext, nsgips []string) {
    ticker1 := time.NewTicker(30*time.Second)
    go processTicker(ctx,ticker1)

    SetNsgs(ctx,nsgips)
    for _,nsgip := range nsgips {
        gw,prefix := getNsgSubnetGateway(ctx, nsgip)
        ctx.nsgs[nsgip].gateway = gw
        ctx.nsgs[nsgip].lanprefix = prefix
    }

    updateRIB(ctx)
    WatchNsgs(ctx)
}

func processTicker(ctx *SessionManagerContext, ticker *time.Ticker) {
    for range ticker.C {
        updateRIB(ctx)
        log.Printf("ticker \n")
    }
}

func updateRIB(ctx *SessionManagerContext) {
    var nsg *nsg
    for _,n := range ctx.nsgs {
        if n.primary {
            nsg = n
            break
        }
    }
    if (nsg == nil) {
        return
    }

    // Get the list of currently advertised prefixes so we can build a delta
    subnetsOut := make(map[string]string)
    rib,_ := ctx.server.ListPath(ctx.context,&api.ListPathRequest{
        Type:   api.Resource_ADJ_OUT,
        Family: &api.Family{
            Afi:  api.Family_AFI_IP,
            Safi: api.Family_SAFI_UNICAST,
        },
        Name:   nsg.address,
    })
    for _,r := range rib {
        for _, p := range r.Paths {
            var nlri api.IPAddressPrefix
            ptypes.UnmarshalAny(p.AnyNlri,&nlri)
            var gw string
            ignore := false
            for _, attr := range p.AnyPattrs {
                var value ptypes.DynamicAny
                ptypes.UnmarshalAny(attr, &value)
                switch t :=  value.Message.(type) {
                    case *api.CommunitiesAttribute:
                        ignore = true
                    case *api.NextHopAttribute:
                        gw = t.NextHop
                }
            }
            if (!ignore) {
                subnetsOut[nlri.Prefix+"/"+strconv.Itoa(int(nlri.PrefixLen))] = gw
            }
        }
    }

    // Get the subnets from Azure and advertise them
    azureSubnets := ctx.azure.GetSubnets()
    for _, sub := range azureSubnets {
        ipnet := strings.Split(sub.prefix,"/")
        net := ipnet[0]
        s,_ := strconv.Atoi(ipnet[1])

        found := false
        for prefix,gw := range subnetsOut {
            if (prefix == sub.prefix && gw == nsg.gateway) {
                found = true
                break
            }
        }

        if (!found) {
            nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
                Prefix:    net,
                PrefixLen: uint32(s),
            })


            a1, _ := ptypes.MarshalAny(&api.OriginAttribute{ Origin: 0, })
            a2, _ := ptypes.MarshalAny(&api.NextHopAttribute{ NextHop: nsg.gateway, })
            attrs := []*any.Any{a1, a2}

            ctx.server.AddPath(ctx.context, &api.AddPathRequest{
                Path: &api.Path{
                    Family:    &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
                    AnyNlri:   nlri,
                    AnyPattrs: attrs,
                },
            })
        }
    }

    // Now for all subnets that were in rib but not in azure anymore, remove them from rib
    // TODO !!
    /*
    for prefix,gw := range subnetsOut {
        found := false
        for _, sub := range azureSubnets {
            if (sub.prefix == prefix && 
        }
    }*/
}

func onPrimaryNsgChanged(nsg *nsg, ctx *SessionManagerContext) {
    if (nsg != nil) {
        log.Printf("------> PRIMARY NSG IS NOW %s <------\n",nsg.address)
        if (nsg.IsActiveSpeaker(ctx)) {
            ctx.azure.ChangeUplink(nsg.address)
            updateRIB(ctx)
        }
    } else {
        log.Printf("------> ALL NSGs ARE IN STANDBY <------\n")
    }
}
