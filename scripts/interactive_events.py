#!/usr/bin/env python3
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

import argparse, json, os, time, math

from datetime import datetime
from collections import OrderedDict

import http.server

import connection
from interactive import Command

class SubscriptionHandler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        data = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        self.send_response(200)
        self.end_headers()

        for event in data['Events']:
            Main.print_event(event, '--human-readable')


class Main(Command):
    intro = 'Get/List/Subscribe to Events'
    prompt = '(nnf)'

    def preloop(self):
        response = self.conn.get('')
        if response.ok:
            print('Connection Established. Starting Program...')

    def do_sub(self, arg):
        'Subscribe to Event Stream'

        port = 8090
        payload = {
            'Context': 'Python Event Monitor',
            'Destination': f':{port}',
            'DeliveryRetryPolicy': 'TerminateAfterRetries',
        }
        self.handle_response(self.conn.post('/Subscriptions', payload))

        print('Starting HTTP server...')
        with http.server.HTTPServer(("", port), SubscriptionHandler) as httpd:
            try:
                httpd.serve_forever()
            except KeyboardInterrupt:
                pass

    def do_list_subs(self, arg):
        'List Event Subscriptions'
        self.handle_response(self.conn.get('/Subscriptions'))

    def do_get_subs(self, arg):
        'Get Event Subscription'
        self.handle_response(self.conn.get(f'/Subscriptions/{arg}'))

    def do_list(self, arg):
        'List Events'
        self.handle_response(self.conn.get('/Events'))

    def do_get(self, arg):
        'Get Event [-H] [-all]'
        args = arg.split(' ')
        if 'all' in args or '--all' in arg:
            self.get_all_events(args)
        else:
            self.handle_response(self.conn.get(f'/Events/{arg}'))

    def get_all_events(self, args):
        response = self.conn.get('/Events')
        if not response.ok:
            print('Failed to retrieve response')
            return

        events = response.json()
        for e in events['Members']:
            id = e['@odata.id'].split('/')[-1]
            event = self.conn.get(f'/Events/{id}')
            if not event.ok:
                print(f'!!! Failed to retrieve event {id} !!!')
                return
            self.print_event(event.json(), args)

    @staticmethod
    def print_event(e, args):
        if "-H" in args or "--human-readable" in args:
            msg = e['Message']
            if 'MessageArgs' in e:
                for idx, arg in enumerate(e['MessageArgs']):
                    msg = msg.replace(f'%{idx+1}', arg)

            origin = 'unspecified'
            if 'OriginOfCondition' in e:
                origin = e['OriginOfCondition']['@odata.id']
            print(f'{e["EventId"]: <3s} : {e["MessageSeverity"]: >7s} : {msg} : {origin}')
        else:
            print(e)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Interactive tool for NNF Event Service')
    connection.addServerArguments(parser)
    conn = connection.connect(parser.parse_args(), '/redfish/v1/EventService')

    Main(conn).cmdloop()