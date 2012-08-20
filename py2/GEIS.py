#!/usr/bin/python
'''
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
'''

# Import useful modules

# --------------------------------------------------------

# Import correct division
from __future__ import division

# Import sys module and ensure path is set
import sys
if 'C:\\Python27\\Lib\\site-packages\\Pysolar' not in sys.path:
    sys.path.append('C:\\Python27\\Lib\\site-packages\\Pysolar')

# Import operating system,  and date/time modules
import os,  datetime, time

# Import json module
import json

# Import Basemap,  numpy,  and matplotlib
from mpl_toolkits.basemap import Basemap
import numpy as np
import matplotlib.pyplot as plt
from matplotlib import cm

# Import PIL for cropping and saving temporary file as bitmap
from PIL import Image

# Import terminator calculation module
import terminator

# Import Windows module (for setting wallpaper)
import ctypes

# Import time zone database
import pytz

# Import math module
import math

# --------------------------------------------------------


class plot:
    """
    Create map with options of daylight and other points and set as wallpaper (Windows only)
    """


    def __init__(self, now=None, filename=None, configFile=None):
        '''
        Constructor - Initialize plotting setup
        '''
        # If no configuration file specified, raise error
        if configFile == None:
            raise Exception('No configuration file specified!')
        # Get current time, if not specified,  set to UTC now
        if now == None or type(now) != datetime.datetime:
            self.utcNow = datetime.datetime.utcnow()
        else:
            self.utcNow = now
        # Setup date & time
        self.utcNowNaive = self.utcNow
        self.utcNow = self.utcNow.replace(tzinfo=pytz.utc)
        self.localNow = datetime.datetime.now()
        
        # Terminator objects
        self.irr = terminator.irradiation(now=self.utcNow)
        self.term = terminator.terminator(now=self.utcNow)
        
        # Configuration defaults
        self.cfg = {
            "blueMarbleMapScale": 0.35,
            "latIntervalRange": 3.9,
            "latInterval": 0.05,
            "latOffset": 1.0,
            "lonInterval": 1.5,
            "screenSize": (1366, 768),
            "dpi": 96,
            "nightTransparency": 0.25,
            "termResolution": 0.05
        }

        # Load configuration
        c = json.load(open(configFile))
        self.cfg.update(c["config"])

        # Calculate plot size
        self.cfg["plotSize"] = (self.cfg["screenSize"][0]/self.cfg["dpi"], self.cfg["screenSize"][1]/self.cfg["dpi"])
        # Set min/max longitude
        self.cfg["maxLon"] = 180
        self.cfg["minLon"] = -180
        # Calculate min/max latitude
        self.cfg["maxLat"] = math.degrees(math.atan(math.sinh(self.cfg["screenSize"][1] * math.pi / self.cfg["screenSize"][0])))
        self.cfg["minLat"] = -self.cfg["maxLat"]
        # LatLon ranges
        self.cfg["latRange"] = self.cfg["maxLat"] - self.cfg["minLat"]
        self.cfg["lonRange"] = self.cfg["maxLon"] - self.cfg["minLon"]
        # Set filename variable
        self.filename = filename
        # Initialize map
        self.map = None
        # Initialize file save tracker
        self.saved = False
        
    
    def plotDaylight(self,  delta=None,  alpha=None,  *args,  **kwargs):
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Set default values if not specified
        if delta == None:
            delta = self.cfg["termResolution"]
        if alpha == None:
            alpha = self.cfg["nightTransparency"]
        # Add night shading to map
        self.map.nightshade(self.utcNowNaive, delta=delta, alpha=alpha, *args, **kwargs)
    
    
    def plotDaylightRadiation(self, *args, **kwargs):
        ''' Plot daylight radiation using Pysolar calculations on LatLon grid '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Get daylight grid
        daylight = self.irr.calcDaylightMesh(self.cfg["lonInterval"],
                                              self.cfg["lonRange"],
                                              self.cfg["latInterval"],
                                              self.cfg["latIntervalRange"],
                                              latOffset=self.cfg["latOffset"],
                                              fast=True, arrayGrid=True)
        # Normalize daylight
        daylight /= daylight.max()
        daylight[:, :, 3] = 1. - daylight[:, :, 3]
        daylight = np.ma.masked_less(daylight, 0.).filled(0.)
        daylight = np.ma.masked_greater(daylight, 0.65).filled(0.65)
        # Plot daylight on map
        self.map.imshow(daylight, *args, **kwargs)
    
    
    def plotTerminatorLine(self, *args, **kwargs):
        ''' Plot terminator line on map '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Calculate terminator location
        termPos = np.array(self.term.calcPath())
        # Prepare points for map
        X, Y = self.map(termPos[:,0], termPos[:,1])
        # Plot terminator on map
        self.map.scatter(X, Y, *args, **kwargs)
    
    
    def plotPoint(self, lon=None, lat=None, *args, **kwargs):
        ''' Plot point on map '''
        # Check for lon and lat
        if lon == None or lat == None:
            return False
        # Send to plotPoints
        self.plotPoints(lons=[lon], lats=[lat], *args, **kwargs)
    
    
    def plotPoints(self, lons=None, lats=None, *args, **kwargs):
        ''' Plot several points on map '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Check for lons and lats
        if lons == None or lats == None:
            return False
        # Plot specified point on map
        X, Y = self.map(lons, lats)
        self.map.scatter(X, Y, *args, **kwargs)
    
    
    def plotGreatCirclePath(self, *args, **kwargs):
        ''' Wrapper for great circle path drawing function in Basemap '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Plot specified point on map
        self.map.drawgreatcircle(*args, **kwargs)
    
    
    def addTextToMap(self, text=None, lon=None, lat=None, *args, **kwargs):
        ''' Add text to map at specified geographical location '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Check for lon and lat
        if lon == None or lat == None or text == None:
            return False
        # Plot specified point on map
        xPos, yPos = self.map(lon, lat)
        # Make point relative on image
        x = xPos/self.map.xmax
        y = yPos/self.map.ymax
        # Add text to figure
        self.figure.text(x, y, text, *args, **kwargs)


    def addTextToFig(self, text=None, x=None, y=None, *args, **kwargs):
        ''' Add text to map at specified figure relative location (0-1) '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Check for lon and lat
        if x == None or y == None or text == None:
            return False
        # Add text to figure
        self.figure.text(x, y, text, *args, **kwargs)
    
    
    def plotWorldtime(self, clockFile=None):
        ''' Plot world time at desired location specified in clock definition json file '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Time format
        fmt = '%H:%M'
        # Initialize counter
        n = 0
        # Load clock definition json file
        if not clockFile == None:
            try:
                with open(clockFile) as cf:
                    # Parse JSON
                    j = json.load(cf)
                    # Text parameters
                    txtparams = j['formatting']['text']
                    # Point parameters
                    ptparams = j['formatting']['point']
                    # Get colors
                    colors = j['formatting']['colors']
                    # Get display latitude & offset
                    dlat = j['formatting']['display']['lat']
                    doffset = j['formatting']['display']['offset']
                    # Sort locations
                    clocks = j['clocks']
                    sortCities = tuple(sorted(j['clocks'].iteritems(), key=lambda k: k[1]['lon']))
                    cityDirection = [False for i in range(len(sortCities))]
                    # Add each city to map
                    for i in range(len(sortCities)):
                        # Get city
                        city = sortCities[i][0]
                        # Convert to local time
                        localTime = self.utcNow.astimezone(pytz.timezone(clocks[city]['tz']))
                        # Get sun altitude at location
                        sunAlt = self.irr.calcSunAltitudeAtPoint(clocks[city]['lon'], clocks[city]['lat'], True)
                        # Color according to sun altitude
                        if sunAlt < -8.:
                            txtparams["color"] = '#09041c'
                        elif -8. <= sunAlt < -2.:
                            txtparams["color"] = '#221c32'
                        elif -2. <= sunAlt < 5.:
                            # txtparams["color"] = '#4071d7'
                            txtparams["color"] = '#1a0662'
                        else:
                            # txtparams["color"] = '#f28705'
                            txtparams["color"] = '#125700'
                        # Display longitude
                        dlon = clocks[city]['lon']
                        # Offset if within 5 degrees of previous
                        direction = False
                        if i > 0 and dlon-clocks[sortCities[i-1][0]]['lon'] < doffset:
                            direction = not cityDirection[i-1]
                        # Update direction
                        cityDirection[i] = direction
                        # Get x & y coordinates
                        xPos, yPos = self.map(dlon, dlat)
                        # Make point relative on image
                        x = xPos/self.map.xmax
                        y = yPos/self.map.ymax
                        # Direction true means up
                        if not direction:
                            # City name & clock lat below reference point
                            cityPos = y - (17 / self.cfg["screenSize"][1])
                            clockPos = y - (28 / self.cfg["screenSize"][1])
                        else:
                            # City name & clock lat above reference point
                            cityPos = y + (20 / self.cfg["screenSize"][1])
                            clockPos = y + (9 / self.cfg["screenSize"][1])
                        # Add city and reference point to map
                        self.plotPoint(lon=dlon, lat=clocks[city]['lat'], c=colors[n], **ptparams)
                        self.plotPoint(lon=dlon, lat=dlat, c=colors[n], **ptparams)
                        # Add location and time text to map above reference point
                        self.addTextToFig(x=x, y=cityPos, text='%s' % (city,), **txtparams)
                        self.addTextToFig(x=x, y=clockPos, text='%s' % (localTime.strftime(fmt),), **txtparams)
                        # Update counter
                        n += 1
                        # Cycle through colors
                        if n >= len(colors):
                            n = 0
            except:
                print "Error loading world time..."
    

    def plotTropicalWX(self, tropicalFile=None, txtX=0.052, txtY=0.0645, **kwargs):
        ''' Plot tropical weather data from json provided by Weather Underground API '''
        # If no map specified, raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        if self.figure == None:
            raise Exception('Map not yet generated!')
        # Tropical icon mapping
        tropicalImgBasePath = 'ico/wx/tropical/small/'
        tropicalFutureImgBasePath = 'ico/wx/tropical/xsmall/'
        tropicalImgPath = {
            -5: 'remnants.png',
            -4: 'invest.png',
            -3: 'extratropical.png',
            -2: 'depression.png',
            -1: 'depression.png',
            0: 'tropical-storm.png',
            1: 'hurricane-1.png',
            2: 'hurricane-2.png',
            3: 'hurricane-3.png',
            4: 'hurricane-4.png',
            5: 'hurricane-5.png'
        }
        tropicalImg = {}
        tropicalFutureImg = {}
        resizeFactor = {
                "current": 1.0,
                "future": 1.0
            }
        imgSize = None
        tropicalText = {
                "size"      : 8, 
                "ha"        : "center", 
                "va"        : "baseline", 
                "family"    : "sans serif",
                "stretch"   : "expanded", 
                "alpha"     : 0.7
            }
        updateTextFmt = {
                "color": '#a60000',
                "family": 'serif',
                "weight": 'semibold',
                "alpha": 0.9,
                "size": 10,
                "zorder": 10,
                "va": 'center'
            }
        updateTextFmt.update(kwargs)
        # Load clock definition json file
        if not tropicalFile == None:
            # try:
            with open(tropicalFile) as tf:
                # Parse JSON
                j = json.load(tf)
                # Process each storm
                if len(j['currenthurricane']) > 0:
                    for storm in j['currenthurricane']:
                        # Get storm details
                        name = storm['stormInfo']['stormName_Nice']
                        cat = storm['Current']['SaffirSimpsonCategory']
                        clat = storm['Current']['lat']
                        clon = storm['Current']['lon']
                        # Get pixel location of storm center
                        xPos,yPos = self.map(clon, clat)
                        # Make point relative on image
                        cx = xPos/self.map.xmax*self.cfg["screenSize"][0]
                        cy = yPos/self.map.ymax*self.cfg["screenSize"][1]
                        # Load icon and resize
                        if cat not in tropicalImg:
                            icon = Image.open(tropicalImgBasePath + tropicalImgPath[cat])
                            tropicalImg[cat] = np.asarray(icon.resize((int(icon.size[0]*resizeFactor["current"]), int(icon.size[1]*resizeFactor["current"]))))
                            if imgSize == None:
                                imgSize = (int(icon.size[0]*resizeFactor["current"]), int(icon.size[1]*resizeFactor["current"]))
                            icon = None
                        # Get icon size
                        icoW = imgSize[0]
                        icoH = imgSize[1]
                        # Plot storm
                        self.figure.figimage(tropicalImg[cat], xo=cx-icoW/2, yo=cy-icoH/2, zorder=9, alpha=0.8)
                        self.figure.text(cx/self.cfg["screenSize"][0], (cy-icoH-2.6)/self.cfg["screenSize"][1], name, **tropicalText)
                        # Set previous point
                        prevPt = {
                                    "cat": cat,
                                    "xo": cx-icoW/2,
                                    "yo": cy-icoH/2,
                                    "lat": clat,
                                    "lon": clon
                                }
                        # Forecast points
                        fcstPts = {}
                        # Process forecast
                        for fcst in storm['forecast']:
                            # Get storm details
                            if fcst['ForecastHour'].find('HR') > -1:
                                ftime = int(fcst['ForecastHour'][:-2])
                            elif fcst['ForecastHour'].find('DAY') > -1:
                                ftime = int(fcst['ForecastHour'][:-3])*24
                            else:
                                ftime = 0
                            cat = fcst['SaffirSimpsonCategory']
                            clat = fcst['lat']
                            clon = fcst['lon']
                            # Get pixel location of storm center
                            xPos,yPos = self.map(clon, clat)
                            # Make point relative on image
                            cx = xPos/self.map.xmax*self.cfg["screenSize"][0]
                            cy = yPos/self.map.ymax*self.cfg["screenSize"][1]
                            # Load icon and resize
                            if cat not in tropicalFutureImg:
                                icon = Image.open(tropicalFutureImgBasePath + tropicalImgPath[cat])
                                tropicalFutureImg[cat] = np.asarray(icon.resize((int(icon.size[0]*resizeFactor["future"]), int(icon.size[1]*resizeFactor["future"]))))
                                if imgSize == None:
                                    imgSize = (int(icon.size[0]*resizeFactor["future"]), int(icon.size[1]*resizeFactor["future"]))
                                icon = None
                            # Get icon size
                            icoW = imgSize[0]
                            icoH = imgSize[1]
                            # Add to forecast points
                            fcstPts[ftime] = {
                                    "cat": cat,
                                    "xo": cx-icoW/2,
                                    "yo": cy-icoH/2,
                                    "lat": clat,
                                    "lon": clon
                                }
                        # Plot forecasted track
                        for i in sorted(fcstPts):
                            # Plot storm
                            self.figure.figimage(tropicalFutureImg[fcstPts[i]['cat']],
                                                 xo=fcstPts[i]['xo'], yo=fcstPts[i]['yo'],
                                                 zorder=8, alpha=0.4)
                            # Update previous point
                            prevPt = fcstPts[i]
                # Plot update time
                updateText = 'Tropical Weather Updated:  {}'.format(time.strftime('%B %d, %Y  %I:%M%p', time.localtime(os.path.getmtime(tropicalFile))))
                self.addTextToFig(x=txtX, y=txtY, text=updateText, **updateTextFmt)
            # except:
            #     print "Error loading tropical weather data..."
    

    def plotDaylightUpdateTime(self, x=0.052, y=0.0645, *args, **kwargs):
        ''' Plot daylight update time '''
        pltArgs = {
                "color": '#a60000',
                "family": 'serif',
                "weight": 'semibold',
                "alpha": 0.95,
                "size": 10,
                "zorder": 10,
                "va": 'center'
            }
        pltArgs.update(kwargs)
        updateText = 'Daylight Updated:  {}'.format(self.localNow.strftime('%B %d, %Y  %I:%M%p'))
        self.addTextToFig(x=x, y=y, text=updateText, *args, **pltArgs)


    def saveMap(self):
        ''' Save map to file '''
        # Raise error if figure or filename do not exist
        if self.figure == None:
            raise Exception('Map not yet generated!')
        if self.filename == None:
            raise Exception('File name not specified')
        # Save figure
        self.figure.savefig(self.filename)
        # Update save tracker
        self.saved = True
        # Free up memory
        #plt.close() # TODO: determine why closing plot causes problems
    
    
    def createMap(self, projection='merc', resolution='l', *args, **kwargs):
        ''' Create figure and initialize map '''
        # Create figure
        self.figure = plt.figure(figsize=self.cfg["plotSize"], dpi=self.cfg["dpi"], linewidth=0.0)
        # Clear surrounding whitespace
        self.figure.add_axes([0, 0, 1, 1], frameon=False)
        # Create map object
        self.map = Basemap(projection=projection, llcrnrlat=self.cfg["minLat"], urcrnrlat=self.cfg["maxLat"],
                           llcrnrlon=self.cfg["minLon"], urcrnrlon=self.cfg["maxLon"], resolution=resolution,
                           *args, **kwargs)
    
    
    def loadImage(self, imageFile=None, replaceColor=None, *args, **kwargs):
        ''' Load an image and add to map '''
        # Raise error if figure,  map,  or filename do not exist
        if self.figure == None or self.map == None:
            raise Exception('Map not yet generated!')
        if imageFile == None:
            raise Exception('Image filename not specified.')

        # If transparency color not specified
        if replaceColor == None:
            # Read specified image and add to map
            self.map.imshow(plt.imread(imageFile), *args, **kwargs)
        else:
            # Open image and get data
            img = Image.open(imageFile)
            img = img.convert("RGBA")
            idata = img.getdata()
            # Update colors
            newColor = (replaceColor[0], replaceColor[1], replaceColor[2], 0)
            newData = [newColor if original[0] == replaceColor[0] and original[1] == replaceColor[1] and original[2] == replaceColorTransparent[2] else original for original in idata]
            # Update image data
            img.putdata(newData)
            # Add image to map
            self.map.imshow(img, *args, **kwargs)

    
    def setBlueMarbleBackground(self):
        ''' Load Blue Marble background image '''
        # Raise error if figure or filename do not exist
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Make background blue marble
        self.map.bluemarble(scale=self.cfg["blueMarbleMapScale"])
    
    
    def setWallpaper(self, cleanup=True):
        ''' Save map as bmp format and set it as the wallpaper '''
        # Check if file has been saved and execute accordingly
        if self.saved:
            # Open image
            bmpImage = Image.open(self.filename)
            # Determine crop coordinates
            w, h = bmpImage.size
            wDiff = int((w-self.cfg["screenSize"][0])*0.5)
            hDiff = int((h-self.cfg["screenSize"][1])*0.5)
            cropBox = (wDiff, hDiff, self.cfg["screenSize"][0]+wDiff, self.cfg["screenSize"][1]+hDiff)
            # Crop image
            bmpImage = bmpImage.crop(cropBox)
            # Build temporary file path
            newPath = os.getcwd()
            newPath = os.path.join(newPath,  'pywallpaper.bmp')
            # Try to save temporary file,  otherwise convert to correct image format and save
            try:
                bmpImage.save(newPath,  "BMP")
            except:
                bmpImage = bmpImage.convert("RGB")
                bmpImage.save(newPath,  "BMP")
            # Set wallpaper
            SPI_SETDESKWALLPAPER = 20
            ctypes.windll.user32.SystemParametersInfoA(SPI_SETDESKWALLPAPER,  0,  newPath ,  0)
            # Clean-up temporary file
            os.remove(newPath)
            # Clean-up original file if requested
            if cleanup:
                os.remove(self.filename)
        else:
            # Raise error if figure or filename do not exist
            if self.figure == None:
                raise Exception('Map not yet generated!')
            # Convert image
            plt.draw()
            bmpImage = Image.fromstring("RGB", self.figure.canvas.get_width_height(), self.figure.canvas.tostring_rgb())
            #plt.close()
            # Determine crop coordinates
            w, h = bmpImage.size
            wDiff = int((w-self.cfg["screenSize"][0])*0.5)
            hDiff = int((h-self.cfg["screenSize"][1])*0.5)
            cropBox = (wDiff, hDiff, self.cfg["screenSize"][0]+wDiff, self.cfg["screenSize"][1]+hDiff)
            # Crop image
            bmpImage = bmpImage.crop(cropBox)
            # Build temporary file path
            newPath = os.getcwd()
            newPath = os.path.join(newPath, 'pywallpaper.bmp')
            # Try to save temporary file,  otherwise convert to correct image format and save again
            try:
                bmpImage.save(newPath, "BMP")
            except:
                bmpImage = bmpImage.convert("RGB")
                bmpImage.save(newPath, "BMP")
            # Set wallpaper
            SPI_SETDESKWALLPAPER = 20
            ctypes.windll.user32.SystemParametersInfoA(SPI_SETDESKWALLPAPER, 0, newPath, 0)
            # Clean-up temporary file
            os.remove(newPath)


