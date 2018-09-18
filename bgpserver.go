package main

import (
    "context"
    "log"
    gobgp "github.com/osrg/gobgp/pkg/server"
    api "github.com/osrg/gobgp/api"
    logrus "github.com/sirupsen/logrus"
    "github.com/golang/protobuf/ptypes"
    "github.com/golang/protobuf/ptypes/any"
)

func SetupServer(as uint32, routerId string, listenPort int32) (*gobgp.BgpServer){
    log.Printf("Starting BGP Client")
    logrus.SetLevel(logrus.DebugLevel)
    s := gobgp.NewBgpServer()
    go s.Serve()
    if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
        Global: &api.Global{
            As:         as,
            RouterId:   routerId,
            ListenPort: listenPort,
        },
    }); err != nil {
        log.Fatal(err)
    }

    // We add a dummy path with a specific community ID in order to advertise our presence to other HA instances of this app
    a1, _ := ptypes.MarshalAny(&api.OriginAttribute{ Origin: 0, })
    a2, _ := ptypes.MarshalAny(&api.CommunitiesAttribute{ Communities: []uint32{242, 242}, })
    a3, _ := ptypes.MarshalAny(&api.NextHopAttribute{ NextHop: "10.0.0.1", })
    attrs := []*any.Any{a1,a2,a3}
    nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
        Prefix:    routerId,
        PrefixLen: 32,
    })
    s.AddPath(context.Background(), &api.AddPathRequest{
        Path: &api.Path{
            Family:    &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
            AnyNlri:   nlri,
            AnyPattrs: attrs,
        },
    })

    return s
}

func AddPeer(server *gobgp.BgpServer, address string, as uint32) {
    conf := &api.PeerConf{
            NeighborAddress: address,
            PeerAs:          as,
        }

    peer := &api.Peer{
        Conf: conf,
        Transport: &api.Transport{
            PassiveMode: true,      // we don't want to connect to that peer. It will connect to us
        },
    }
    server.AddPeer(context.Background(), &api.AddPeerRequest{
        Peer: peer,
    })
}
