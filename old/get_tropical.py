#!/usr/bin/python
"""
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
"""

import requests, json


def retrieve_tropical_wx(API_KEY, target_file):
    """Retrieve tropical weather data from Wunderground"""
    # Base API URL
    url = 'http://api.wunderground.com/api/{0}/currenthurricane/view.json'.format(API_KEY)
    # Get data
    resp = requests.get(url)
    if not resp.ok:
        return {'error': True, 'msg': 'Could not load Wunderground API data'}
    try:
        data = resp.json()
    except:
        return {'error': True, 'msg': 'Could not parse Wunderground API data'}
    # Save data to file
    with open(target_file, 'w') as f:
        json.dump(data, f)
    # Return complete
    return {'error': False, 'msg': 'Wunderground API data downloaded successfully'}


if __name__ == '__main__':
    # Import command line argument parser
    from optparse import OptionParser
    # Parse for options
    parser = OptionParser()
    parser.add_option("-k", "--key", dest="key", help="Wunderground API key")
    parser.add_option("-t", "--target", dest="target", help="Target image file")
    parser.add_option("-c", "--config", dest="config", help="Configuration file")
    (options, args) = parser.parse_args()
    # Retrieve tropical weather data from Wunderground API
    retrieve_tropical_wx(options.key, options.target)

