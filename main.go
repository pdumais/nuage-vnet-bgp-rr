package main

import (
    "context"
    "log"
    "flag"
    "os"
)


func main() {
    log.SetPrefix("[APP] --> ")

    flag.Usage = func() {


        log.Printf("Usage of %s:",os.Args[0])
        flag.PrintDefaults()
    }

    nsg1 := flag.String("nsg1","192.168.1.216","Address of NSG 1")
    nsg2 := flag.String("nsg2","192.168.1.217","Address of NSG 2")
    routerId := flag.String("rid","1.1.1.1","This router ID")
    las := flag.Uint("las",65242,"Local autonomous system ID")
    ras := flag.Uint("ras",65242,"NSG autonomous system ID")
    subscriptionId := flag.String("subscriptionid","","Azure subscription ID")
    clientId := flag.String("clientid","","Azure client ID")
    password := flag.String("password","","Azure password")
    tenantId := flag.String("tenantid","","Azure tenant ID")
    rgroup := flag.String("rgroup","","Azure resource group name")
    vnet := flag.String("vnet","","Azure vnet name")
    listenPort := int32(179)

    flag.Parse()

    nsgs := []string{*nsg1,*nsg2}
    ctx := new(SessionManagerContext)
    ctx.context = context.Background()
    ctx.azure = NewAzure(*subscriptionId,*clientId,*password,*tenantId,*rgroup,*vnet,ctx.context, nsgs)

    ctx.las = uint32(*las)
    ctx.ras = uint32(*ras)
    ctx.routerId = *routerId
    ctx.server = SetupServer(ctx.las, ctx.routerId, listenPort)
    StartStats(ctx)
    Monitor(ctx,nsgs)
}
