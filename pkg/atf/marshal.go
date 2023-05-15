package atf

import (
	"fmt"
	"net"
)

func (ipn *IPNet) MarshalText() (text []byte, err error) {
	mask, _ := ipn.Mask.Size()
	retval := fmt.Sprintf("%s/%d", ipn.IP.String(), mask)
	return []byte(retval), nil
}

func (ipn *IPNet) UnmarshalText(text []byte) error {
	_, ipnet, err := net.ParseCIDR(string(text))
	if err != nil {
		return err
	}
	ipn.IPNet = ipnet
	return nil
}
