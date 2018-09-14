package main

import (
    "context"
    gobgp "github.com/osrg/gobgp/pkg/server"
    api "github.com/osrg/gobgp/api"
    log "github.com/sirupsen/logrus"
)

func SetupServer(as uint32, routerId string, listenPort int32) (*gobgp.BgpServer){
    log.SetLevel(log.DebugLevel)
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
