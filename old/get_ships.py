#!/usr/bin/python
"""
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2014 Newell Designs, David Newell.
"""

import requests, json


def retrieve_ship_locations(API_KEY, target_file):
    """Retrieve ship location data from Fleetmon"""
    # Base API URL
    # url = 'http://www.fleetmon.com/api/p/personal-v1/myfleet/?username=ddnewell&api_key={0}&format=json'.format(API_KEY)
    url = 'http://www.marinetraffic.com/en/ais/details/ships/9319753/vessel:TOMBARRA'
    # Get data
    resp = requests.get(url, headers={'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.101 Safari/537.36'})
    if not resp.ok:
        return {'error': True, 'msg': 'Could not load Fleetmon API data'}
    try:
        # data = resp.json()
        raw_data = resp.text
    except:
        return {'error': True, 'msg': 'Could not parse Fleetmon API data'}
    # Save data to file
    rd1 = raw_data.split('centerx:')[1]
    rd2 = rd1.split('/centery:')
    lon = float(rd2[0])
    lat = float(rd2[1].split('/zoom')[0])
    data = {
        'objects': [
            {'vessel': {
                    'latitude': lat,
                    'longitude': lon
                }
            }
        ]
    }
    with open(target_file, 'w') as f:
        json.dump(data, f)
    # Return complete
    return {'error': False, 'msg': 'Fleetmon API data downloaded successfully'}


if __name__ == '__main__':
    # Import command line argument parser
    from optparse import OptionParser
    # Parse for options
    parser = OptionParser()
    parser.add_option("-k", "--key", dest="key", help="Fleetmon API key")
    parser.add_option("-t", "--target", dest="target", help="Target json file")
    parser.add_option("-c", "--config", dest="config", help="Configuration file")
    (options, args) = parser.parse_args()
    # Retrieve ship location data from Fleetmon API
    retrieve_ship_locations(options.key, options.target)

