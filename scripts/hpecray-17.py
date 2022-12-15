#!/usr/bin/env python3
#
# Copyright 2022 Hewlett Packard Enterprise Development LP
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

import argparse, sys, time

import common, connection

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Interactive tool for NNF Element Controller')
    parser.add_argument('--size', type=str, action=common.ByteSizeAction, default=str(500 * 2**30), help='storage pool size')
    parser.add_argument('--endpoint', type=str, default='0', help='the endpoint id to attach to')

    connection.addServerArguments(parser)

    args = parser.parse_args()

    conn = connection.connect(args, '/redfish/v1/StorageServices/NNF')

    # Storage Pool - Create
    payload = {
        'Capacity': {'Data': {'AllocatedBytes': int(args.size)}},
        'Oem': {'Compliance': 'relaxed'}
    }

    print(f'Creating Storage Pool: Size: {args.size}...', end='')
    response = conn.post('/StoragePools', payload)
    if not response.ok:
        print(f'Storage Pool Create: Error: {response.status_code}')
        sys.exit(0)

    pool_id = response.json()['Id']
    print(f'Created: Id: {pool_id}')

    endpoint_id = args.endpoint

    print(f'Beginning Storage Group Create/Delete Loop: Pool: {pool_id} Endpoint: {endpoint_id}')
    while True:
        # Storage Group - Create
        payload = {
            'Links': {
                'StoragePool': { '@odata.id': f'{conn.base}/StoragePools/{pool_id}'},
                'ServerEndpoint': { '@odata.id': f'{conn.base}/Endpoints/{endpoint_id}'}
            }
        }

        print(f'Creating Storage Group...', end='')
        response = conn.post('/StorageGroups', payload)
        if not response.ok:
            print(f'Storage Group Create: Error: {response.status_code}')
            sys.exit(0)

        group_id = response.json()['Id']
        print(f'Created: Id: {group_id}')

        print(f'Pause 5 seconds for Storage Group to come ready')
        time.sleep(5.0)

        # Storage Group - Delete
        print(f'Delete Storage Group {group_id}....', end='')
        response = conn.delete(f'/StorageGroups/{group_id}')
        if not response.ok:
            print(f'Storage Group Delete: Error: {response.status_code}')
            sys.exit(0)
        print(f'Deleted: Id: {group_id}')

        print(f'Pause 5 seconds for Storage Group to delete')
        time.sleep(5.0)