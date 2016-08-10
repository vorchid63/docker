package mojanet

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/ns"
	"github.com/vishvananda/netlink"
)

const (
	dummyPrefix      = "dm-" // mojanet prefix for dummy parent interface
	mojanetKernelVer    = 3     // minimum mojanet kernel support
	mojanetMajorVer     = 9     // minimum mojanet major kernel support
)

// Create the mojanet slave specifying the source name
func createMojanet(containerIfName, parent, mojanetMode string) (string, error) {
	// Set the mojanet mode. Default is bridge mode
	_, err := setMojanetMode(mojanetMode)
	if !parentExists(parent) {
		return "", fmt.Errorf("the requested parent interface %s was not found on the Docker host", parent)
	}
	// Get the link for the master index (Example: the docker host eth iface)
	parentLink, err := ns.NlHandle().LinkByName(parent)
	if err != nil {
		return "", fmt.Errorf("error occoured looking up the %s parent iface %s error: %s", mojanetType, parent, err)
	}
	// Create a mojanet link
	mojanet := &netlink.Mojanet{
		LinkAttrs: netlink.LinkAttrs{
			Name:        containerIfName,
			ParentIndex: parentLink.Attrs().Index,
		},
		Mode:  netlink.MOJANET_MODE_BRIDGE,
	}
	if err := ns.NlHandle().LinkAdd(mojanet); err != nil {
		return "", fmt.Errorf("failed to create the %s port: %v", mojanetType, err)
	}

	return mojanet.Attrs().Name, nil
}


// setMojanetMode setter for mojanet port types
func setMojanetMode(mode string) (netlink.MojanetMode, error) {
	switch mode {
	case modeBridge:
		return netlink.MOJANET_MODE_BRIDGE, nil
	default:
		return 0, fmt.Errorf("unknown mojanet mode: %s", mode)
	}
}

// parentExists check if the specified interface exists in the default namespace
func parentExists(ifaceStr string) bool {
	_, err := ns.NlHandle().LinkByName(ifaceStr)
	if err != nil {
		return false
	}

	return true
}

// createVlanLink parses sub-interfaces and vlan id for creation
func createMojanetLink(parentName string) error {
        // get the parent link to attach a vlan subinterface
     	parentLink, err := ns.NlHandle().LinkByName(parentName)
	if err != nil {
	   	return fmt.Errorf("failed to find master interface %s on the Docker host: %v", parentName, err)
	}
	mojanetLink := &netlink.Mojanet{
		LinkAttrs: netlink.LinkAttrs{
			Name:        parentName,
			ParentIndex: parentLink.Attrs().Index,
		},
		Mode:  netlink.MOJANET_MODE_BRIDGE,
	}
	// create the subinterface
	if err := ns.NlHandle().LinkAdd(mojanetLink); err != nil {
		return fmt.Errorf("failed to create %s  link: %v", mojanetLink.Name, err)
	}
	// Bring the new netlink iface up
	if err := ns.NlHandle().LinkSetUp(mojanetLink); err != nil {
		return fmt.Errorf("failed to enable %s the mojanet parent link %v", mojanetLink.Name, err)
	}
	logrus.Debugf("Added a mojanet netlink subinterface: %s: %d", parentName)
	return nil
}

// delVlanLink verifies only sub-interfaces with a vlan id get deleted
func delMojanetLink(linkName string) error {
	// delete the vlan subinterface
	mojanetLink, err := ns.NlHandle().LinkByName(linkName)
	if err != nil {
		return fmt.Errorf("failed to find interface %s on the Docker host : %v", linkName, err)
	}
	// verify a parent interface isn't being deleted
	if mojanetLink.Attrs().ParentIndex == 0 {
		return fmt.Errorf("interface %s does not appear to be a slave device: %v", linkName, err)
	}
	// delete the mojanet slave device
	if err := ns.NlHandle().LinkDel(mojanetLink); err != nil {
		return fmt.Errorf("failed to delete  %s link: %v", linkName, err)
	}
	logrus.Debugf("Deleted a netlink subinterface: %s", linkName)
	return nil
}

// createDummyLink creates a dummy0 parent link
func createDummyLink(dummyName, truncNetID string) error {
	// create a parent interface since one was not specified
	parent := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: dummyName,
		},
	}
	if err := ns.NlHandle().LinkAdd(parent); err != nil {
		return err
	}
	parentDummyLink, err := ns.NlHandle().LinkByName(dummyName)
	if err != nil {
		return fmt.Errorf("error occoured looking up the %s parent iface %s error: %s", mojanetType, dummyName, err)
	}
	// bring the new netlink iface up
	if err := ns.NlHandle().LinkSetUp(parentDummyLink); err != nil {
		return fmt.Errorf("failed to enable %s the mojanet parent link: %v", dummyName, err)
	}

	return nil
}

// delDummyLink deletes the link type dummy used when -o parent is not passed
func delDummyLink(linkName string) error {
	// delete the vlan subinterface
	dummyLink, err := ns.NlHandle().LinkByName(linkName)
	if err != nil {
		return fmt.Errorf("failed to find link %s on the Docker host : %v", linkName, err)
	}
	// verify a parent interface is being deleted
	if dummyLink.Attrs().ParentIndex != 0 {
		return fmt.Errorf("link %s is not a parent dummy interface", linkName)
	}
	// delete the mojanet dummy device
	if err := ns.NlHandle().LinkDel(dummyLink); err != nil {
		return fmt.Errorf("failed to delete the dummy %s link: %v", linkName, err)
	}
	logrus.Debugf("Deleted a dummy parent link: %s", linkName)

	return nil
}

// getDummyName returns the name of a dummy parent with truncated net ID and driver prefix
func getDummyName(netID string) string {
	return fmt.Sprintf("%s%s", dummyPrefix, netID)
}