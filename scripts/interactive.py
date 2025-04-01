#!/usr/bin/env python3
#
# Copyright 2021-2025 Hewlett Packard Enterprise Development LP
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

import cmd, argparse, json, sys

import connection

from common import ByteSizeStringToIntegerBytes

class Command(cmd.Cmd):
    def __init__(self, conn: connection.Connection, **kwargs):
        self.conn = conn
        super().__init__(**kwargs)

    def handle_response(self, response):
        if response.ok:
            if response.json() != None:
                print(json.dumps(response.json(), indent=4))
        else:
            print(f'Error: {response.status_code}')

    def do_back(self,arg):
        return True

    def do_exit(self,arg):
        return sys.exit(0)

class StoragePool(Command):
    intro = 'Create/Put/Get/List/Delete/Patch Storage Pools'
    prompt = '(nnf)' + '(storage pool)'

    def create_payload(self, size):
        if size is None or size == "":
            size = '1GB'
            print(f'No size specified - defaulting to {size}')

        try:
            size = ByteSizeStringToIntegerBytes(size)
            payload = {
                'Name': "storage pool",
                'Description': "Pretty good storage group",
                'Capacity': {'Data': {'AllocatedBytes': int(size)}},
                'Oem': {'Compliance': 'strict'}
            }
            return payload
        except ValueError:
            print('*** SIZE argument should be integer')
            return None


    def do_create(self, arg):
        'Create a Storage Pool of specific [SIZE = 1GB]'
        payload = self.create_payload(arg)
        if payload is None:
            return

        self.handle_response(self.conn.post('/StoragePools', payload))

    def do_put(self, arg):
        'Create a Storage Pool with [POOL ID] [SIZE = 1GB]'

        if arg is None or arg == '':
            print('*** POOL ID is required parameter')
            return

        args = arg.split()
        size = None
        if len(args) >= 2:
            id, size, *_ = args
        else:
            id = arg

        payload = self.create_payload(size)
        if payload is None:
            return

        self.handle_response(self.conn.put(f'/StoragePools/{id}', payload))

    def do_get(self, arg):
        'Get a Storage Pool by POOL ID'
        self.handle_response(self.conn.get(f'/StoragePools/{arg}'))

    def do_list(self,arg):
        'List Storage Pools'
        self.handle_response(self.conn.get(f'/StoragePools'))

    def do_delete(self, arg):
        'Delete a Storage Pool by POOL ID'
        self.handle_response(self.conn.delete(f'/StoragePools/{arg}'))

    def patch_payload(self):
        payload = {
        }
        return payload

    def do_patch(self, arg):
        'Patch Storage Pool'

        if arg is None or arg == '':
            print('*** POOL ID is required parameter')
            return

        payload = self.patch_payload()
        if payload is None:
            return

        self.handle_response(self.conn.patch(f'/StoragePools/{arg}', payload))

    def do_storage(self, arg):
        'List Storage for provided POOL ID'
        if arg is None or arg == '':
            print('*** POOL ID is required parameter')

        response = self.conn.get(f'/StoragePools/{arg}/CapacitySources/0/ProvidingVolumes')
        if not response.ok:
            print(f'Error: {response.status_code}')
            return

        self.conn.push_path('') # Use @odata.id directly
        volumes = response.json()['Members']
        for volume in volumes:
            try:
                volumeId = volume['@odata.id']
                volume = self.conn.get(volumeId)

                storageId = volume.json()['Links']['OwningStorageResource']['@odata.id']
                storage = self.conn.get(storageId)

                locationType = storage.json()['Location']['PartLocation']['LocationType']
                locationValue = storage.json()['Location']['PartLocation']['LocationOrdinalValue']
                nqn = storage.json()['Identifiers'][0]['DurableName']
                nsid = volume.json()['Identifiers'][0]['DurableName']
                capacityBytes = volume.json()['CapacityBytes']

                print(f'{locationType} {locationValue}\t{nqn} {nsid} {capacityBytes}')
            except:
                print("Missing volume")

        self.conn.pop_path()

