package main

import (
    "fmt"
    "sort"
    "context"
    api "github.com/osrg/gobgp/api"
    bgp "github.com/osrg/gobgp/pkg/packet/bgp"
    gobgp "github.com/osrg/gobgp/pkg/server"
    "github.com/golang/protobuf/ptypes"
)


type SessionManagerContext struct {
    nsgs []*nsg
    primaryNsg *nsg
    server *gobgp.BgpServer
    routerId string
    las uint32
    ras uint32
}

func SetNsgs(ctx *SessionManagerContext, addrs []string) {
    for _,s := range addrs {
        n := new(nsg)
        n.address=s
        n.sessionConnected=false
        ctx.nsgs = append(ctx.nsgs,n)
        AddPeer(ctx.server, s, ctx.ras);
    }
}

func onStateChanged(ctx *SessionManagerContext, peer string, state bgp.FSMState) {
    for _, v := range ctx.nsgs {
        if v.address == peer {
            fmt.Printf("Peer %s state = %v\n",peer,state)
            v.sessionConnected = state == bgp.BGP_FSM_ESTABLISHED
            onNsgsChanged(ctx)
        }
    }
}

func onRoutesChanged(ctx *SessionManagerContext) {
    for _, v := range ctx.nsgs {
        rib,err := ctx.server.ListPath(context.Background(),&api.ListPathRequest{
            Type:   api.Resource_ADJ_IN,
            Family: &api.Family{
                Afi:  api.Family_AFI_IP,
                Safi: api.Family_SAFI_UNICAST,
            },
            Name:   v.address,
        })

        v.pathCount = 0
        v.haPeers = []string{ctx.routerId}
        if err != nil {
            fmt.Printf("Routes for %s can't be found\n",v.address)
            continue
        }

        for _, r := range rib {
            for _, p := range r.Paths {
                isHAPeer := false

                for _, attr := range p.AnyPattrs {
                    var value ptypes.DynamicAny
                    ptypes.UnmarshalAny(attr, &value)
                    switch t :=  value.Message.(type) {
                        case *api.CommunitiesAttribute:
                            if t.Communities[0] == 242 && t.Communities[1] == 242 {
                                isHAPeer = true
                                var nlri api.IPAddressPrefix
                                ptypes.UnmarshalAny(p.AnyNlri,&nlri)
                                v.haPeers = append(v.haPeers,nlri.Prefix)
                            }
                    }
                }

                if !isHAPeer {
                    v.pathCount++
                }
            }
        }

        sort.Strings(v.haPeers)
        fmt.Printf("Routes for %s = %v\n",v.address, v.pathCount)
    }

    onNsgsChanged(ctx)
}

func onNsgsChanged(ctx *SessionManagerContext) {
    fmt.Printf("=================== NSGs ==================\n")

    activeCount := 0
    var primaryNsg *nsg
    for _, v := range ctx.nsgs {
        active := v.sessionConnected && v.pathCount!=0
        isPrimary := v.SetActive(active, activeCount)
        if (isPrimary) {
            primaryNsg = v
        }
        v.Show(ctx)
        if (active) { activeCount++}
    }

    if (ctx.primaryNsg != primaryNsg) {
        onPrimaryNsgChanged(primaryNsg, ctx)
        ctx.primaryNsg = primaryNsg
    }
}

func onPrimaryNsgChanged(nsg *nsg, ctx *SessionManagerContext) {
    if (nsg != nil) {
        fmt.Printf("------> PRIMARY NSG IS NOW %s <------\n",nsg.address)
        if (nsg.IsActiveSpeaker(ctx)) {
            fmt.Printf("Will make a change in Azure\n")
            //TODO: do Azure API call here
        }
    } else {
        fmt.Printf("------> ALL NSGs ARE IN STANDBY <------\n")
    }
}

func WatchNsgs(ctx *SessionManagerContext) {
    w := ctx.server.Watch(gobgp.WatchBestPath(true), gobgp.WatchPeerState(true))
    for {
        select {
        case ev  := <-w.Event():
            switch msg := ev.(type) {
            case *gobgp.WatchEventBestPath:
                fmt.Printf("Received Best Path Event\n")
                onRoutesChanged(ctx)
            case *gobgp.WatchEventPeerState:
                fmt.Printf("Received Peer State Event\n")
                onStateChanged(ctx, msg.PeerAddress.String(),msg.State)
            }
        }
    }
}
