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

package atf

import (
	"net"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func CIDR(cidr string) *net.IPNet {
	_, net, err := net.ParseCIDR(cidr)
	if err != nil {
		panic("CIDR parsing failed")
	}
	// allow wrongness
	return net
}

func TestIPNet_MarshalUnmarshalText(t *testing.T) {
	atf := &File{
		Superblock: &IPNet{CIDR("10.42.0.0/16")},
		Allocations: []*Allocation{
			{
				Network:     &IPNet{CIDR("10.42.0.0/15")},
				Description: "half the slice",
			},
		},
	}
	bytes, err := yaml.Marshal(&atf)
	if err != nil {
		t.Fatal(err)
	}

	atf2 := &File{}
	err = yaml.Unmarshal(bytes, atf2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(atf, atf2) {
		t.Fatal("deepEqual failed")
	}
}

func TestIPNet_MarshalUnmarshalTextBadNet(t *testing.T) {
	atf := `superBlock: 10.99.0.0/16
allocations:
- cidr: 10.99.43.0/23
	`

	atf2 := &File{}
	err := yaml.Unmarshal([]byte(atf), atf2)

	if err == nil {
		t.Fatal("unmarshalled bad data")
	}
}
