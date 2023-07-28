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

package render

import (
	"atfutil/pkg/atf"
	"atfutil/pkg/netpool"
	"fmt"
	"io"
	"log"
)

type DefaultAzureMarkdownRenderer struct {
	target io.Writer
	fmtStr string
}

const (
	/* MarkerAlloc    = "ALLOC"
	MarkerFree     = "FREE"
	MarkerReserved = "RESERVED" */
	AzureMarkerAlloc    = "✅"
	AzureMarkerFree     = ""
	AzureMarkerReserved = "⚠️"
)

func NewDefaultAzureMarkdownRenderer(target io.Writer) *DefaultAzureMarkdownRenderer {
	return &DefaultAzureMarkdownRenderer{
		fmtStr: "|%s|%s|%s|%s|%s|%s|%s|%s|\n",
		target: target,
	}
}

func (dmr *DefaultAzureMarkdownRenderer) PrintHeader() {
	fmt.Fprintf(dmr.target, dmr.fmtStr, "Alloc", "Ident", "Block", "SubAllocation", "Subscription", "Resource Group", "VNET", "Description")
	fmt.Fprintf(dmr.target, dmr.fmtStr, "-", "-", "-", "-", "-", "-", "-", "-")
}

func (dmr *DefaultAzureMarkdownRenderer) PrintAllocation(parsed *netpool.ParsedATF, netstring string, isSub bool, includeFree bool) {
	//isFree := true
	alloc := parsed.GetAtfAllocationByNet(netstring)
	status := AzureMarkerAlloc

	if alloc == nil {
		log.Printf("ALLOC %s is empty", netstring)
		alloc = &atf.Allocation{}
		status = AzureMarkerFree
	}

	if alloc.IsReserved {
		status = AzureMarkerReserved
	}

	var textBlock = netstring
	var textSubAlloc = ""

	if isSub {
		textSubAlloc = textBlock
		textBlock = ""
	}
	fmt.Fprintf(
		dmr.target,
		dmr.fmtStr,
		status,
		alloc.Ident,
		textBlock,
		textSubAlloc,
		alloc.Reference.Azure.Subscription,
		alloc.Reference.Azure.ResourceGroup,
		alloc.Reference.Azure.VirtualNetwork,
		alloc.Description,
	)

	// we get the (sub) pool here instead of iterating over the actual SubAllocations
	// from the ATF so we can let FindAllAllocations add free space
	if subBlock := parsed.GetPoolByNet(netstring); subBlock != nil {
		for _, block := range subBlock.FindAllAllocations() {
			dmr.PrintAllocation(parsed, block.Net.String(), true, includeFree)
		}
	}
}
