package render

import (
	"fmt"
	"io"
	"strings"

	"ops-networking/pkg/atf"
	"ops-networking/pkg/netcalc"
)

var (
	/* MarkerAlloc    = "ALLOC"
	MarkerFree     = "FREE"
	MarkerReserved = "RESERVED" */
	MarkerAlloc    = "✅"
	MarkerFree     = ""
	MarkerReserved = "⚠️"
)

func RenderPoolToMarkdown(target io.Writer, atfFile *atf.File, pool *netcalc.IPNetPool, includeFree bool) {
	ipToAllocMap := make(map[string]atf.Allocation)

	for i := range atfFile.Allocations {
		net := atfFile.Allocations[i].Network.String()
		ipToAllocMap[net] = atfFile.Allocations[i]
	}

	if atfFile.Name == nil {
		addedStr := " "
		if includeFree {
			addedStr += "(with empty blocks)"
		}
		fmt.Fprintf(target, "# %s%s\n\n", atfFile.Superblock.String(), addedStr)
	} else {
		addedStr := ""
		if includeFree {
			addedStr += ", with empty blocks"
		}
		fmt.Fprintf(target, "# %s (%s%s)\n\n", *atfFile.Name, atfFile.Superblock.String(), addedStr)
	}

	fmt.Fprintf(target, "[//]: # (%s)\n\n", "Generated by atfutil, DO NOT EDIT")

	fmt.Fprintf(target, "|%s|%s|%s|%s|\n|-|-|-|-|\n", "Alloc", "Net", "Desc", "Ref")
	for _, block := range pool.FindAllAllocations() {
		if alloc, ok := ipToAllocMap[block.Net.String()]; ok {
			status := MarkerAlloc
			if alloc.IsReserved {
				status = MarkerReserved
			}
			fmt.Fprintf(target, "|%s|%s|%s|%s|\n", status, block.Net.String(), alloc.Description, GetMarkdownRefsByAllocation(alloc))
		} else {
			if includeFree {
				fmt.Fprintf(target, "|%s|%s|%s|%s|\n", MarkerFree, block.Net.String(), "", "")
			}
		}
	}
}

func GetMarkdownRefsByAllocation(alloc atf.Allocation) string {
	strs := make([]string, 0, 2)
	if alloc.Reference != nil {
		if alloc.Reference.SubAlloc != "" {
			strs = append(strs, fmt.Sprintf("[SubAlloc](%s)", alloc.Reference.SubAlloc))
		}
		if alloc.Reference.DocumentationURI != "" {
			strs = append(strs, fmt.Sprintf("[Docs](%s)", alloc.Reference.DocumentationURI))
		}
		if alloc.Reference.AWSCloudFormation != "" {
			strs = append(strs, fmt.Sprintf("[AWS:CF](%s)", alloc.Reference.AWSCloudFormation))
		}
		if alloc.Reference.Git != "" {
			strs = append(strs, fmt.Sprintf("[Git](%s)", alloc.Reference.Git))
		}
	}
	return strings.Join(strs, " ")
}