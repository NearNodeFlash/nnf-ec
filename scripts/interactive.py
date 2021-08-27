#!/usr/bin/env python3

import cmd, argparse, json

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
        return True

class StoragePool(Command):
    intro = 'Create/Get/List/Delete Storage Pools'
    prompt = '(nnf)' + '(storage pool)'

    def do_create(self, arg):
        'Create a Storage Pool of specific [SIZE = 1GB]'

        if arg is None or arg == "":
            arg = '1GB'
            print(f"No size specified - defaulting to {arg}")
            
        try:
            arg = ByteSizeStringToIntegerBytes(arg)
            payload = {
                'Capacity': {'Data': {'AllocatedBytes': int(arg)}},
                'Oem': {'Compliance': 'relaxed'}
            }
        except ValueError:
            print('*** SIZE argument should be integer')
            return
        
        self.handle_response(self.conn.post('/StoragePools', payload))

    def do_get(self, arg):
        'Get a Storage Pool by POOL ID'
        self.handle_response(self.conn.get(f'/StoragePools/{arg}'))

    def do_list(self,arg):
        'List Storage Pools'
        self.handle_response(self.conn.get(f'/StoragePools'))

    def do_delete(self, arg):
        'Delete a Storage Pool by POOL ID'
        self.handle_response(self.conn.delete(f'/StoragePools/{arg}'))

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
    intro = 'Create/Get/List/Delete Storage Groups'
    prompt = '(nnf)' + '(storage group)'

    def do_create(self,arg):
        'Create a Storage Group from [STOARGE POOL ID] [SERVER ENDPOINT ID]'
        
        try:
            pool, server = arg.split()[:2]
        except ValueError:
            print('Expected two arguments [STOARGE POOL ID] [SERVER ENDPOINT ID]')
            return

        payload = {
            'Links': {
                'StoragePool': { '@odata.id': f'{self.conn.base}/StoragePools/{pool}'},
                'ServerEndpoint': { '@odata.id': f'{self.conn.base}/Endpoints/{server}'}
            }
        }

        self.handle_response(self.conn.post(f'/StorageGroups', payload))

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
    intro = 'Create/Get/List/Delete File Systems'
    prompt = '(nnf)' + '(file system)'

    def do_create(self,arg):
        'Create a File System of [TYPE] [NAME] [STORAGE POOL ID]'

        try:
            typ, name, pool = arg.split()[:3]
        except ValueError:
            print('Expected three arguments [TYPE] [NAME] [STORAGE POOL ID]')
            return

        payload = {
            'Links': { 'StoragePool': {'@odata.id': f'{self.conn.base}/StoragePools/{pool}'}},
            'Oem': { 'Type': typ, 'Name': name}
        }

        self.handle_response(self.conn.post(f'/FileSystems', payload))
    
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
    intro = 'Create/Get/List/Delete File Shares'
    prompt = '(nnf)' + '(file system)' + '(file share)'

    def __init__(self, conn, fs, **kwargs):
        self.fs = fs
        super().__init__(conn, **kwargs)

    def do_create(self,arg):
        'Create a File Share to a [SERVER ENDPOINT] with [MOUNTPOINT]'

        try:
            server, mount = arg.split()[:2]
        except ValueError:
            print('Expected three arguments [SERVER ENDPOINT ID] [MOUNTPOINT]')
            return

        payload = {
            'Links': { 'Endpoint': {'@odata.id': f'{self.conn.base}/Endpoints/{server}'}},
            'FileSharePath': mount
        }

        self.handle_response(self.conn.post(f'/FileSystems/{self.fs}/ExportedFileShares', payload))

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
    
