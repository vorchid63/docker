package isolhttps

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/docker/libnetwork/drivers/remote/api"
	"github.com/docker/libnetwork/ipamapi"
	iapi "github.com/docker/libnetwork/ipams/remote/api"
)

const (
	networkReceiver = "NetworkDriver"
	ipamReceiver    = ipamapi.PluginEndpointType
)

type Driver interface {
	GetCapabilities() (*api.GetCapabilityResponse, error)
	CreateNetwork(create *api.CreateNetworkRequest) error
	DeleteNetwork(delete *api.DeleteNetworkRequest) error
	CreateEndpoint(create *api.CreateEndpointRequest) (*api.CreateEndpointResponse, error)
	DeleteEndpoint(delete *api.DeleteEndpointRequest) error
	EndpointInfo(req *api.EndpointInfoRequest) (*api.EndpointInfoResponse, error)
	JoinEndpoint(j *api.JoinRequest) (response *api.JoinResponse, error error)
	LeaveEndpoint(leave *api.LeaveRequest) error
	DiscoverNew(discover *api.DiscoveryNotification) error
	DiscoverDelete(delete *api.DiscoveryNotification) error
}

type isolhttps struct {
	d Driver
	i ipamapi.Ipam
}

func main(socket net.Listener, driver Driver, ipamDriver ipamapi.Ipam) error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	isolhttps := &isolhttps{driver, ipamDriver}

	router.Methods("POST").Path("/Plugin.Activate").HandlerFunc(isolhttps.handshake)

	handleMethod := func(receiver, method string, h http.HandlerFunc) {
		router.Methods("POST").Path(fmt.Sprintf("/%s.%s", receiver, method)).HandlerFunc(h)
	}

	handleMethod(networkReceiver, "GetCapabilities", isolhttps.getCapabilities)

	if driver != nil {
		handleMethod(networkReceiver, "CreateNetwork", isolhttps.createNetwork)
		handleMethod(networkReceiver, "DeleteNetwork", isolhttps.deleteNetwork)
		handleMethod(networkReceiver, "CreateEndpoint", isolhttps.createEndpoint)
		handleMethod(networkReceiver, "DeleteEndpoint", isolhttps.deleteEndpoint)
		handleMethod(networkReceiver, "EndpointOperInfo", isolhttps.infoEndpoint)
		handleMethod(networkReceiver, "Join", isolhttps.joinEndpoint)
		handleMethod(networkReceiver, "Leave", isolhttps.leaveEndpoint)
	}

	if ipamDriver != nil {
		handleMethod(ipamReceiver, "GetDefaultAddressSpaces", isolhttps.getDefaultAddressSpaces)
		handleMethod(ipamReceiver, "RequestPool", isolhttps.requestPool)
		handleMethod(ipamReceiver, "ReleasePool", isolhttps.releasePool)
		handleMethod(ipamReceiver, "RequestAddress", isolhttps.requestAddress)
		handleMethod(ipamReceiver, "ReleaseAddress", isolhttps.releaseAddress)
	}

	return http.Serve(socket, router)
}

func decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
	}
	return err
}

// === protocol handlers

type handshakeResp struct {
	Implements []string
}

func (isolhttps *isolhttps) handshake(w http.ResponseWriter, r *http.Request) {
	var resp handshakeResp
	if isolhttps.d != nil {
		resp.Implements = append(resp.Implements, "NetworkDriver")
	}
	if isolhttps.i != nil {
		resp.Implements = append(resp.Implements, "IpamDriver")
	}
	err := json.NewEncoder(w).Encode(&resp)
	if err != nil {
		sendError(w, "encode error", http.StatusInternalServerError)
		return
	}
}

func (isolhttps *isolhttps) getCapabilities(w http.ResponseWriter, r *http.Request) {
	caps, err := isolhttps.d.GetCapabilities()
	objectOrErrorResponse(w, caps, err)
}

func (isolhttps *isolhttps) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create api.CreateNetworkRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, isolhttps.d.CreateNetwork(&create))
}

func (isolhttps *isolhttps) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, isolhttps.d.DeleteNetwork(&delete))
}

