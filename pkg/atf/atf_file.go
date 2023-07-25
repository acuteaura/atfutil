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
