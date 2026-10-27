package test

import (
	"testing"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"github.com/seancfoley/ipaddress-go/ipaddr/addrstr"
)

func TestMixedStringPreservesPrefixedRanges(t *testing.T) {
	for _, input := range []string{
		"1::fff:*:*/63",
		"1::*:*/63",
		"1:*::fff:*:*/16",
		"1::fff:1-2:*/63",
		"1::fff:*:*",
		"1::/64",
		"1::a:*/112",
	} {
		t.Run(input, func(t *testing.T) {
			original := ipaddr.NewIPAddressString(input).GetAddress().ToIPv6()
			custom := new(addrstr.IPv6StringOptionsBuilder).SetMixed(true).ToOptions()
			mixed, err := original.ToMixedString()
			if err != nil {
				t.Fatal(err)
			}
			normalized, err := original.ToCustomString(custom)
			if err != nil {
				t.Fatal(err)
			}
			grouping, err := original.GetMixedAddressGrouping()
			if err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{mixed, normalized, grouping.String()} {
				parsed := ipaddr.NewIPAddressString(text).GetAddress()
				if parsed == nil || !original.Equal(parsed.ToIPv6()) || !original.GetPrefixLen().Equal(parsed.GetPrefixLen()) {
					t.Errorf("%s became %s", original, text)
				}
			}
		})
	}
}
