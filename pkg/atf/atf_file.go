// atf is the allocation table format
package atf

import (
	"net"
)

// IPNet represents a net.IPNet, only marshallable
type IPNet struct {
	*net.IPNet `yaml:""`
}

type File struct {
	Name	    *string      `yaml:"name"`
	Superblock  *IPNet       `yaml:"superBlock"`
	Allocations []Allocation `yaml:"allocations"`
}

type Allocation struct {
	IsReserved  bool       `yaml:"reserved,omitempty"`
	Network     *IPNet     `yaml:"cidr"`
	Description string     `yaml:"description"`
	Reference   *Reference `yaml:"ref,omitempty"`
}

type Reference struct {
	DocumentationURI  string `yaml:"documentedAt,omitempty"`
	AWSCloudFormation string `yaml:"awsCF,omitempty"`
	Git               string `yaml:"git,omitempty"`
	SubAlloc          string `yaml:"subAlloc,omitempty"`
}
