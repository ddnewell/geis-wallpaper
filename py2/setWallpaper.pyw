#!/usr/bin/python
'''
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
'''

if __name__ == "__main__":
    import GEIS, datetime
    p = GEIS.plot(configFile='json/config.json')
    p.createMap(resolution='c', projection='merc')
    p.loadImage(imageFile='NaturalEarth2_Mercator_Dell.png', zorder=1, origin='upper')
    p.loadImage(imageFile='wx.png', zorder=3, origin='upper', alpha=0.4)
    p.plotDaylightRadiation(interpolation='bicubic', zorder=2)
    p.plotTropicalWX("json/hurricane.json", txtX=0.68889)
    p.plotWorldtime("json/clocks.json")
    p.plotDaylightUpdateTime()
    p.setWallpaper(False)
