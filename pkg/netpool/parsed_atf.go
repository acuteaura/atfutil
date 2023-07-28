/*
 * Copyright 2023 Aurelia Schittler
 *
 * Licensed under the EUPL, Version 1.2 or – as soon they
   will be approved by the European Commission - subsequent
   versions of the EUPL (the "Licence");
 * You may not use this work except in compliance with the
   Licence.
 * You may obtain a copy of the Licence at:
 *
 * https://joinup.ec.europa.eu/software/page/eupl5
 *
 * Unless required by applicable law or agreed to in
   writing, software distributed under the Licence is
   distributed on an "AS IS" basis,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
   express or implied.
 * See the Licence for the specific language governing
   permissions and limitations under the Licence.
*/

package netpool

import (
	"atfutil/pkg/atf"
	"atfutil/pkg/netcalc"
	"errors"
	"log"
	"net"
)

type ParsedATF struct {
	File               *atf.File
	Pool               *netcalc.IPNetPool
	subAllocationsPool map[string]*netcalc.IPNetPool
	subAllocationsFile map[string]*atf.Allocation
}

func (patf *ParsedATF) GetPoolByNet(netstring string) *netcalc.IPNetPool {
	pool := patf.subAllocationsPool[netstring]
	return pool
}

func (patf *ParsedATF) GetAtfAllocationByNet(netstring string) *atf.Allocation {
	alloc := patf.subAllocationsFile[netstring]
	return alloc
}

func FromAtf(atfFile *atf.File) (*ParsedATF, error) {
	var subAllocPool map[string]*netcalc.IPNetPool = make(map[string]*netcalc.IPNetPool, 128)
	var subAllocFile map[string]*atf.Allocation = make(map[string]*atf.Allocation, 128)

	parsed, err := fromAllocations(atfFile.Superblock.String(), atfFile.Allocations, 0, subAllocPool, subAllocFile)
	if err != nil {
		return nil, err
	}
	parsed.File = atfFile
	parsed.subAllocationsFile = subAllocFile
	parsed.subAllocationsPool = subAllocPool

	return parsed, nil
}

func fromAllocations(super string, allocations []*atf.Allocation, depth int, subAllocPool map[string]*netcalc.IPNetPool, subAllocFile map[string]*atf.Allocation) (*ParsedATF, error) {
	if depth >= 2 {
		return nil, errors.New("nested suballocations are not supported")
	}
	ipNet := make([]*net.IPNet, 0, len(allocations))
	for _, alloc := range allocations {
		networkName := alloc.Network.String()
		ipNet = append(ipNet, alloc.Network.IPNet)
		if len(alloc.SubAlloc) >= 1 {
			subPool, err := fromAllocations(networkName, alloc.SubAlloc, depth+1, subAllocPool, subAllocFile)
			if err != nil {
				return nil, err
			}
			log.Printf("ASSIGN SUBNET %s\t%v", networkName, alloc)
			subAllocPool[networkName] = subPool.Pool
		}
		log.Printf("ASSIGN NET    %s\t%v", networkName, alloc)
		subAllocFile[networkName] = alloc
	}
	pool, err := netcalc.NewIPNetPool(super, ipNet...)
	if err != nil {
		return nil, err
	}
	return &ParsedATF{
		Pool: pool,
	}, nil
}
