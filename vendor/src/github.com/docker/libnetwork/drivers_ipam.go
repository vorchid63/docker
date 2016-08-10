package libnetwork

import (
        log "github.com/Sirupsen/logrus"
       	"github.com/docker/libnetwork/drvregistry"
	"github.com/docker/libnetwork/ipamapi"
	builtinIpam "github.com/docker/libnetwork/ipams/builtin"
	nullIpam "github.com/docker/libnetwork/ipams/null"
	remoteIpam "github.com/docker/libnetwork/ipams/remote"
)

func initIPAMDrivers(r *drvregistry.DrvRegistry, lDs, gDs interface{}) error {
	for _, fn := range [](func(ipamapi.Callback, interface{}, interface{}) error){
		builtinIpam.Init,
		remoteIpam.Init,
		nullIpam.Init,
	} {
		log.Errorf("VLU-initIPAMDrivers: ===========================================================")
                log.Errorf("VLU-initIPAMDrivers: controller's init IPAM drivers for builtin, remote and Null")
		if err := fn(r, lDs, gDs); err != nil {
			return err
		}
	}

	return nil
}
