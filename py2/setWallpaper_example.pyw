#!/usr/bin/python
'''
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
'''

if __name__ == "__main__":
    print "Importing... ",
    import GEIS, datetime
    print "complete."
    print "Plot object... ",
    p = GEIS.plot()
    print "complete."
    print "Map object... ",
    p.createMap(resolution='c', projection='merc')
    print "complete."
    # print "Blue Marble Background... ",
    # p.setBlueMarbleBackground()
    # print "complete."
    # print "Coloring land and ocean... ",
    # p.map.drawlsmask(land_color='#ffffbf', ocean_color='#91bfdb', lakes=True,zorder=1)
    # print "complete."
    # print "Drawing countries... ",
    # p.map.drawcountries(linewidth=0.5, color='#fee090', antialiased=1,zorder=3)
    # print "complete."
    print "Loading base map... ",
    #p.loadImage(imageFile='baseMap.png', zorder=0, origin='upper')
    # p.loadImage(imageFile='NaturalEarth2_Mercator.png', zorder=0, origin='upper')
    p.loadImage(imageFile='NaturalEarth2_Mercator_Dell.png', zorder=0, origin='upper')
    #p.loadImage(imageFile='blueMarbleBaseMap.png', zorder=0, origin='upper')
    print "complete."
    # print "Drawing coastlines... ",
    # p.map.drawcoastlines(linewidth=0.75, color='#fee090', antialiased=1,zorder=3)
    # print "complete."
    # print "Drawing Rivers... ",
    # p.map.drawrivers(linewidth=0.1, color='#91bfdb', antialiased=1,zorder=2)
    # print "complete."
    # print "Plotting daylight... ",
    # p.plotDaylight(alpha=0.1, zorder=5)
    # print "complete."
    print "Plotting daylight radiation... ",
    p.plotDaylightRadiation(interpolation='bicubic', zorder=6)
    print "complete."
    # print "Plotting equator... ",
    # p.map.drawparallels([0], color='#e34a33', linewidth=0.75)
    # print "complete."
    print "Plotting Times... ",
    p.plotWorldtime("json/clocks.json")
    p.plotDaylightUpdateTime()
    print "complete."
    print "Setting wallpaper... ",
    p.setWallpaper(False)
    print "complete."