class ServerEndpoint(Command):
    intro = 'Get/List Server Endpoints'
    prompt = '(nnf)' + '(server endpoints)'

    def do_get(self,arg):
        'Get a Server Endpoint by [ENDPOINT ID]'
        self.handle_response(self.conn.get(f'/Endpoints/{arg}'))

    def do_list(self,arg):
        'List Server Endpoints'
        self.handle_response(self.conn.get(f'/Endpoints'))

class StorageGroup(Command):
    intro = 'Create/Put/Get/List/Delete Storage Groups'
    prompt = '(nnf)' + '(storage group)'

    def create_payload(self, pool, server):
        return {
            'Links': {
                'StoragePool': { '@odata.id': f'{self.conn.base}/StoragePools/{pool}'},
                'ServerEndpoint': { '@odata.id': f'{self.conn.base}/Endpoints/{server}'}
            }
        }

    def do_create(self,arg):
        'Create a Storage Group from [STORAGE POOL ID] [SERVER ENDPOINT ID]'

        try:
            pool, server = arg.split()[:2]
        except ValueError:
            print('Expected two arguments [STORAGE POOL ID] [SERVER ENDPOINT ID]')
            return

        payload = self.create_payload(pool, server)
        self.handle_response(self.conn.post(f'/StorageGroups', payload))

    def do_put(self,arg):
        'Create a Storage Group by [STORAGE GROUP ID] [STORAGE POOL ID] [SERVER ENDPOINT ID]'

        try:
            id, pool, server, *_ = arg.split()
        except ValueError:
            print('Expected three arguments [STORAGE GROUP ID] [STORAGE POOL ID] [SERVER ENDPOINT ID]')
            return

        payload = self.create_payload(pool, server)
        self.handle_response(self.conn.put(f'/StorageGroups/{id}', payload))

    def do_get(self,arg):
        'Get a Storage Group by [STORAGE GROUP ID]'
        self.handle_response(self.conn.get(f'/StorageGroups/{arg}'))

    def do_list(self,arg):
        'List Storage Groups'
        self.handle_response(self.conn.get(f'/StorageGroups'))

    def do_delete(self,arg):
        'Delete a Storage Group by [STORAGE GROUP ID]'
        self.handle_response(self.conn.delete(f'/StorageGroups/{arg}'))

class FileSystem(Command):
    intro = 'Create/Put/Get/List/Delete File Systems'
    prompt = '(nnf)' + '(file system)'

    def create_payload(self, type, name, pool, options):

        payload = {
            'Links': { 'StoragePool': { '@odata.id': f'{self.conn.base}/StoragePools/{pool}' } },
            'Oem': { 'Type': type, 'Name': name }
        }

        for option in options:
            key, value = option.split("=")
            payload['Oem'][key] = value

        return payload

    def do_create(self,arg):
        'Create a File System of [TYPE] [NAME] [STORAGE POOL ID] [OPTION=VALUE, ...]'

        try:
            type, name, pool, *options = arg.split()
        except ValueError:
            print('Expected three arguments [TYPE] [NAME] [STORAGE POOL ID]')
            return

        payload = self.create_payload(type, name, pool, options)
        self.handle_response(self.conn.post(f'/FileSystems', payload))

    def do_put(self, arg):
        'Create a File System by [FILE SYSTEM ID] [TYPE] [NAME] [STORAGE POOL ID] [OPTION=VALUE, ...]'

        try:
            id, type, name, pool, *options = arg.split()
        except ValueError:
            print('Expected four arguments [FILE SYSTEM ID] [TYPE] [NAME] [STORAGE POOL ID]')
            return

        payload = self.create_payload(type, name, pool, options)
        self.handle_response(self.conn.put(f'/FileSystems/{id}', payload))

    def do_get(self,arg):
        'Get a File System by [FILE SYSTEM ID]'
        self.handle_response(self.conn.get(f'/FileSystems/{arg}'))

    def do_list(self,arg):
        'List File Systems'
        self.handle_response(self.conn.get(f'/FileSystems'))

    def do_delete(self,arg):
        'Delete a File System by [FILE SYSTEM ID]'
        self.handle_response(self.conn.delete(f'/FileSystems/{arg}'))

    def do_share(self,arg):
        'File Share operations for [FILE SYSTEM ID]'
        if arg == None or arg == "":
            print('Expected one argument [FILE SYSTEM ID]')
            return

        FileShare(self.conn, arg).cmdloop()

