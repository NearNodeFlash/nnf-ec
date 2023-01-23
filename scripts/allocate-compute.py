#!/usr/bin/env python3

import argparse, sys

import common, connection

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Allocate storage and assign it to a compute node(s)',
        epilog='Must be executed on Rabbit with nnf-ec running')

    parser.add_argument('--node', nargs='+', choices=range(0, 16), default=None, help='the node index to attach to (default all 16 nodes)')
    parser.add_argument('--size', type=str, action=common.ByteSizeAction, default=common.ByteSizeStringToIntegerBytes('500GB'), help='storage pool size (default 500GB)')

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
