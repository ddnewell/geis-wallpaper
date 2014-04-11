#!/usr/bin/python
"""
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
"""

if __name__ == '__main__':
    import wmap, os
    p = wmap.Plot(config_file='json/config.json', save_file='wallpaper.png')
    p.create_map()
    try:
        p.load_image(imageFile='img/NaturalEarth_Mac13Retina.png', zorder=1, origin='upper', extent=[-180, 180, -90, 90])
    except:
        pass
    p.update_satellite(imageFile='data/wx.png')
    try:
        p.load_image(imageFile='data/wx.png', zorder=3, origin='upper', alpha=0.35, extent=[-180, 180, -89, 89])
    except:
        pass
    p.plot_daylight(zorder=2)
    p.plot_tropical_wx('json/hurricane.json')
    p.plot_worldtime('json/clocks.json')
    p.plot_daylight_update_time()
    p.set_wallpaper()

