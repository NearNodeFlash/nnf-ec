#!/usr/local/bin/python3
#
# This script will patch the provided yaml file with fixes
# for known errors in the schema.
#
# This script should run before an openapi-generator call
# 
# Author: Nate Roiger
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#

import argparse

def fix_collection(ln):
    # Need to patch things like 
    #    ManagerAccountCollection_ManagerAccountCollection_ManagerAccountCollection
    #    Certificate_v1_2_1_Certificate_v1_2_1_Certificate
    #    Certificate_v1_2_1_Certificate_v1_2_1_RekeyRequestBody

    url = ln.split(':')[-1]
    name = url.split('/')[-1].rstrip()
    names = name.split('_')


    if 'v1' in name: # Versioned model
        model = '_'.join(names[0:4])
        if  name.count(model) == 2:
            new_name = f'{model}_{names[-1]}'
            return ln.replace(name, new_name)
        return ln

    else: # Collection type
        if len(names) == 2:
            return ln
        new_name = f'{names[0]}_{names[0]}'
        return ln.replace(name, new_name)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Find a generated model from the provided yaml')
    parser.add_argument('src')

    args = parser.parse_args()

    with open(args.src) as sp:
        
        for num, ln in enumerate(sp):
            try:
                
                if ln.lstrip().startswith('$ref:') and not ln.rstrip().endswith('\''):
                    ln = fix_collection(ln)
                
                # This is a known bad entry to a legacy swordfish schema definition
                if "redfish.dmtf.org/schemas/swordfish/v1/DriveCollection.yaml#/components/schemas/DriveCollection_DriveCollection" in ln:
                    ln = ln.replace('swordfish/', '')

                print(ln, end='')

            except Exception as e:
                print(num+1)
                raise e

        