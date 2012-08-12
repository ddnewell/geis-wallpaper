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
import os,  datetime

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

# --------------------------------------------------------

# Import configuration module
import GEISconfig as cfg


class plot:
    """
    Create map with options of daylight and other points and set as wallpaper (Windows only)
    """


    def __init__(self, now=None, filename=None):
        '''
        Constructor - Initialize variables
        '''
        # Get current time,  if not specified,  set to UTC now
        if now == None or type(now) != datetime.datetime:
            self.utcNow = datetime.datetime.utcnow()
        else:
            self.utcNow = now
        
        self.utcNow = self.utcNow.replace(tzinfo=pytz.utc)
        self.localNow = datetime.datetime.now()
        
        # Terminator objects
        self.irr = terminator.irradiation(now=self.utcNow)
        self.term = terminator.terminator(now=self.utcNow)
        
        # Blue marble scale
        self.mapScale = cfg.mapScale
        # Latitude ranges
        self.maxLat = cfg.maxLat
        self.minLat = cfg.minLat
        self.latRange = cfg.latRange
        self.latIntervalRange = cfg.latIntervalRange
        self.latInterval = cfg.latInterval
        # Longitude ranges
        self.maxLon = cfg.maxLon
        self.minLon = cfg.minLon
        self.lonRange = cfg.lonRange
        self.lonInterval = cfg.lonRange
        # Screen size in pixels
        self.screenSize = cfg.screenSize
        # Plot size and resolution
        self.dpi = cfg.dpi
        self.plotSize = cfg.plotSize
        # Width and height adjustments
        self.widthAdj = cfg.widthAdj
        self.heightAdj = cfg.heightAdj
        # Terminator properties
        self.nightTransparency = cfg.nightTransparency
        self.termResolution = cfg.termResolution
        self.transparencyRange = cfg.transparencyRange
        # Set filename variable
        self.filename = filename
        # Initialize map
        self.map = None
        # Initialize file save tracker
        self.saved = False
        
    
    def plotDaylight(self,  delta=None,  alpha=None,  *args,  **kwargs):
        # If no map specified,  raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Set default values if not specified
        if delta == None:
            delta = self.termResolution
        if alpha == None:
            alpha = self.nightTransparency
        # Add night shading to map
        self.map.nightshade(self.utcNow, delta=delta, alpha=alpha, *args, **kwargs)
    
    
    def plotDaylightRadiation(self, *args, **kwargs):
        # If no map specified,  raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Get daylight grid
        self.dayL = self.irr.calcDaylightMesh(self.lonInterval, self.lonRange, self.latInterval, self.latIntervalRange, latOffset=-0.6, fast=True, arrayGrid=True)
        # Normalize daylight
        self.dayL /= self.dayL.max()
        self.dayL[:, :, 3] = 1. - self.dayL[:, :, 3]
        self.dayL = np.ma.masked_less(self.dayL, 0.).filled(0.)
        self.dayL = np.ma.masked_greater(self.dayL, 0.65).filled(0.65)
        #self.dayL = np.ma.masked_greater(self.dayL, 0.9).filled(0.9)
        # Plot daylight on map
        self.map.imshow(self.dayL, *args, **kwargs)
    
    
    def plotTerminatorLine(self, *args, **kwargs):
        # If no map specified,  raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Calculate terminator location
        termPos = np.array(self.term.calcPath())
        # Prepare points for map
        X, Y = self.map(termPos[:, 0], termPos[:, 1])
        # Plot terminator on map
        self.map.scatter(X, Y, *args, **kwargs)
    
    
    def plotPoint(self, lon=None, lat=None, *args, **kwargs):
        # If no map specified,  raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Check for lon and lat
        if lon == None or lat == None:
            return False
        # Plot specified point on map
        x, y = self.map(lon, lat)
        self.map.scatter(x, y, *args, **kwargs)
    
    
    def plotPoints(self, lons=None, lats=None, *args, **kwargs):
        # If no map specified,  raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Check for lons and lats
        if lons == None or lats == None:
            return False
        # Plot specified point on map
        x, y = map(lons, lats)
        self.map.scatter(x, y, *args, **kwargs)
    
    
    def plotGreatCirclePath(self, *args, **kwargs):
        # If no map specified,  raise error
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Plot specified point on map
        self.map.drawgreatcircle(*args, **kwargs)
    
    
    def addText(self, lon=None, lat=None, text=None, *args, **kwargs):
        # If no map specified,  raise error
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
    
    
    def plotWorldtime(self):
        ''' Plot world time at bottom of screen '''
        # Time format
        fmt = '%H:%M'
        # Display latitude
        dispLat = -63.
        dispLatAlt = -59.
        # Text parameters
        txtparams = {
                'size'      : 8, 
                'ha'        : 'center', 
                'va'        : 'baseline', 
                'family'    : 'serif', 
                'weight'    : 'semibold', 
                'stretch'   : 'expanded', 
                'alpha'     : 1.
            }
        # Point parameters
        ptparams = {
                's'         : 18, 
                'marker'    : 'o', 
                'lw'        : 0.25, 
                'zorder'    : 10
            }
        # Locations
        clockCities = {
                #'City Name' : ( lon ,  lat ,  time zone ,  display latitude ,  point color )
                'SFO' : ( -122.2 ,  37.42 ,  'US/Pacific',  'std',  'r' ), 
                'HNL' : ( -157.85826988561186 ,  21.307001044273267 ,  'US/Hawaii',  'std',  'c' ), 
                'DEN' : ( -106.5 ,  39.74 ,  'US/Mountain' ,  'std',  'g' ), 
                'HOU' : ( -95.278888 ,  29.645419 ,  'US/Central',  'std',  'b' ), 
                'NYC' : ( -74.00592573426336 ,  40.7144039602474 ,  'US/Eastern',  'std',  'y' ), 
                'GRU' : ( -46.638753972131965 ,  -23.548921483335253 ,  'America/Sao_Paulo',  'std',  'b' ), 
                'LON' : ( -0.12 ,  51.5 ,  'Europe/London' ,  'std',  'm' ), 
                'MUC' : ( 11.6 ,  48.14 ,  'Europe/Berlin',  'std',  'r' ), 
                'MOW' : ( 37.6176647445313 ,  55.75585020857309 ,  'Europe/Moscow',  'std',  'w' ), 
                'DXB' : ( 55.31170142207517 ,  25.264506396060682 ,  'Asia/Dubai' ,  'std',  'b' ), 
                'DEL' : ( 77.22498415127225 ,  28.635358763353015 ,  'Asia/Kolkata',  'std',  'y' ), 
                'PEK' : ( 116.4074417328215 ,  39.90426446373647 ,  'Asia/Shanghai',  'std',  'c' ), 
                'SYD' : ( 151.21116441699496 ,  -33.85991741399278 ,  'Australia/Sydney',  'std',  'm' ), 
                'TYO' : ( 139.69172891760888 ,  35.6895526991927 ,  'Asia/Tokyo' ,  'std',  'g' ), 
                'AKL' : ( 172.7 ,  -36.848415576882836 ,  'Pacific/Auckland',  'std',  'w' )
            }
        # Add each city to map
        for city in clockCities:
            # Convert to local time
            localTime = self.utcNow.astimezone(pytz.timezone(clockCities[city][2]))
            # Get sun altitude at location
            sunAlt = self.irr.calcSunAltitudeAtPoint(clockCities[city][0], clockCities[city][1], True)
            # Color according to sun altitude
            if sunAlt < -8.:
                txtparams["color"] = '#09041c'
            elif -8. <= sunAlt < -2.:
                txtparams["color"] = '#221c32'
            elif -2. <= sunAlt < 5.:
                txtparams["color"] = '#4071d7'
            else:
                txtparams["color"] = '#f28705'
            # Choose display latitude
            if clockCities[city][3] == 'alt':
                dLat = dispLatAlt
            elif clockCities[city][3] == 'std':
                dLat = dispLat
            # Add point and reference point to map
            self.plotPoint(lon=clockCities[city][0], lat=clockCities[city][1], c=clockCities[city][4], **ptparams)
            self.plotPoint(lon=clockCities[city][0], lat=dLat+1.75, c=clockCities[city][4], **ptparams)
            # Add location and time text to map
            self.addText(lon=clockCities[city][0], lat=dLat, text='%s' % (city, ), **txtparams)
            self.addText(lon=clockCities[city][0], lat=dLat-1.25, text='%s' % (localTime.strftime(fmt), ), **txtparams)
            
    
    def saveMap(self):
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
        #plt.close()
    
    
    def createMap(self, projection='merc', resolution='l', *args, **kwargs):
        # Create figure
        self.figure = plt.figure(figsize=self.plotSize, dpi=self.dpi, linewidth=0.0)#, frameon=False)
        self.figure.add_axes([0, 0, 1, 1], frameon=False)
        # Create map object
        self.map = Basemap(projection=projection, llcrnrlat=self.minLat, urcrnrlat=self.maxLat, \
                           llcrnrlon=self.minLon, urcrnrlon=self.maxLon, resolution=resolution, \
                           *args, **kwargs)
    
    
    def loadImage(self, imageLoc=None, *args, **kwargs):
        # Raise error if figure,  map,  or filename do not exist
        if self.figure == None or self.map == None:
            raise Exception('Map not yet generated!')
        if imageLoc == None:
            raise Exception('File name not specified')
        # Read specified image and plot to 
        self.map.imshow(plt.imread(imageLoc), *args, **kwargs)
    
    
    def setBlueMarbleBackground(self):
        # Raise error if figure or filename do not exist
        if self.map == None:
            raise Exception('Map not yet generated!')
        # Make background blue marble
        self.map.bluemarble(scale=self.mapScale)
    
    
    def setWallpaper(self, cleanup=True):
        ''' Save map as bmp format and set it as the wallpaper '''
        # Check if file has been saved and execute accordingly
        if self.saved:
            # Open image
            bmpImage = Image.open(self.filename)
            # Determine crop coordinates
            w, h = bmpImage.size
            wDiff = int((w-self.screenSize[0])*self.widthAdj)
            hDiff = int((h-self.screenSize[1])*self.heightAdj)
            cropBox = ( wDiff, hDiff, self.screenSize[0]+wDiff, self.screenSize[1]+hDiff )
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
            wDiff = int((w-self.screenSize[0])*self.widthAdj)
            hDiff = int((h-self.screenSize[1])*self.heightAdj)
            cropBox = ( wDiff, hDiff, self.screenSize[0]+wDiff, self.screenSize[1]+hDiff )
            # Crop image
            bmpImage = bmpImage.crop(cropBox)
            # Build temporary file path
            newPath = os.getcwd()
            newPath = os.path.join(newPath,  'pywallpaper.bmp')
            # Try to save temporary file,  otherwise convert to correct image format and save again
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


