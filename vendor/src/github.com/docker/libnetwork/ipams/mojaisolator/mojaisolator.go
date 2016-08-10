package mojaisolator

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/plugins"
	"github.com/docker/libnetwork/discoverapi"
	"github.com/docker/libnetwork/ipamapi"
	"github.com/docker/libnetwork/ipams/mojaisolator/api"
	"github.com/docker/libnetwork/types"
)

type allocator struct {
	endpoint *plugins.Client
	name     string
}

// PluginResponse is the interface for the plugin request responses
type PluginResponse interface {
	IsSuccess() bool
	GetError() string
}

func newAllocator(name string, client *plugins.Client) ipamapi.Ipam {
        log.Errorf("VLU-newAllocator: mojaisolator allocate %s", name)
	a := &allocator{name: name, endpoint: client}
	return a
}

// Init registers a mojaisolator ipam when its plugin is activated
func Init(cb ipamapi.Callback, l, g interface{}) error {
     	log.Errorf("VLU-Init: mojaisolator IPAM's  init ipam driver:%s", ipamapi.PluginEndpointType)
	plugins.Handle(ipamapi.PluginEndpointType, func(name string, client *plugins.Client) {
		log.Errorf("VLU-Init: init mojaisolator IPAM driver:%s add functon:%s to endpont handlers  for plugin client=%s", 
			   ipamapi.PluginEndpointType, client )
		a := newAllocator(name, client)
		if cps, err := a.(*allocator).getCapabilities(); err == nil {
		   	log.Errorf("VLU-Init: init mojaisolator IPAM driver:%s retrieved capabilites for plugin client=%s", 
			            ipamapi.PluginEndpointType, client)
		   	log.Errorf("VLU-Init: init mojaisolator IPAM driver:%s calling plugin client=%s call back to Register IAPM with capibilit", 
			            ipamapi.PluginEndpointType, client)
			if err := cb.RegisterIpamDriverWithCapabilities(name, a, cps); err != nil {
				log.Errorf("error registering mojaisolator ipam driver %s due to %v", name, err)
			}
		} else {
			log.Infof("mojaisolator ipam driver %s does not support capabilities", name)
			log.Debug(err)
			if err := cb.RegisterIpamDriver(name, a); err != nil {
				log.Errorf("error registering mojaisolator ipam driver %s due to %v", name, err)
			}
		}
	})
	return nil
}

func (a *allocator) call(methodName string, arg interface{}, retVal PluginResponse) error {
	method := ipamapi.PluginEndpointType + "." + methodName
     	log.Errorf("VLU-call: mojaisolator IPAM's  call:%s", method)
	err := a.endpoint.Call(method, arg, retVal)
	if err != nil {
		return err
	}
	if !retVal.IsSuccess() {
		return fmt.Errorf("mojaisolator: %s", retVal.GetError())
	}
	return nil
}

func (a *allocator) getCapabilities() (*ipamapi.Capability, error) {
	var res api.GetCapabilityResponse
     	log.Errorf("VLU-call: mojaisolator IPAM's  call:GetCapabilities")
	if err := a.call("GetCapabilities", nil, &res); err != nil {
		return nil, err
	}
	return res.ToCapability(), nil
}

// GetDefaultAddressSpaces returns the local and global default address spaces
func (a *allocator) GetDefaultAddressSpaces() (string, string, error) {
	res := &api.GetAddressSpacesResponse{}
     	log.Errorf("VLU-call: mojaisolator IPAM's  call:GetDefaultAddressSpaces")
	if err := a.call("GetDefaultAddressSpaces", nil, res); err != nil {
		return "", "", err
	}
	return res.LocalDefaultAddressSpace, res.GlobalDefaultAddressSpace, nil
}

// RequestPool requests an address pool in the specified address space
func (a *allocator) RequestPool(addressSpace, pool, subPool string, options map[string]string, v6 bool) (string, *net.IPNet, map[string]string, error) {
        log.Errorf("VLU-RequestPool: mojaisolator call RequestPool ")
	req := &api.RequestPoolRequest{AddressSpace: addressSpace, Pool: pool, SubPool: subPool, Options: options, V6: v6}
	res := &api.RequestPoolResponse{}
	if err := a.call("RequestPool", req, res); err != nil {
		return "", nil, nil, err
	}
	retPool, err := types.ParseCIDR(res.Pool)
	return res.PoolID, retPool, res.Data, err
}

// ReleasePool removes an address pool from the specified address space
func (a *allocator) ReleasePool(poolID string) error {
	req := &api.ReleasePoolRequest{PoolID: poolID}
	res := &api.ReleasePoolResponse{}
        log.Errorf("VLU-ReleasePool: mojaisolator call ReleasePool ")
	return a.call("ReleasePool", req, res)
}

// RequestAddress requests an address from the address pool
func (a *allocator) RequestAddress(poolID string, address net.IP, options map[string]string) (*net.IPNet, map[string]string, error) {
	var (
		prefAddress string
		retAddress  *net.IPNet
		err         error
	)
	if address != nil {
		prefAddress = address.String()
	}
	req := &api.RequestAddressRequest{PoolID: poolID, Address: prefAddress, Options: options}
	res := &api.RequestAddressResponse{}
        log.Errorf("VLU-RequestAddress: mojaisolator call RequestAddress ")
	if err := a.call("RequestAddress", req, res); err != nil {
		return nil, nil, err
	}
	if res.Address != "" {
		retAddress, err = types.ParseCIDR(res.Address)
	}
	return retAddress, res.Data, err
}

// ReleaseAddress releases the address from the specified address pool
func (a *allocator) ReleaseAddress(poolID string, address net.IP) error {
	var relAddress string
	if address != nil {
		relAddress = address.String()
	}
	req := &api.ReleaseAddressRequest{PoolID: poolID, Address: relAddress}
	res := &api.ReleaseAddressResponse{}
        log.Errorf("VLU-ReleaseAddress: mojaisolator call ReleaseAddress ")
	return a.call("ReleaseAddress", req, res)
}

// DiscoverNew is a notification for a new discovery event, such as a new global datastore
func (a *allocator) DiscoverNew(dType discoverapi.DiscoveryType, data interface{}) error {
        log.Errorf("VLU-DiscoverNew: mojaisolator NIL ")
	return nil
}

// DiscoverDelete is a notification for a discovery delete event, such as a node leaving a cluster
func (a *allocator) DiscoverDelete(dType discoverapi.DiscoveryType, data interface{}) error {
        log.Errorf("VLU-DiscoverDelete: mojaisolator NIL ")
	return nil
}
