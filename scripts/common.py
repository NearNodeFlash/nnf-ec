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

import argparse, re, math

def ByteSizeStringToIntegerBytes(arg: str):
    m = re.match('(\d+)([kKMGTP]i?)B?$', arg)
    if m:
        powers = {'K': math.pow(10, 3), 'M': math.pow(10, 6), 'G': math.pow(10,9), 'T': math.pow(10,12), 'P': math.pow(10,15),
                    'Ki': math.pow(2, 10), 'Mi': math.pow(2,20), 'Gi': math.pow(2,30), 'Ti': math.pow(2,40), 'Pi': math.pow(2,50) }
        val = int(m.group(1))
        return int(val * powers[m.group(2)])
    return int(arg)

class ByteSizeAction(argparse.Action):
    def __init__(self, option_strings, *args, **kwargs):
        super(ByteSizeAction, self).__init__(option_strings=option_strings, *args, **kwargs)

    def __call__(self, parser, namespace, values, option_string=None):
        setattr(namespace, self.dest, ByteSizeStringToIntegerBytes(values))
