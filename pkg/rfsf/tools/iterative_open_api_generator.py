#!/usr/local/bin/python3
#
# Development script that iterative expands the provided
# yaml file from each path and then does a build as the
# paths are added.
#
# I used this script to find out where the schema is broken
# 
# Author: Nate Roiger
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#

import argparse, sys, subprocess, os

def prepare(args):

    if args.clean:
        subprocess.check_call(args=['rm', '-rf', args.out])

def run_generator(yaml, out, opts='--skip-validate-spec', silent=False):
    
    stdout = stderr = None
    if silent:
        stdout = stderr = subprocess.DEVNULL

    subprocess.check_call(args=f'openapi-generator generate -g go-server -i {yaml} -o {out} {opts}'.split(),
        stderr=stderr,
        stdout=stdout)

def run_evolve(args):
    
    tmp = f'{args.yaml}.tmp'
    subprocess.call(args=['rm', tmp,])

    first = False
    with open(args.yaml) as fp:
        with open(tmp, 'w+') as wp:
            for _, line in enumerate(fp):
                    

                if line.startswith('  /redfish/v1/'):
                    if not first:
                        if args.start == None or args.start in line:
                            first = True
                        wp.write(line)
                        continue

                    print(line)

                    wp.flush()
                    run_generator(tmp, args.out, '', args.silent)


                wp.write(line)

                
            


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Run the gnerate code on the yaml')
    parser.add_argument('yaml')
    parser.add_argument('out')
    parser.add_argument('--start', type=str, default=None)
    parser.add_argument('--clean', action='store_true')
    parser.add_argument('--evolve', action='store_true')
    parser.add_argument('--silent', action='store_true')

    args = parser.parse_args()

    prepare(args)

    if args.evolve:
        run_evolve(args)
    else:
        run_generator(args.yaml, args.out)
    