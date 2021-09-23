#!/usr/local/bin/python3
#
# Script to patch Go constants in an given directory such that
# names align as needed. This is needed because the OpenAPI 
# Generator is currently broken when generating constants as 
# shown below:
#
#    type Resource_v1_10_0_Reference string
#
#    // List of Resource_v1_10_0_Reference
#    const (
#        TOP ResourceV1100Reference = "Top"
#        ...
#    )
#
# We wish to rename ResourceV1100Reference values to the
# type definition.
#
# Author: Nate Roiger
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#

import argparse, os, string, subprocess

def is_const_file(file):
    with open(file) as fp:
        for _, ln in enumerate(fp):
            if is_enum_type(ln):
                return True
    return False

def is_enum_type(ln):
    # Returns True if this is a enum declration i.e.
    #    type Chassis_v1_14_0_IndicatorLED string
    return ln.startswith('type') and ln.endswith('string\n')

def get_enum_name(ln):
    # Returns the enum type's name i.e. Chassis_v1_14_0_IndicatorLED
    #    type Chassis_v1_14_0_IndicatorLED string
    return ln.split()[1]

# this will return change types of invalid name: 
#   ManagerAccount_v1_6_2_AccountTypes
# into
#     ManagerAccountV162AccountTypes
def make_type_name(ln):
    # Returns a valid type name from the given line
    #     type Chassis_v1_14_0_IndicatorLED string
    # ->
    #      ChassisV1140IndicatorLED string
    name = get_enum_name(ln)
    name = name.replace('v1', 'V1', 1)
    name = name.replace('_', '')
    return name

def make_type_suffix(ln):
    # Returns a suffix from the given line
    #    type Chassis_v1_14_0_IndicatorLED string
    # ->
    #    CV1140ILED
    name = get_enum_name(ln)

    
    name = name.replace('v1', 'V1', 1) # always capitalize version 
    name = name.translate(str.maketrans('', '', string.ascii_lowercase)).replace('_', '') # drop all lowercase letters

    # Special kind of hack for ending in Status & State. Make state unique.
    if get_enum_name(ln).endswith('State'):
        name = name + 'T'

    return name

def get_const_name(ln):
    # Returns the constant name i.e. UNKNOWN
    #    UNKNOWN ChassisV1140IndicatorLed = "Unknown"
    return ln.lstrip().split(' ')[0]

def get_const_type(ln):
    # Returns the constant type i.e. ChassisV1140IndicatorLed
    #    UNKNOWN ChassisV1140IndicatorLed = "Unknown"
    return ln.lstrip().split(' ')[1]

def is_const_type(ln):
    # Returns TRUE if const type detected on line
    #    UNKNOWN ChassisV1140IndicatorLed = "Unknown"
    return get_const_name(ln).isupper() or get_const_name(ln).replace('_', '').isnumeric()

def make_const_type(ln, name, suffix):
    cnst = get_const_name(ln)
    ln = ln.replace(cnst, cnst + '_' + suffix, 1)
    ln = ln.replace(get_enum_name(ln), name, 1)
    return ln
    

def patch_enums(file):
    src_name, dest_name = None, None

    lines = []
    with open(file) as fp:
        for _, ln in enumerate(fp):
            if is_enum_type(ln):
                
                name = make_type_name(ln)
                suffix = make_type_suffix(ln)

                dest_name = name
                
                ln = ln.replace(get_enum_name(ln), name)
                
            if is_const_type(ln):
                # Avoid patching an existing patch
                if not get_const_type(ln).endswith(suffix):
                    src_name = get_const_type(ln)
                    ln = make_const_type(ln, name, suffix)
            lines.append(ln)
            
    with open(file, 'w') as fp:
        fp.writelines(lines)

    return src_name, dest_name

def patch_models(src, dst, dir):
    
    if src == dst:
        return
    print(f'  Patching models from {src} to {dst}')
    for _, _, filenames in os.walk(dir):
        for file in filenames:
            path = os.path.join(dir, file)
            subprocess.check_call(
                args=['sed',
                '-i',
                '',
                's/' + src + '/' + dst + '/g',
                path]
            )

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Patch go constants found in a given directory')
    parser.add_argument('dir', help='directory to analyze and patch constants')

    args = parser.parse_args()
    
    for _, _, filenames in os.walk(args.dir):
        for file in filenames:
            path = os.path.join(args.dir, file)
            if is_const_file(path):
                print("Found patch file {0}".format(path))
                src, dst = patch_enums(path)

                patch_models(src, dst, args.dir)
