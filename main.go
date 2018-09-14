package main

import (
    "fmt"
    "flag"
    "os"
)


func main() {

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s: -c | -vm=/path/to/vmdisk.qcow2 ...\n", os.Args[0])
        flag.PrintDefaults()
    }

    nsg1 := flag.String("nsg1","192.168.1.216","Address of NSG 1")
    routerId := flag.String("rid","1.1.1.1","This router ID")
    as := flag.Uint("as",65242,"Autonomous system")
    listenPort := int32(179)

    flag.Parse()

    ctx := new(SessionManagerContext)
    ctx.as = uint32(*as)
    ctx.server = SetupServer(ctx.as,*routerId,listenPort)
    SetNsgs(ctx,[]string{*nsg1})
    WatchNsgs(ctx)
}
