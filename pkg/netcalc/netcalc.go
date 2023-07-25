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

package netcalc

import (
	"encoding/binary"
	"errors"
	"net"
	"sort"

	"atfutil/pkg/cidr"
)

const AWS_MIN_SUBNET_SIZE = 28
const MASK_BITS = 32

var (
	// ErrRootNotNetworkAddr indicates the given root address is not a network address of given CIDR
	ErrRootNotNetworkAddr = errors.New("netcalc: given root address is not network address of given CIDR")

	// ErrAllocationOutOfBounds indicates at least one of the given allocations is out of bounds of root network
	ErrAllocationOutOfBounds = errors.New("netcalc: given allocations out of bounds of root network")
)

// IPNetPool is a list of allocations in one larger superblock
type IPNetPool struct {
	super *net.IPNet
	alloc []*net.IPNet
}

// Block is an IP network that is either allocated or free within a superblock
type Block struct {
	Net   *net.IPNet
	Alloc bool
}

// Len is the number of elements in the collection.
func (ipnp *IPNetPool) Len() int {
	return len(ipnp.alloc)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ipnp *IPNetPool) Less(i, j int) bool {
	return binary.BigEndian.Uint32(ipnp.alloc[i].IP) < binary.BigEndian.Uint32(ipnp.alloc[j].IP)
}

// Swap swaps the elements with indexes i and j.
func (ipnp *IPNetPool) Swap(i, j int) {
	jPtr := ipnp.alloc[i]
	ipnp.alloc[i] = ipnp.alloc[j]
	ipnp.alloc[j] = jPtr
}

func (ipnp *IPNetPool) Alloc(requestedSize int) (*net.IPNet, error) {
	allocBlockAt := -1
	currentSize := -1
	blocks := ipnp.FindAllAllocations()

	for i, block := range blocks {
		if block.Alloc {
			continue
		}
		blockSize, _ := block.Net.Mask.Size()
		if allocBlockAt == -1 && blockSize <= requestedSize {
			allocBlockAt = i
			currentSize = blockSize
			continue
		}

		if blockSize <= requestedSize && blockSize > currentSize {
			allocBlockAt = i
			currentSize = blockSize
		}
	}
	if allocBlockAt == -1 {
		return nil, errors.New("no space to allocate a subnet of the requested size")
	}
	ipnet := blocks[allocBlockAt].Net
	// downsize the network
	ipnet.Mask = net.CIDRMask(requestedSize, MASK_BITS)
	ipnp.alloc = append(ipnp.alloc, ipnet)

	err := ipnp.fixAndVerifyInternalState()
	if err != nil {
		panic("internal state of the ipnetpool might be messed up, fixme please")
	}

	return ipnet, nil
}

// FindAllAllocations calculates all Blocks in the IPNetPool. Unallocated
// Space is reduced to the largest allocatable blocks.
func (ipnp *IPNetPool) FindAllAllocations() []*Block {
	blocks := make([]*Block, 0, MASK_BITS)

	// this is the IP we're currently trying to find a matching free or allocated
	// network for
	var currentIP net.IP
	currentIP = ipnp.super.IP

	// because allocations are sorted we only ever need to match against the
	// next allocation, this is the index that allocation.
	allocIndex := 0
	var alloc *net.IPNet

	// we build a reusable slice of the allocations and one entry for our
	// network to see if it fits in.
	checkSlc := append(ipnp.alloc, nil)
	checkSliWorkIndex := len(checkSlc) - 1

mainloop:
	for {
		// a logic error might leave us never moving forward, detect this
		var didIterate bool

		// iterate over all possible subnet sizes our current IP can support,
		// largest network (lowest number) to smallest network (highest number)
		for _, netSize := range LegalNetworkSizes(currentIP, ipnp.super.Mask) {
			// have we found all allocated blocks?
			if len(ipnp.alloc) > allocIndex {
				alloc = ipnp.alloc[allocIndex]
			} else {
				alloc = nil
			}
			// populate our working pointer
			checkSlc[checkSliWorkIndex] = &net.IPNet{
				IP:   currentIP,
				Mask: netSize,
			}

			// check if the current net is the allocation
			if alloc != nil && alloc.IP.Equal(currentIP) && alloc.Mask.String() == netSize.String() {
				blocks = append(blocks, &Block{
					Net:   alloc,
					Alloc: true,
				})
				allocIndex++
				prefixLen, _ := alloc.Mask.Size()
				nextNet, _ := cidr.NextSubnet(checkSlc[checkSliWorkIndex], prefixLen)
				currentIP = nextNet.IP
				didIterate = true
				break
			}

			// check if the current net does not overlap with any other net
			if err := cidr.VerifyNoOverlap(checkSlc, ipnp.super); err == nil {
				// found a free block
				blocks = append(blocks, &Block{
					Net:   checkSlc[checkSliWorkIndex],
					Alloc: false,
				})
				prefixLen, _ := checkSlc[checkSliWorkIndex].Mask.Size()
				nextNet, _ := cidr.NextSubnet(checkSlc[checkSliWorkIndex], prefixLen)
				currentIP = nextNet.IP
				didIterate = true
				break
			}
		}
		// break out if we stepped out of bounds
		if !ipnp.super.Contains(currentIP) {
			break mainloop
		}
		if !didIterate {
			panic("error in logic, panicing to avoid endless loop")
		}
	}

	return blocks
}

func (ipnp *IPNetPool) fixAndVerifyInternalState() error {
	sort.Sort(ipnp)

	err := cidr.VerifyNoOverlap(ipnp.alloc, ipnp.super)
	if err != nil {
		return err
	}
	return nil
}

// NewIPNetPool creates a new pool of IP allocation blocks within one superblock
func NewIPNetPool(superCidr string, allocations ...*net.IPNet) (*IPNetPool, error) {
	ip, super, err := net.ParseCIDR(superCidr)
	if err != nil {
		return nil, err
	}
	if !super.IP.Equal(ip) {
		return nil, ErrRootNotNetworkAddr
	}

	ipnp := &IPNetPool{super, allocations}

	err = ipnp.fixAndVerifyInternalState()
	if err != nil {
		return nil, err
	}

	return ipnp, nil
}

func LegalNetworkSizes(ip net.IP, min net.IPMask) []net.IPMask {
	masks := make([]net.IPMask, 0, 1)
	minMask, _ := min.Size()
	for i := minMask + 1; i <= MASK_BITS; i++ {
		mask := net.CIDRMask(i, MASK_BITS)
		if ip.Equal(ip.Mask(mask)) {
			masks = append(masks, mask)
		}
	}
	return masks
}
