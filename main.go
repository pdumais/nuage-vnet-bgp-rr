package main

import (
    "fmt"
    "flag"
    "os"
)


func main() {

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s: -nsg1 <addr> -nsg2 <addr> -routerID <this bgp rid> -las <local as> -ras <remote as>\n", os.Args[0])
        flag.PrintDefaults()
    }

    nsg1 := flag.String("nsg1","192.168.1.216","Address of NSG 1")
    nsg2 := flag.String("nsg2","192.168.1.217","Address of NSG 2")
    routerId := flag.String("rid","1.1.1.1","This router ID")
    las := flag.Uint("las",65242,"Autonomous system")
    ras := flag.Uint("ras",65242,"Autonomous system")
    listenPort := int32(179)

    flag.Parse()

    ctx := new(SessionManagerContext)
    ctx.las = uint32(*las)
    ctx.ras = uint32(*ras)
    ctx.routerId = *routerId
    ctx.server = SetupServer(ctx.las, ctx.routerId, listenPort)
    SetNsgs(ctx,[]string{*nsg1,*nsg2})


    WatchNsgs(ctx)
}
