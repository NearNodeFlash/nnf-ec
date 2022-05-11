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
import requests

def addServerArguments(parser):

    parser.add_argument('--host', type=str, default='localhost', help='specify a server endpoint')
    parser.add_argument('--port', type=int, default=8080, help='specify a server port')

class Connection:
    def __init__(self, host, port, base):
        self.host = host
        self.port = port
        self.base = base
        self.path = f'http://{host}:{port}{base}'
        self.paths = [base,]

    def get(self, odataid):
        return requests.get(self.path + odataid)

    def put(self, odataid, json):
        return requests.put(self.path + odataid, json=json)
        
    def post(self, odataid, json):
        return requests.post(self.path + odataid, json=json)

    def patch(self, odataid, json):
        return requests.patch(self.path + odataid, json=json)

    def delete(self, odataid):
        return requests.delete(self.path + odataid)

    def push_path(self, path):
        self.path = f'http://{self.host}:{self.port}{path}'
        self.paths.append(path)

    def pop_path(self):
        self.paths.pop()
        path = self.paths[-1]
        self.path = f'http://{self.host}:{self.port}{path}'

def connect(args, base):
    return Connection(args.host, args.port, base)