class FileShare(Command):
    intro = 'Create/Put/Get/List/Delete File Shares'
    prompt = '(nnf)' + '(file system)' + '(file share)'

    def __init__(self, conn, fs, **kwargs):
        self.fs = fs
        super().__init__(conn, **kwargs)

    def create_payload(self, server, mount):
        return {
            'Links': { 'Endpoint': {'@odata.id': f'{self.conn.base}/Endpoints/{server}'}},
            'FileSharePath': mount
        }

    def do_create(self,arg):
        'Create a File Share to a [SERVER ENDPOINT] with [MOUNTPOINT]'

        try:
            server, mount, *_ = arg.split()
        except ValueError:
            print('Expected two arguments [SERVER ENDPOINT ID] [MOUNTPOINT]')
            return

        payload = self.create_payload(server, mount)
        self.handle_response(self.conn.post(f'/FileSystems/{self.fs}/ExportedFileShares', payload))

    def do_put(self,arg):
        'Put a File Share by [SHARE ID] [SERVER ENDPOINT ID] [MOUNTPOINT]'

        try:
            id, server, mount, *_ = arg.split()
        except ValueError:
            print('Expected three arguments [SHARE ID] [SERVER ENDPOINT ID] [MOUNTPOINT]')
            return

        payload = self.create_payload(server, mount)
        self.handle_response(self.conn.put(f'/FileSystems/{self.fs}/ExportedFileShares/{id}', payload))

    def do_get(self,arg):
        'Get a File Share with [FILE SHARE ID]'
        self.handle_response(self.conn.get(f'/FileSystems/{self.fs}/ExportedFileShares/{arg}'))

    def do_list(self,arg):
        'List File Shares on this File System'
        self.handle_response(self.conn.get(f'/FileSystems/{self.fs}/ExportedFileShares'))

    def do_delete(self,arg):
        'Delete a File Share with [FILE SHARE ID]'
        self.handle_response(self.conn.delete(f'/FileSystems/{self.fs}/ExportedFileShares/{arg}'))

class Quick(Command):
    intro = 'Quick commands to do a bunch of things at once on the Rabbit'
    prompt = '(nnf)' + '(quick)'

    def preloop(self):
        # Check if the system is in a good state to do some quick commands
        pass

    def do_setup(self,arg):
        'Quickly setup a file system of type [TYPE]'
        capacity = 1024 * 1024 * 1024
        StoragePool(self.conn).onecmd(f'create {capacity}')
        StorageGroup(self.conn).onecmd('create 0 0')
        FileSystem(self.conn).onecmd(f'create {arg} test 0')
        FileShare(self.conn, '0').onecmd('create 0 /mnt/test')

    def do_teardown(self,arg):
        'Quickly teardown a system'
        StoragePool(self.conn).onecmd('delete 0')

class Main(Command):
    intro = 'Command Interpreter for NNF Storage Element Controller'
    prompt = '(nnf)'

    def preloop(self):
        response = self.conn.get('')
        if response.ok:
            print('Connection Established. Starting Program...')

    def do_pool(self, arg):
        'Storage Pool Commands'
        StoragePool(self.conn).cmdloop()
    def do_server(self,arg):
        'Server Endpoint Commands'
        ServerEndpoint(self.conn).cmdloop()
    def do_group(self,arg):
        'Storage Group Commands'
        StorageGroup(self.conn).cmdloop()
    def do_fs(self,arg):
        'File System Commands'
        FileSystem(self.conn).cmdloop()

    def do_quick(self,arg):
        'Quick commands to setup and teardown a bunch of things at once'
        Quick(self.conn).cmdloop()

if __name__ == '__main__':

    parser = argparse.ArgumentParser(description='Interactive tool for NNF Element Controller')
    connection.addServerArguments(parser)
    conn = connection.connect(parser.parse_args(), '/redfish/v1/StorageServices/NNF')


    Main(conn).cmdloop()

