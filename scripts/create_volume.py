#!/usr/bin/env python3

import argparse, sys

from common import (
    ByteSizeAction,

    addServerArguments,
    connect,
)

def get_servers(c):
    servers = []
    rsp = c.get("/redfish/v1/StorageServices/NNF/Endpoints")
    if not rsp.ok:
        raise Exception("NNF Server Endpoints not found")

    for member in rsp.json()['Members']:
        odataid = member['@odata.id']
        rsp = c.get(odataid)
        if not rsp.ok:
            raise Exception(f"NNF Server Endpoint {odataid} not found")

        servers.append(rsp.json())
    return servers

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Create a logical volume on the NNF Controller')
    parser.add_argument('-c', '--capacity', type=str, action=ByteSizeAction, default=str(1024 * 1024 * 1024), help='specify the capacity (bytes or SI)')
    parser.add_argument('-r', '--relaxed', action='store_true', help='set the allocation policy in a relaxed mode')
    parser.add_argument('-s', '--servers', nargs='+', type=str, help='specificy a server or list of servers to connect storage to, as defined by their name')
    parser.add_argument('--server-list', action='store_true', help='list the available servers and their status\', then exit')
    
    addServerArguments(parser)

    args = parser.parse_args()
    args.capacity = int(args.capacity)

    c = connect(args)

    rsp = c.get("/redfish/v1/StorageServices/NNF")
    if not rsp.ok:
        raise Exception("NNF Storage Service not found")

    if args.server_list:
        header = ["ID", "Name", "State"]
        formatter = "{:>3}{:>18}{:>24}"
        print(formatter.format(*header))
        for server in get_servers(c):
            print(formatter.format(server['Id'], server['Name'], server['Status']['State']))
        sys.exit(0)

    print(f'Create Capacity of {args.capacity / (1024 * 1024)} MiB')

    payload = {'Capacity': {'Data': {'AllocatedBytes': args.capacity}}}
    if args.relaxed:
        payload['Oem'] = {'Compliance': 'relaxed'}
    
    rsp = c.post('/redfish/v1/StorageServices/NNF/StoragePools', payload)
    print(rsp)
    print(rsp.json())
    if not rsp.ok:
        raise Exception("Unable to create Storage Pool")
    poolid = rsp.json()['@odata.id']

    servers = []
    if args.servers:
        rsp = c.get("/redfish/v1/StorageServices/NNF/Endpoints")
        if not rsp.ok:
            raise Exception("NNF Server Endpoints not found")

        print(rsp)
        print(rsp.json())
        for member in rsp.json()['Members']:
            odataid = member['@odata.id']
            rsp = c.get(odataid)
            if rsp.json()['Name'] in args.servers:
                servers.append(odataid)


    for serverid in servers:
        payload = {
            'Links': {
                'StoragePool': { '@odata.id': poolid },
                'ServerEndpoint': { '@odata.id': serverid },
            },
        }

        rsp = c.post('/redfish/v1/StorageServices/NNF/StorageGroups', payload)
        print(rsp)
        print(rsp.json())

