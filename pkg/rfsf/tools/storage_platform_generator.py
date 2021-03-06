#!/usr/local/bin/python3
#
# Copyright 2021, 2022 Hewlett Packard Enterprise Development LP
# Other additional copyright holders may be indicated within.
#
# The entirety of this work is licensed under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
#
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#
# This script will patch the provided directory with
# specific implementation of the Storage Platform function
# calls.
#
# This script should run after an openapi-generator call
#
# Author: Nate Roiger

import argparse, os, sys

# helper function to find custom endpoints in the provided file.
# this was used to extract the cray specifics in api_default.go prior
# the storage platform auto-generation code.
def find_custom_endpoints(args):
    packages = """
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/chassis"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/drives"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/event"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/filesystem"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/serviceroot"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/storagepool"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/storageservices"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/template"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/volume"
"github.com/NearNodeFlash/nnf-ec/internal/rfsf/pkg/hamanager"
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
	openapi "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"
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




