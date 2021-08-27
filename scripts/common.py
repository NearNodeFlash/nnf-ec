
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
