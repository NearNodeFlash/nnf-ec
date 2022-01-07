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