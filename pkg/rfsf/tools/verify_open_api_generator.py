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

import argparse, os, collections

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Find a generated model from the provided yaml')
    parser.add_argument('yaml', help='the yaml file containing the definitions')
    parser.add_argument('dir', help='the directory containing the models')
    parser.add_argument('-v', '--verbose', action='store_true', help='verbose debug output')

    args = parser.parse_args()

    yaml = collections.OrderedDict()

    print(f'finding yaml definitions for {args.yaml}')
    with open(args.yaml) as fp:

        in_get = False
        endpoint, line = None, None
        index = 0

        for lnum, ln in enumerate(fp):
            if ln.startswith('  /redfish'):
                line = lnum + 1
                endpoint = ln.rstrip(':\n').split('/')[-1]
            if ln.startswith('    get:'):
                in_get = True
            if in_get and ln.startswith('                $ref:'):
                # Schema reference is something like 'ServiceRoot_v1_9_0_ServiceRoot'
                index = index + 1
                schema = ln.split('/')[-1].strip()
                yaml[schema] = {
                    'url': ln.split(':')[1].strip(),
                    'index': index,
                    'endpoint': endpoint,
                    'line': line
                }
                yaml.move_to_end(schema, last = True)
                in_get = False

    print(f'finding model definitions in {args.dir}')

    models = []
    for _, _, files in os.walk(args.dir):
        for file in files:
            def _format(s):

                if 'v1' not in s:
                    return ''

                # Model name is something like 'model_service_root_v1_9_0_service_root.go'
                schema = s[len('model_'):-len('.go')]

                # everything to the left of 'v1' is the mode name 'service_root'
                name = schema[0:schema.index('v1')-1]

                # extract the versions
                version = '_'.join(schema.replace(name, '').split()[0:3]).strip('_')

                reg_name = name.replace('_', ' ').title().replace(' ', '')

                if args.verbose:
                    print(f'{schema} {name} {version}')

                # format the string
                return f'{reg_name}_{version}_{reg_name}'

            models.append(_format(file))


    for key, value in yaml.items():

        if key not in models:

            print(f'{value["index"]: <3} {key: <80}: line: {value["line"]}  endpoint: {value["endpoint"]}')





