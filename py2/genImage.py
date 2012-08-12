#!/usr/bin/python
'''
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
'''

if __name__ == '__main__':
    print
    print "Importing... ",
    import GEIS,datetime
    print "complete."
    print "Plot object... ",
    p = GEIS.Plot.plot(filename='AirRoutes.png')
    p.screenSize = (2000,1250)
    print "complete."
    print "Map object... ",
    p.createMap(resolution='c',projection='merc')
    print "complete."
#    print "Blue Marble Background... ",
#    p.setBlueMarbleBackground()
#    print "complete."
#    print "Coloring land and ocean... ",
#    p.map.drawlsmask(land_color='#ffffbf',ocean_color='#91bfdb',lakes=True,zorder=1)
#    print "complete."
#    print "Drawing countries... ",
#    p.map.drawcountries(linewidth=0.5, color='#fee090', antialiased=1,zorder=3)
#    print "complete."
    #print "Loading base map... ",
    #p.loadImage(imageLoc='baseMap.png',zorder=0,origin='upper')
    #p.loadImage(imageLoc='blueMarbleBaseMap.png',zorder=0,origin='upper')
    #print "complete."
#    print "Drawing coastlines... ",
#    p.map.drawcoastlines(linewidth=0.75, color='#fee090', antialiased=1,zorder=3)
#    print "complete."
#    print "Drawing Rivers... ",
#    p.map.drawrivers(linewidth=0.1, color='#91bfdb', antialiased=1,zorder=2)
#    print "complete."
#    print "Plotting daylight... ",
#    p.plotDaylight(alpha=0.1,zorder=5)
#    print "complete."
    #print "Plotting daylight radiation... ",
    #p.plotDaylightRadiation(interpolation='bicubic',zorder=6)
    #print "complete."
#    print "Plotting equator... ",
#    p.map.drawparallels([0],color='#e34a33',linewidth=0.75)
#    print "complete"
#    print "Plotting Times... ",
#    p.plotWorldtime()
#    p.addText(lon=-177.,lat=-70.,text='Daylight Updated:  %s' % (p.localNow.strftime('%B %d, %Y  %I:%M%p'),),\
#              color='r',family='serif',weight='semibold',alpha=0.95,size=10,zorder=10,va='center')
#    print "complete."
#    print "Setting wallpaper... ",
#    p.setWallpaper()
#    print "complete."
    print "Saving image... ",
    p.saveMap(facecolor='#be5600')
    print "complete."
    print

