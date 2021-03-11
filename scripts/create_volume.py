#!/usr/bin/env python3

import argparse, json

from common import (
    ByteSizeAction,

    addServerArguments,
    connect,
)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Create a logical volume on the NNF Controller')
    parser.add_argument('-c', '--capacity', type=str, action=ByteSizeAction, default=str(1024 * 1024 * 1024), help='specify the capacity (bytes or SI)')
    parser.add_argument('-r', '--relaxed', action='store_true', help='set the allocation policy in a relaxed mode')
    
    addServerArguments(parser)

    args = parser.parse_args()
    args.capacity = int(args.capacity)

    c = connect(args)

    rsp = c.get("/redfish/v1/StorageServices/NNF")
    if not rsp.ok:
        raise Exception("NNF Endpoint not found")
    print(rsp.json())

    print(f'Create Capacity of {args.capacity / (1024 * 1024)} MiB')

    payload = {'Capacity': {'Data': {'AllocatedBytes': args.capacity}}}
    if args.relaxed:
        payload['Oem'] = {'Standard': 'relaxed'}
    
    rsp = c.post('/redfish/v1/StorageServices/NNF/StoragePools', payload)

    print(rsp)


