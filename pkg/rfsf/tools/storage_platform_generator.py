#!/usr/local/bin/python3
#
# This script will patch the provided directory with
# specific implementation of the Storage Platform function
# calls.
#
# This script should run after an openapi-generator call
# 
# Author: Nate Roiger
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#

import argparse, os, sys

# helper function to find custom endpoints in the provided file.
# this was used to extract the cray specifics in api_default.go prior
# the storage platform auto-generation code.
def find_custom_endpoints(args):
    packages = """
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/chassis"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/drives"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/event"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/filesystem"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/serviceroot"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/storagepool"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/storageservices"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/template"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/volume"
"stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/hamanager"
""".splitlines(False)[1:]
    packages = map(lambda s: os.path.basename(s.strip('"')), packages)
    
    with open(os.path.join(args.dir, 'temp.go'), 'w') as dfp:

        with open(os.path.join(args.dir, args.src)) as fp:
            in_func = False
            print_func = False
            for _, ln in enumerate(fp):
                if ln.startswith('// Redfish'):
                    in_func = True
                    lines = [ln,]
                    print(ln)
                elif in_func:
                    if ln.startswith('func Redfish'):
                        ln = ln.replace('func Redfish', 'func (c *StoragePlatformApiController) Redfish')
                    lines.append(ln)
                    if ln.lstrip().split('.')[0] in packages:
                        print_func = True
                    if ln.startswith('}'):
                        in_func = False
                        if print_func:
                            print_func = False
                            dfp.writelines(lines)
                            dfp.write('\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Generate the Storage Platform API Service Defaults')
    parser.add_argument('src', nargs='?', default='sp_api.go')
    parser.add_argument('default', nargs='?', default='api_default.go')
    parser.add_argument('dest', nargs='?', default='sp_api_default.go')
    parser.add_argument('dir', nargs='?', default=os.path.join('pkg', 'routermux'))
    parser.add_argument('-c', action='store_true', help='Find custom resources in src, instead of generate')

    args = parser.parse_args()

    if args.c:
        find_custom_endpoints(args)
        sys.exit(0)

    # Build the list of functions defined in the source file
    src_funcs = []
    with open(os.path.join(args.dir, args.src)) as fp:
         for _, ln in enumerate(fp):
             if ln.startswith('func'):
                 # Extract the function name from 
                 #   func (c *StoragePlatformApiController) RedfishV1CompositionServiceResourceBlocksResourceBlockIdSystemsComputerSystemIdStorageStorageIdVolumesPost(w http.ResponseWriter, r *http.Request) {
                 name = ln.split(' ')[3].rstrip('(w')
                 print(f'Found SP function {name}')
                 src_funcs.append(name)
             

    # Copy from the default api service into the destination; ignoring the
    # definitions defined in the source service. This implements all the
    # unused / deadend endpoints.

    with open(os.path.join(args.dir, args.dest), 'w') as dfp:

        dfp.write(
"""/*
 * Redfish
 * 
 * This contains the default implementation of a Storage Platform Redfish service
 * 
 * Author: Auto-generated from sp_generator.py
 *
 * Copyright 2020 Hewlett Packard Enterprise Development LP
 *
 */
 
 package routermux

 import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	openapi "stash.us.cray.com/rabsw/nnf-ec/pkg/rfsf/pkg/models"
 )

// A StoragePlatformApiController binds http requests to an api service and writes the service results to the http response
type StoragePlatformApiController struct {
	service DefaultApiServicer
}

// NewStoragePlatformApiController creates a storage platform api controller
func NewStoragePlatformApiController(s DefaultApiServicer) Router {
	return &StoragePlatformApiController{service: s}
}

"""
 )
        with open(os.path.join(args.dir, args.default)) as sfp:
        
            in_func = False
            for idx, ln in enumerate(sfp):
                if in_func:
                    if ln.startswith('func (c *DefaultApiController) '):
                        ln = ln.replace('DefaultApiController', 'StoragePlatformApiController')
                    dfp.write(ln)
                    if ln.startswith('}'):
                        in_func = False
                        dfp.write('\n')
                elif ln.startswith('// Redfish'):
                    endpoint = ln.split(' ')[1]
                    if endpoint not in src_funcs:
                        dfp.write(ln)
                        in_func = True
                    else:
                        print(f'Endpoint exists in SP API: {endpoint}')
                elif ln.startswith('// Routes'):
                    in_func = True
                    dfp.write(ln)
                
                


