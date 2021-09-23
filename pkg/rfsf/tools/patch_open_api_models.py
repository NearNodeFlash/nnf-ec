#!/usr/local/bin/python3
#
# This script will patch the files in the provided directory
# to ensure they reference the Open API definitions
# 
# Author: Nate Roiger
#
# Copyright 2020 Hewlett Packard Enterprise Development LP
#

import argparse, os

def patch_line_tokens(ln, tokens):
    for i, token in enumerate(tokens):
        if token[0].isupper():
            ln = ln.replace(token, "openapi." + token)
            print(ln, end='')
    return ln

def patch_declare(ln):
    # Handles patches to function declartions i.e.
    #  RedfishV1AccountServiceAccountsPost(openapi.ManagerAccountV162ManagerAccount) (interface{}, error)
    #  RedfishV1AccountServiceAccountsManagerAccountIdPut(string, openapi.ManagerAccountV162ManagerAccount)
    tokens = ln.strip().split('(', 1)[1:2] + ln.strip().split()[1:]
    return patch_line_tokens(ln, tokens)

def patch_function(ln):
    # Handles patches to function implementations i.e.
    #  func (s *DefaultApiService) RedfishV1AccountServiceAccountsManagerAccountIdPatch(managerAccountId string, managerAccountV162ManagerAccount ManagerAccountV162ManagerAccount)
    print(ln)
    if '()' in ln:
        return ln
    tokens = ln.split('(', 2)[2].split()[1:] + ln.rstrip().split()[6::2]
    return patch_line_tokens(ln, tokens)

def patch_variable(ln):
    # Handles patches to variable delertions i.e.
    #  certificateV121RenewRequestBody := &CertificateV121RenewRequestBody{}
    var = ln.lstrip().split()[0]
    if var == 'body':
        return ln

    struct = ln.rstrip().split()[-1].lstrip('&').rstrip('{}')

    if var.lower() != struct.lower():
        return ln
    print(ln)
    if struct.startswith('openapi.'):
        return ln

    return ln.replace(struct, 'openapi.' + struct)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='patch .go files with openapi package')
    parser.add_argument('dir', help='directory for which files should be patched')
    parser.add_argument('--file', default=None, help='restrict parsing to a particular file')

    args = parser.parse_args()

    for _, _, files in os.walk(args.dir):
        for file in files:
            
            if args.file and file != args.file:
                continue

            lines = []
            with open(os.path.join(args.dir, file)) as fp:
                in_import = False
                for _, ln in enumerate(fp):
                    if ln.lstrip().startswith('RedfishV1'):
                        ln = patch_declare(ln)
                    elif ln.startswith('func (s *DefaultApiService) RedfishV1'):
                        ln = patch_function(ln)
                    elif ln.endswith('{}\n'):
                        ln = patch_variable(ln)
                    elif ln.strip() == 'import (':
                        in_import = True
                    elif in_import:
                        if 'openapi' in ln:
                            in_import = False
                        if ln.strip() == ')':
                            ln = '\topenapi "stash.us.cray.com/rabsw/nnf-ec/pkg/rfsf/pkg/models"\n' + ln
                    lines.append(ln)

            with open(os.path.join(args.dir, file), 'w') as fp:
                fp.writelines(lines)

