package main

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/docker/libnetwork/netlabel"
	"github.com/mojaisolator/mojaipam"
)

type mojaplugin struct {
	i mojaipam.Ipam
}

func main() {

	// create a http demultiplexer
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	isolator := mojaplugin{}
	handler := mojaipam.NewHandler(isolator.i)

	handler.ServeTCP("test_isolator", ":8080")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

//======================================================================

// GetCapabilities find out if mojaisolator requires mac address or not
func (isolator *mojaplugin) GetCapabilities() (*mojaipam.CapabilitiesResponse, error) {
	r := mojaipam.CapabilitiesResponse{false}
	return &r, nil
}

// GetDefaultAddressSpaces get mojaisolator's default address space
func (isolator *mojaplugin) GetDefaultAddressSpaces() (*mojaipam.AddressSpacesResponse, error) {
	r := mojaipam.AddressSpacesResponse{"mojaLocal", "mojaGlobal"}
	return &r, nil
}

// RequestPool request an address pool from mojaisolator
func (isolator *mojaplugin) RequestPool(*mojaipam.RequestPoolRequest) (*mojaipam.RequestPoolResponse, error) {
	r := mojaipam.RequestPoolResponse{"poolId", "pool", map[string]string{netlabel.Prefix: "10.2.3.0"}}
	return &r, nil
}

// ReleasePool relase an address pool from mojaisolator
func (isolator *mojaplugin) ReleasePool(*mojaipam.ReleasePoolRequest) error {
	return nil
}

// RequestAddress request an address from mojaisolator
func (isolator *mojaplugin) RequestAddress(*mojaipam.RequestAddressRequest) (*mojaipam.RequestAddressResponse, error) {
	r := mojaipam.RequestAddressResponse{"10.2.3.x", map[string]string{netlabel.DriverPrefix: "10.2.3.5"}}
	return &r, nil
}

// ReleaseAddress release an address from mojaisolator
func (isolator *mojaplugin) ReleaseAddress(*mojaipam.ReleaseAddressRequest) error {
	return nil
}
