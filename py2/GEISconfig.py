'''
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
'''

# Blue marble scale
mapScale = 0.35
# Latitude ranges
maxLat = 71.0
minLat = -71.0
latRange = maxLat - minLat
latIntervalRange = 3.9
# latInterval = 0.01
latInterval = 0.05
latOffset = 1.0
# Longitude ranges
maxLon = 180.0
minLon = -180.0
lonRange = maxLon - minLon
# lonInterval = 0.591
lonInterval = 1.5
# Screen size in pixels
# screenSize = (1440, 900)
screenSize = (1366, 768)
# Plot size and resolution
dpi = 96
plotSize = (screenSize[0]/dpi,  screenSize[1]/dpi)
# Width and height adjustments
widthAdj = 0.5
heightAdj = 0.5
# Terminator properties
nightTransparency = 0.25
termResolution = 0.05
transparencyRange = (0.3, 0.0)
