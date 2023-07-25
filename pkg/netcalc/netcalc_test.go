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
	"fmt"
	"net"
	"reflect"
	"testing"
)

func TestLegalNetworkSizes(t *testing.T) {
	tests := []struct {
		ip   net.IP
		want []net.IPMask
	}{
		{
			net.ParseIP("10.0.0.0"),
			[]net.IPMask{
				net.CIDRMask(7, 32),
				net.CIDRMask(8, 32),
				net.CIDRMask(9, 32),
				net.CIDRMask(10, 32),
				net.CIDRMask(11, 32),
				net.CIDRMask(12, 32),
				net.CIDRMask(13, 32),
				net.CIDRMask(14, 32),
				net.CIDRMask(15, 32),
				net.CIDRMask(16, 32),
				net.CIDRMask(17, 32),
				net.CIDRMask(18, 32),
				net.CIDRMask(19, 32),
				net.CIDRMask(20, 32),
				net.CIDRMask(21, 32),
				net.CIDRMask(22, 32),
				net.CIDRMask(23, 32),
				net.CIDRMask(24, 32),
				net.CIDRMask(25, 32),
				net.CIDRMask(26, 32),
				net.CIDRMask(27, 32),
				net.CIDRMask(28, 32),
				net.CIDRMask(29, 32),
				net.CIDRMask(30, 32),
				net.CIDRMask(31, 32),
				net.CIDRMask(32, 32),
			},
		},
		{
			net.ParseIP("10.0.0.16"),
			[]net.IPMask{
				net.CIDRMask(28, 32),
				net.CIDRMask(29, 32),
				net.CIDRMask(30, 32),
				net.CIDRMask(31, 32),
				net.CIDRMask(32, 32),
			},
		},
		{
			net.ParseIP("10.0.0.8"),
			[]net.IPMask{
				net.CIDRMask(29, 32),
				net.CIDRMask(30, 32),
				net.CIDRMask(31, 32),
				net.CIDRMask(32, 32),
			},
		},
		{
			net.ParseIP("10.0.0.32"),
			[]net.IPMask{
				net.CIDRMask(27, 32),
				net.CIDRMask(28, 32),
				net.CIDRMask(29, 32),
				net.CIDRMask(30, 32),
				net.CIDRMask(31, 32),
				net.CIDRMask(32, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.ip.String(), func(t *testing.T) {
			minSize := net.CIDRMask(0, 32)
			if got := LegalNetworkSizes(tt.ip, minSize); !reflect.DeepEqual(got, tt.want) {
				gotCidr := []int{}
				for _, elementGotten := range got {
					cidrBits, _ := elementGotten.Size()
					gotCidr = append(gotCidr, cidrBits)
				}
				t.Errorf("LegalNetworkSizes() = %v, want %v", gotCidr, tt.want)
			}
		})
	}
}

func CIDR(cidr string) *net.IPNet {
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		panic("CIDR parsing failed")
	}
	if !net.IP.Equal(ip) {
		panic("CIDR is not a network address")
	}
	return net
}

func PrintBlocks(t *testing.T, blocks []*Block) {
	for _, block := range blocks {
		state := "FREE"
		if block.Alloc {
			state = "ALLOC"
		}
		t.Logf("%s\t%s\n", state, block.Net.String())
	}
}

func TestAllocations1(t *testing.T) {
	pool, err := NewIPNetPool("10.42.0.0/24",
		CIDR("10.42.0.16/28"),
		CIDR("10.42.0.64/28"),
		CIDR("10.42.0.0/30"),
	)
	if err != nil {
		t.Fatal(err)
	}
	blocks := pool.FindAllAllocations()

	numExpectedBlocks := 9

	if len(blocks) != numExpectedBlocks {
		t.Fatalf("got %d blocks, expected %d", len(blocks), numExpectedBlocks)
	}
	expectedBlocks := []*Block{
		{Alloc: true, Net: CIDR("10.42.0.0/30")},
		{Alloc: false, Net: CIDR("10.42.0.4/30")},
		{Alloc: false, Net: CIDR("10.42.0.8/29")},
		{Alloc: true, Net: CIDR("10.42.0.16/28")},
		{Alloc: false, Net: CIDR("10.42.0.32/27")},
		{Alloc: true, Net: CIDR("10.42.0.64/28")},
		{Alloc: false, Net: CIDR("10.42.0.80/28")},
		{Alloc: false, Net: CIDR("10.42.0.96/27")},
		{Alloc: false, Net: CIDR("10.42.0.128/25")},
	}
	if !reflect.DeepEqual(blocks, expectedBlocks) {
		t.Log("blocks aren't as expected")
		t.Log("expected:")
		PrintBlocks(t, expectedBlocks)
		t.Log("actual: ")
		PrintBlocks(t, blocks)
		t.Fail()
	}
}

func TestAllocations2(t *testing.T) {
	pool, err := NewIPNetPool("10.0.0.0/8",
		CIDR("10.0.0.0/16"),
		CIDR("10.1.0.0/16"),
		CIDR("10.2.0.0/15"),
	)
	if err != nil {
		t.Fatal(err)
	}
	blocks := pool.FindAllAllocations()

	numExpectedBlocks := 9

	if len(blocks) != numExpectedBlocks {
		t.Fatalf("got %d blocks, expected %d", len(blocks), numExpectedBlocks)
	}
	expectedBlocks := []*Block{
		{Alloc: true, Net: CIDR("10.0.0.0/16")},
		{Alloc: true, Net: CIDR("10.1.0.0/16")},
		{Alloc: true, Net: CIDR("10.2.0.0/15")},
		{Alloc: false, Net: CIDR("10.4.0.0/14")},
		{Alloc: false, Net: CIDR("10.8.0.0/13")},
		{Alloc: false, Net: CIDR("10.16.0.0/12")},
		{Alloc: false, Net: CIDR("10.32.0.0/11")},
		{Alloc: false, Net: CIDR("10.64.0.0/10")},
		{Alloc: false, Net: CIDR("10.128.0.0/9")},
	}
	if !reflect.DeepEqual(blocks, expectedBlocks) {
		t.Log("blocks aren't as expected")
		t.Log("expected:")
		PrintBlocks(t, expectedBlocks)
		t.Log("actual: ")
		PrintBlocks(t, blocks)
		t.Fail()
	}
}

func TestAllocationBadSuperblock(t *testing.T) {
	_, err := NewIPNetPool("10.42.0.1/24",
		CIDR("10.42.0.128/25"),
	)
	if err == nil {
		t.Fatal("should have errored")
	}
}

func TestAllocationJunkSuperblock(t *testing.T) {
	_, err := NewIPNetPool("hello there",
		CIDR("10.42.0.128/25"),
	)
	if err == nil {
		t.Fatal("should have errored")
	}
}

func TestAllocationBadSpec(t *testing.T) {
	_, err := NewIPNetPool("127.0.0.1/32",
		CIDR("10.10.10.0/24"),
	)
	if err == nil {
		t.Fatal("should have errored")
	}
}

func TestAllocationBadSpec2(t *testing.T) {
	_, err := NewIPNetPool("10.42.0.0/24",
		CIDR("10.42.0.0/26"),
		CIDR("10.42.0.64/26"),
		CIDR("10.42.0.0/25"),
	)
	if err == nil {
		t.Fatal("should have errored")
	}
}

func TestAllocationBadSpec3(t *testing.T) {
	_, err := NewIPNetPool("10.42.0.0/24",
		CIDR("10.42.0.0/26"),
		CIDR("10.42.0.64/26"),
		CIDR("10.42.0.16/30"),
	)
	if err == nil {
		t.Fatal("should have errored")
	}
}

func TestIPNetPool_Alloc(t *testing.T) {
	pool, err := NewIPNetPool("10.0.0.0/8",
		CIDR("10.0.0.0/16"),
		CIDR("10.1.0.0/16"),
		CIDR("10.2.0.0/15"),
	)
	if err != nil {
		t.Fatal(err)
	}

	for i := 4; i < 104; i++ {
		net, err := pool.Alloc(16)
		if err != nil {
			t.Fatal(err)
		}
		expectedNet := fmt.Sprintf("10.%d.0.0/16", i)
		if net.String() != fmt.Sprintf("10.%d.0.0/16", i) {
			t.Fatalf("expected net %s, got %s", expectedNet, net.String())
		}
	}
}

func TestIPNetPool_Alloc2(t *testing.T) {
	pool, err := NewIPNetPool("10.0.0.0/8",
		CIDR("10.42.0.0/24"),
		CIDR("10.42.1.128/25"),
		CIDR("10.42.2.128/25"),
		CIDR("10.0.0.0/16"),
		CIDR("10.1.0.0/16"),
		CIDR("10.2.0.0/15"),
	)
	if err != nil {
		t.Fatal(err)
	}

	net, err := pool.Alloc(25)
	if err != nil {
		t.Fatal(err)
	}
	net2, err := pool.Alloc(25)
	if err != nil {
		t.Fatal(err)
	}

	expectedNet := "10.42.1.0/25"
	if net.String() != expectedNet {
		t.Fatalf("expected net %s, got %s", expectedNet, net.String())
	}

	expectedNet = "10.42.2.0/25"
	if net2.String() != expectedNet {
		t.Fatalf("expected net %s, got %s", expectedNet, net2.String())
	}
}

func TestIPNetPool_AllocLargerThanSupper(t *testing.T) {
	pool, err := NewIPNetPool("10.0.0.0/12")
	if err != nil {
		t.Fatal(err)
	}

	_, err = pool.Alloc(8)
	if err == nil {
		t.Fatal("should have failed to allocate net larger than superblock")
	}
}
