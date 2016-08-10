package mojanet

import (
	"testing"

	"github.com/vishvananda/netlink"
)

// TestValidateLink tests the parentExists function
func TestValidateLink(t *testing.T) {
	validIface := "lo"
	invalidIface := "foo12345"

	// test a valid parent interface validation
	if ok := parentExists(validIface); !ok {
		t.Fatalf("failed validating loopback %s", validIface)
	}
	// test a invalid parent interface validation
	if ok := parentExists(invalidIface); ok {
		t.Fatalf("failed to invalidate interface %s", invalidIface)
	}
}

// TestValidateSubLink tests valid 802.1q naming convention
func TestValidateSubLink(t *testing.T) {
	validSubIface := "lo.10"
	invalidSubIface1 := "lo"
	invalidSubIface2 := "lo:10"
	invalidSubIface3 := "foo123.456"

	// test a valid parent_iface.vlan_id
	_, _, err := parseVlan(validSubIface)
	if err != nil {
		t.Fatalf("failed subinterface validation: %v", err)
	}
	// test a invalid vid with a valid parent link
	_, _, err = parseVlan(invalidSubIface1)
	if err == nil {
		t.Fatalf("failed subinterface validation test: %s", invalidSubIface1)
	}
	// test a valid vid with a valid parent link with a invalid delimiter
	_, _, err = parseVlan(invalidSubIface2)
	if err == nil {
		t.Fatalf("failed subinterface validation test: %v", invalidSubIface2)
	}
	// test a invalid parent link with a valid vid
	_, _, err = parseVlan(invalidSubIface3)
	if err == nil {
		t.Fatalf("failed subinterface validation test: %v", invalidSubIface3)
	}
}

// TestSetMojanetMode tests the mojanet mode setter
func TestSetMojanetMode(t *testing.T) {
	// test mojanet bridge mode
	mode, err := setMojanetMode(modeBridge)
	if err != nil {
		t.Fatalf("error parsing %v vlan mode: %v", mode, err)
	}
	if mode != netlink.MOJANET_MODE_BRIDGE {
		t.Fatalf("expected %d got %d", netlink.MOJANET_MODE_BRIDGE, mode)
	}
	// test invalid mode
	mode, err = setMojanetMode("foo")
	if err == nil {
		t.Fatal("invalid mojanet mode should have returned an error")
	}
	if mode != 0 {
		t.Fatalf("expected 0 got %d", mode)
	}
	// test null mode
	mode, err = setMojanetMode("")
	if err == nil {
		t.Fatal("invalid mojanet mode should have returned an error")
	}
	if mode != 0 {
		t.Fatalf("expected 0 got %d", mode)
	}
}
