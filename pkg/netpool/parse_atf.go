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
	"net"
)

type IPNetPool struct {
	Pool           *netcalc.IPNetPool
	SubAllocations map[string]*netcalc.IPNetPool
}

func FromAtf(atfFile *atf.File) (*IPNetPool, error) {
	return FromAllocations(atfFile.Superblock.String(), atfFile.Allocations, 0)
}

func FromAllocations(super string, allocations []atf.Allocation, depth int) (*IPNetPool, error) {
	var subAlloc map[string]*netcalc.IPNetPool
	if depth >= 2 {
		return nil, errors.New("nested suballocations are not supported")
	}
	ipNet := make([]*net.IPNet, 0, len(allocations))
	for _, alloc := range allocations {
		ipNet = append(ipNet, alloc.Network.IPNet)
		if len(alloc.SubAlloc) >= 1 {
			subPool, err := FromAllocations(alloc.Network.String(), alloc.SubAlloc, depth+1)
			if err != nil {
				return nil, err
			}
			if subAlloc == nil {
				subAlloc = make(map[string]*netcalc.IPNetPool, 128)
			}
			subAlloc[alloc.Network.String()] = subPool.Pool
		}
	}
	pool, err := netcalc.NewIPNetPool(super, ipNet...)
	if err != nil {
		return nil, err
	}
	return &IPNetPool{
		Pool:           pool,
		SubAllocations: subAlloc,
	}, nil
}
