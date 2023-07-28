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
	"fmt"
	"net"

	"github.com/pkg/errors"
)

func (ipn *IPNet) MarshalText() (text []byte, err error) {
	mask, _ := ipn.Mask.Size()
	retval := fmt.Sprintf("%s/%d", ipn.IP.String(), mask)
	return []byte(retval), nil
}

func (ipn *IPNet) UnmarshalText(text []byte) error {
	ip, ipnet, err := net.ParseCIDR(string(text))
	if err != nil {
		return err
	}
	if !ip.Equal(ipnet.IP) {
		return errors.Errorf("provided non-net CIDR '%s', did you mean '%s'?", text, ipnet.String())
	}
	ipn.IPNet = ipnet
	return nil
}
