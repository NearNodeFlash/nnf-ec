#!/usr/bin/env python3

# Copyright 2023 Hewlett Packard Enterprise Development LP
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


import argparse
import sys

import common
import connection

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Allocate and attach Rabbit NVMe storage to one or more compute nodes',
        epilog='NOTE: "nnf-ec" must be running on the Rabbit')

    parser.add_argument('--node', nargs='+', type=int, choices=range(0, 16), default=None,
                        help='compute node index to attach to (default all 16 nodes), specify "0" "1" ... for the compute nodes you want to use')
    parser.add_argument('--size', type=str, action=common.ByteSizeAction,
                        default=common.ByteSizeStringToIntegerBytes('500GB'), help='storage pool size (default 500GB)')

    connection.addServerArguments(parser)

    args = parser.parse_args()

    if args.node is None:
        args.node = range(0, 16)

    conn = connection.connect(args, '/redfish/v1/StorageServices/NNF')

    for node in args.node:
        print(f"Starting Node '{node}' Sequence...")

        print('\tCreating Storage Pool...', end=' ')

        payload = {
            'Capacity': {'Data': {'AllocatedBytes': int(args.size)}},
            'Oem': {'Compliance': 'relaxed'}
        }

        response = conn.post('/StoragePools', payload)
        if not response.ok:
            print(f'ERROR: {response.status_code}')
            print(f'RESPONSE: {response}')
            sys.exit(1)

        pool_id = response.json()['Id']
        print(f"Created Storage Pool ID '{pool_id}'")

        print('\tCreating Storage Group...', end=' ')

        endpoint_id = node + 1  # Endpoint 0 is reserved for Rabbit

        payload = {
            'Links': {
                'StoragePool': {'@odata.id': f'{conn.base}/StoragePools/{pool_id}'},
                'ServerEndpoint': {'@odata.id': f'{conn.base}/Endpoints/{endpoint_id}'}
            }
        }

        response = conn.post('/StorageGroups', payload)
        if not response.ok:
            print(f'ERROR: {response.status_code}')
            print(f'RESPONSE: {response}')
            sys.exit(1)

        group_id = response.json()['Id']
        print(f"Created Storage Group ID '{group_id}'")

        print(f"Storage Node '{node}' Ready.")

    print("")
    print("All Nodes Ready")
    print("""
You can now create an LVM volume on each node that groups the NVMe block devices
into a common file system. Refer to the lvm.sh script.
""")
