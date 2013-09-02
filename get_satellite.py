#!/usr/bin/python
"""
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
"""

import requests, json


def retrieve_satellite(API_KEY, target_file, config_file):
    """Retrieve tropical satellite data from Wunderground"""
    # Load configuration
    cfg = {}
    try:
        c = json.load(open(config_file))
        cfg = c["config"]
    except:
        pass
    # Set configuration variables (or defaults)
    dpi = cfg['dpi'] if 'dpi' in cfg else 96
    screen_size = cfg['screen_size'] if 'screen_size' in cfg else (1366, 768)
    img_size = (screen_size[0], screen_size[0]/2)
    # Base API URL
    url = 'http://api.wunderground.com/api/{0}/satellite/image.png'.format(API_KEY)
    # url = 'http://api.wunderground.com/api/{0}//radar/satellite/image.png'.format(API_KEY)
    # API Parameters
    # params = {
    #     'rad.maxlat': 80,
    #     'rad.maxlon': 180,
    #     'rad.minlat': -80,
    #     'rad.minlon': -180,
    #     'rad.width': img_size[0],
    #     'rad.height': img_size[1],
    #     'rad.rainsnow': 1,
    #     'rad.smooth': 1,
    #     'rad.noclutter': 1,
    #     'rad.reproj.automerc': 1,
    #     'sat.maxlat': 80,
    #     'sat.maxlon': 180,
    #     'sat.minlat': -80,
    #     'sat.minlon': -180,
    #     'sat.width': img_size[0],
    #     'sat.height': img_size[1],
    #     'sat.key': 'sat_ir4_bottom',
    #     'sat.gtt': 107,
    #     'sat.proj': 'me',
    #     'sat.timelabel': 0
    # }
    params = {
        'maxlat': 89,
        'maxlon': 180,
        'minlat': -89,
        'minlon': -180,
        'width': img_size[0],
        'height': img_size[1],
        'key': 'sat_ir4_bottom',
        'gtt': 107,
        'timelabel': 1,
        'timelabel.x': img_size[0]*0.85,
        'timelabel.y': img_size[1]*0.05,
        'proj': 'll'
    }
    # Get image
    req = requests.get(url, params=params)#, stream=True)
    if not req.ok:
        return {'error': True, 'msg': 'Wunderground API image could not be loaded', 'status': req.status_code}
    # Save image to file
    # with open(target_file, 'wb') as f:
    #     for block in req.iter_content(1024):
    #         if not block:
    #             break
    #         f.write(block)
    with open(target_file, 'wb') as f:
        f.write(req.content)
    # Return complete
    return {'error': False, 'msg': 'Wunderground API image downloaded successfully', 'url': req.url, 'status': req.status_code}


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
    print retrieve_satellite(options.key, options.target, options.config)

