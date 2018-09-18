package main

import (
    "log"
)


type nsg struct {
    address string
    active bool
    sessionConnected bool
    pathCount uint64
    haPeers []string
    primary bool
}

func (self *nsg) SetActive(active bool, activeCountBeforeMe int) (bool){
    self.active = active
    primary := (active && activeCountBeforeMe==0)

    self.primary = primary
    return primary
}

func (self *nsg) IsActiveSpeaker(ctx *SessionManagerContext) (bool){
    return self.haPeers[0] == ctx.routerId
}

func (self *nsg) Show(ctx *SessionManagerContext) {
    log.Printf("NSG %s:\n",self.address)
    log.Printf("    Considered Primary:  %v\n",self.active)
    log.Printf("    Num Paths:          %v\n",self.pathCount)
    log.Printf("    BGP Apps:\n")
    for i, haPeer := range self.haPeers {
        var attrs []string
        if (i == 0) {
            attrs = append(attrs,"Active Speaker")
        } else {
            attrs = append(attrs,"Standby Speaker")
        }
        if (haPeer == ctx.routerId) {
            attrs = append(attrs,"This instance")
        }
        log.Printf("        %s ",haPeer)
        log.Printf("%v",attrs)
        log.Printf("\n")
    }
}
