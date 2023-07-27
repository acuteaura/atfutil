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

// atf is the allocation table format
package atf

import (
	"net"

	"github.com/pkg/errors"
)

// IPNet represents a net.IPNet, only marshallable
type IPNet struct {
	*net.IPNet `yaml:""`
}

type File struct {
	Name        *string      `yaml:"name"`
	Superblock  *IPNet       `yaml:"superBlock"`
	Allocations []Allocation `yaml:"allocations"`
}

type Allocation struct {
	Ident       string       `yaml:"ident"`
	IsReserved  bool         `yaml:"reserved,omitempty"`
	Network     *IPNet       `yaml:"cidr"`
	Description string       `yaml:"description,omitempty"`
	Reference   Reference    `yaml:"ref,omitempty"`
	SubAlloc    []Allocation `yaml:"subAlloc,omitempty"`
}

type Reference struct {
	AWS   ReferenceAWS   `yaml:"aws,omitempty"`
	Azure ReferenceAzure `yaml:"azure,omitempty"`

	DocumentationURI *string `yaml:"documentedAt,omitempty"`
	Git              *string `yaml:"git,omitempty"`
}

type ReferenceAzure struct {
	Subscription   string `yaml:"subscription,omitempty"`
	ResourceGroup  string `yaml:"resourceGroup,omitempty"`
	VirtualNetwork string `yaml:"virtualNetwork,omitempty"`
}

type ReferenceAWS struct {
	CloudFormationURL string `yaml:"cloudFormationUrl,omitempty"`
}

func (f *File) Validate() error {
	for _, alloc := range f.Allocations {
		for _, subAllocL1 := range alloc.SubAlloc {
			if len(subAllocL1.SubAlloc) >= 1 {
				return errors.Errorf("allocation %s has nested suballocations; this is not supported", alloc.Network.String())
			}
		}
	}
	return nil
}