func (isolhttps *isolhttps) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var create api.CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		sendError(w, "unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	res, err := isolhttps.d.CreateEndpoint(&create)
	objectOrErrorResponse(w, res, err)
}

func (isolhttps *isolhttps) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, isolhttps.d.DeleteEndpoint(&delete))
}

func (isolhttps *isolhttps) infoEndpoint(w http.ResponseWriter, r *http.Request) {
	var req api.EndpointInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	info, err := isolhttps.d.EndpointInfo(&req)
	objectOrErrorResponse(w, info, err)
}

func (isolhttps *isolhttps) joinEndpoint(w http.ResponseWriter, r *http.Request) {
	var join api.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&join); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	res, err := isolhttps.d.JoinEndpoint(&join)
	objectOrErrorResponse(w, res, err)
}

func (isolhttps *isolhttps) leaveEndpoint(w http.ResponseWriter, r *http.Request) {
	var l api.LeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, isolhttps.d.LeaveEndpoint(&l))
}

func (isolhttps *isolhttps) discoverNew(w http.ResponseWriter, r *http.Request) {
	var disco api.DiscoveryNotification
	if err := json.NewDecoder(r.Body).Decode(&disco); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, isolhttps.d.DiscoverNew(&disco))
}

func (isolhttps *isolhttps) discoverDelete(w http.ResponseWriter, r *http.Request) {
	var disco api.DiscoveryNotification
	if err := json.NewDecoder(r.Body).Decode(&disco); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, isolhttps.d.DiscoverDelete(&disco))
}

// ===

func (isolhttps *isolhttps) getDefaultAddressSpaces(w http.ResponseWriter, r *http.Request) {
	local, global, err := isolhttps.i.GetDefaultAddressSpaces()
	response := &iapi.GetAddressSpacesResponse{
		LocalDefaultAddressSpace:  local,
		GlobalDefaultAddressSpace: global,
	}
	objectOrErrorResponse(w, response, err)
}

func (isolhttps *isolhttps) requestPool(w http.ResponseWriter, r *http.Request) {
	var rq iapi.RequestPoolRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	poolID, pool, data, err := isolhttps.i.RequestPool(rq.AddressSpace, rq.Pool, rq.SubPool, rq.Options, rq.V6)
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	response := &iapi.RequestPoolResponse{
		PoolID: poolID,
		Pool:   pool.String(),
		Data:   data,
	}
	objectResponse(w, response)
}

func (isolhttps *isolhttps) releasePool(w http.ResponseWriter, r *http.Request) {
	var rq iapi.ReleasePoolRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	err := isolhttps.i.ReleasePool(rq.PoolID)
	emptyOrErrorResponse(w, err)
}

func (isolhttps *isolhttps) requestAddress(w http.ResponseWriter, r *http.Request) {
	var rq iapi.RequestAddressRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	address, data, err := isolhttps.i.RequestAddress(rq.PoolID, net.ParseIP(rq.Address), rq.Options)
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	response := &iapi.RequestAddressResponse{
		Address: address.String(),
		Data:    data,
	}
	objectResponse(w, response)
}

func (isolhttps *isolhttps) releaseAddress(w http.ResponseWriter, r *http.Request) {
	var rq iapi.ReleaseAddressRequest
	if err := decode(w, r, &rq); err != nil {
		return
	}
	err := isolhttps.i.ReleaseAddress(rq.PoolID, net.ParseIP(rq.Address))
	emptyOrErrorResponse(w, err)
}

// ===

func notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}

func errorResponse(w http.ResponseWriter, fmtString string, item ...interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"Err": fmt.Sprintf(fmtString, item...),
	})
}

func objectResponse(w http.ResponseWriter, obj interface{}) {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		sendError(w, "Could not JSON encode response", http.StatusInternalServerError)
		return
	}
}

func emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(map[string]string{})
}

func objectOrErrorResponse(w http.ResponseWriter, obj interface{}, err error) {
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	objectResponse(w, obj)
}

func emptyOrErrorResponse(w http.ResponseWriter, err error) {
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	emptyResponse(w)
}
