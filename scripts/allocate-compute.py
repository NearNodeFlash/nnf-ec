#!/usr/bin/env python3

import argparse, sys, time

import common, connection

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Allocate storage and assign it to a compute node(s)',
        epilog = 'Must executed on Rabbit with nnf-ec running')
    
    parser.add_argument('--node', nargs='+', choices=range(0,16), default=None, help='the node index to attach to')
    parser.add_argument('--size', type=str, action=common.ByteSizeAction, default=common.ByteSizeStringToIntegerBytes('500GB'), help='storage pool size (default 500GB)')

    connection.addServerArguments(parser)

    args = parser.parse_args()

    if args.node is None:
        args.node = range(0, 16)

    conn = connection.connect(args, '/redfish/v1/StorageServices/NNF')

    for node in args.node:
        print(f"Starting Node '{node}' Sequence...")

        print(f'\tCreating Storage Pool...', end=' ')

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

        print(f'\tCreating Storage Group...', end=' ')

        endpoint_id = node + 1 # Endpoint 0 is reserved for Rabbit

        payload = {
            'Links': {
                'StoragePool': { '@odata.id': f'{conn.base}/StoragePools/{pool_id}'},
                'ServerEndpoint': { '@odata.id': f'{conn.base}/Endpoints/{endpoint_id}'}
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


"""
cm power status -t node x9000c3s*b*n* | awk -F :  '{print $1}'
readarray SYSTEMS < <(cm power status -t node x9000c3s*b*n* | awk -F :  '{print $1}')

for SYSTEM in ${SYSTEMS[@]}; do ssh $SYSTEM 'for DRIVE in $(ls -v /dev/nvme* | grep -E "nvme[[:digit:]]+n[[:digit:]]+$"); do nvme id-ns $DRIVE | grep "NVME"; done'; done

SYSTEM=${SYSTEMS[0]}
ssh $SYSTEM 'for DRIVE in $(ls -v /dev/nvme* | grep -E "nvme[[:digit:]]+n[[:digit:]]+$"); do nvme id-ns $DRIVE | grep "NVME"; done'
ssh $SYSTEM ls /dev/rabbit/rabbit

"""