#!/usr/bin/env python3

import argparse, json, os, time, math

from datetime import datetime
from collections import OrderedDict

import connection
from interactive import Command





class Definitions(Command):
    intro = 'Get/List Metric Definitions'
    prompt = '(nnf)' + '(definitions)'

    def do_list(self, arg):
        'List Metric Definitions'
        self.handle_response(self.conn.get('/MetricDefinitions'))
    
    def do_get(self, arg):
        'Get Metric Definition'
        self.handle_response(self.conn.get(f'/MetricDefinitions/{arg}'))

class ReportDefinitions(Command):
    intro = 'Get/List/Edit Metric Report Definitions'
    prompt = '(nnf)' + '(report definitions)'

    def do_list(self,arg):
        'List Metric Report Definitions'
        self.handle_response(self.conn.get('/MetricReportDefinitions'))

    def do_get(self, arg):
        'Get Metric Report Definitions'
        self.handle_response(self.conn.get(f'/MetricReportDefinitions/{arg}'))

    def do_schedule(self, arg):
        'Change the schedule of [METRIC REPORT DEFINITION ID] (NOT YET IMPLEMENTED)'
        # TODO


class Reports(Command):
    intro = 'Get/List/Monitor Metric Report'
    prompt = '(nnf)' + '(report)'

    def do_list(self,arg):
        'List Metric Reports'
        self.handle_response(self.conn.get('/MetricReports'))

    def do_get(self,arg):
        'Get Metric Report'
        self.handle_response(self.conn.get(f'/MetricReports/{arg}'))

class Monitor(Command):
    intro = 'Monitor Metrics for things like Bandwidth'
    prompt = '(nnf)' + '(monitor)'

    def do_bandwidth(self,arg):
        'Bandwidth Monitor'

        if arg == '':
            id, options = 'SwitchPortTxRx', []
        elif len(arg.split()) == 1:
            id, options = arg, []
        else:
            id, options = arg.split(maxsplit=1)
            id = id[0]

        # Get the report definition - it should have the wildcards
        # we can use the parse the report
        response = self.conn.get(f'/MetricReports/{id}')
        if not response.ok:
            print(f'Failed to retrive metric {id}: Status Code: {response.status_code}')
            return

        report = response.json()

        defId = report['MetricReportDefinition']['@odata.id'].split('/')[-1]
        response = self.conn.get(f'/MetricReportDefinitions/{defId}')
        if not response.ok:
            print(f'Failed to retrieve metric report definition {defId}: Status Code: {response.status_code}')
            return

        definition = response.json()

        # Metrics are grouped by wildcards, and we should find the combination of
        # wildcards in the report. Use this to monitor the metric as they are read.

        wildcards = definition['Wildcards']
        props = definition['MetricProperties']

        def enumerateWildcard(prop, wildcards):
            wildcard = wildcards[0]
            returnProps = []
            for val in wildcard['Values']:
                newProp = prop.replace('{' + wildcard['Name'] + '}', val)
                if len(wildcards) > 1:
                    props = enumerateWildcard(newProp, wildcards[1:])
                    returnProps += props
                else:
                    returnProps.append(newProp)
                
            return returnProps

        def parseTimestamp(timestamp):
            if timestamp[-1] == 'Z':
                return datetime.strptime(timestamp[0:-4], "%Y-%m-%dT%H:%M:%S.%f")
            return datetime.strptime(timestamp, "%Y-%m-%dT%H:%M:%S.%f%z")

        metrics = {}
        for prop in props:
            for property in enumerateWildcard(prop, wildcards):
                metrics[property] = {}
        
        for value in report['MetricValues']:
            metrics[value['MetricProperty']] = {
                'Previous': int(value['MetricValue']), 
                'Timestamp': parseTimestamp(value['Timestamp']),
                'Throughput': 0}

        # Now, retrieve the properties every so often and display them on the screen
        print('Starting Bandwidth Monitor...CTRL+C to exit')
        while True:
            time.sleep(5)
            
            response = self.conn.get(f'/MetricReports/{id}')
            if not response.ok:
                print(f'Failed to retrive metric {id}: Status Code: {response.status_code}')
                return

            report = response.json()

            refresh = False
            for value in report['MetricValues']:
                prop = value['MetricProperty']
                timestamp = parseTimestamp(value['Timestamp'])
                if metrics[prop]['Timestamp'] != timestamp:
                    deltaBytes = int(value['MetricValue']) - metrics[prop]['Previous']
                    elapsedSeconds = (timestamp - metrics[prop]['Timestamp']).total_seconds()
                    metrics[prop]['Throughput'] = float(deltaBytes) / elapsedSeconds

                    metrics[prop]['Previous'] = int(value['MetricValue'])
                    metrics[prop]['Timestamp'] = timestamp

                    refresh = True

            def getPortMetrics(switchId, portId):
                def suffixGet(rate):
                    siSuffixes = [
                        [1e15, 'P'],
                        [1e12, 'T'],
                        [1e9,'G'],
                        [1e6,'M'],
                        [1e3,'K'],
                        [1e0,''],
                        [1e-3,'m'],
                        [1e-6,'u'],
                        [1e-9,'n'],
                        [1e-12,'p'],
                        [1e-15,'f'],
                    ]

                    for magnitude, suffix in siSuffixes:
                        if rate > magnitude:
                            rate = rate / magnitude
                            return rate, suffix
                    return rate, ''

                            #/redfish/v1/Fabrics/Rabbit/Switches/1/Ports/17/Metrics/TxBytes
                rx = metrics[f'/redfish/v1/Fabrics/Rabbit/Switches/{switchId}/Ports/{portId}/Metrics/RxBytes']['Throughput']
                rxrate, rxsuffix = suffixGet(rx)
                tx = metrics[f'/redfish/v1/Fabrics/Rabbit/Switches/{switchId}/Ports/{portId}/Metrics/TxBytes']['Throughput']
                txrate, txsuffix = suffixGet(tx)
                return f'{rxrate:5.1f}{rxsuffix}B/s', f'{txrate:5.1f}{txsuffix}B/s'

            def display_port_metric(switchId, portId):
                rx, tx = getPortMetrics(switchId, portId)
                print(f'{portId:<2} {rx:>12} {tx:>12}')
                
            def display_switch_metrics(switchId):
                title = f'Switch {switchId}'
                print(f'{title:=^28}')
                #      ==========Switch 0==========
                #      0    5210.2PB/s   5210.2PB/s
                print("Port    RxBytes      TxBytes")
                for portId in range(19):
                    display_port_metric(switchId, portId)
                
            if refresh:
                os.system('cls' if os.name == 'nt' else 'clear')
                for switchId in [0,1]:
                    display_switch_metrics(switchId)

class Main(Command):
    intro = 'Get/List/Edit/Monitor Metric Definitions and Data'
    prompt = '(nnf)' 

    def preloop(self):
        response = self.conn.get('')
        if response.ok:
            print('Connection Established. Starting Program...')

    def do_def(self,arg):
        'Metric Definitions'
        Definitions(self.conn).cmdloop()

    def do_reportdef(self,arg):
        'Metric Report Definitions'
        ReportDefinitions(self.conn).cmdloop()

    def do_report(self,arg):
        'Metric Reports'
        Reports(self.conn).cmdloop()
    
    def do_monitor(self,arg):
        'Monitor Tools'
        Monitor(self.conn).cmdloop()
    


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Interactive tool for NNF Telemetry Manager')
    connection.addServerArguments(parser)
    conn = connection.connect(parser.parse_args(), '/redfish/v1/TelemetryService')

    Main(conn).cmdloop()