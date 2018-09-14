package main

import (
    "fmt"
    "context"
    api "github.com/osrg/gobgp/api"
    bgp "github.com/osrg/gobgp/pkg/packet/bgp"
    gobgp "github.com/osrg/gobgp/pkg/server"
)


type nsg struct {
    address string
    active bool
    pathCount uint64
}

type SessionManagerContext struct {
    nsgs []*nsg
    server *gobgp.BgpServer
    as uint32
}

func SetNsgs(ctx *SessionManagerContext, addrs []string) {
    for _,s := range addrs {
        n := new(nsg)
        n.address=s
        n.active=false
        ctx.nsgs = append(ctx.nsgs,n)
        AddPeer(ctx.server, s, ctx.as);
    }
}

func onStateChanged(ctx *SessionManagerContext, peer string, state bgp.FSMState) {
    for _, v := range ctx.nsgs {
        if v.address == peer {
            fmt.Printf("Peer %s state = %v\n",peer,state)
            v.active = state == bgp.BGP_FSM_ESTABLISHED
            onNsgsChanged(ctx)
        }
    }
}



func onRoutesChanged(ctx *SessionManagerContext) {
    for _, v := range ctx.nsgs {
        p,err := ctx.server.GetTable(context.Background(),&api.GetTableRequest{
            Type:   api.Resource_ADJ_IN,
            Family: &api.Family{
                Afi:  api.Family_AFI_IP,
                Safi: api.Family_SAFI_UNICAST,
            },
            Name:   v.address,
        })

        if err != nil {
            fmt.Printf("Routes for %s can't be found\n",v.address)
            v.pathCount = 0
            continue
        }

        v.pathCount = p.NumPath
        fmt.Printf("Routes for %s = %v\n",v.address, p.NumPath)
    }

    onNsgsChanged(ctx)
}

func onNsgsChanged(ctx *SessionManagerContext) {
    for _, v := range ctx.nsgs {
        fmt.Printf("Peer %s is active: %v\n",v.address,(v.active && v.pathCount!=0))
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
