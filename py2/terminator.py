#!/usr/bin/python
'''
@author: David Newell
@license: MIT

Global Event Information System
  Calculate Earth terminator position
Copyright 2012 Newell Designs, David Newell.
'''

# Import useful modules

# ---------------------------------------------------------

# Import correct division
from __future__ import division

# Import math module
import math

# Import date and time modules
import datetime

# Import solar irradiation module
import solar

# Import numpy for array
import numpy as np

# Import plotting library
import matplotlib

# Import UTC time zone
import pytz

# ---------------------------------------------------------


class terminator:
    '''
    Calculate location of Terminator at specified date/time
    '''


    def __init__(self,now=None):
        '''
        Constructor
        '''
        self.setTime(now)
    
    
    def setTime(self,now=None):
        ''' Set the time for calculation '''
        # If when not specified or of wrong type, set to now
        if now == None or type(now) != datetime.datetime:
            self.utcNow = datetime.datetime.utcnow()
        else:
            self.utcNow = now
        # Set previous December 31
        self.setPrevDec()
        # Set most recent midnight
        self.setMidnight()
        # Update constants
        self.updateConstants()
    
    
    def setPrevDec(self):
        ''' Find the most recent December 31 '''
        if self.utcNow.month == 12 and self.utcNow.day == 31:
            year = self.utcNow.year
        else:
            year = self.utcNow.year - 1
        month = 12
        day = 31
        hour = 0
        minute = 0
        second = 0
        microsecond = 0
        self.prev = datetime.datetime(year,month,day,hour,minute,second,microsecond,tzinfo=pytz.utc)
        self.DateDiff = self.utcNow-self.prev
    
    
    def setMidnight(self):
        ''' Find the most recent midnight '''
        year = self.utcNow.year
        month = self.utcNow.month
        day = self.utcNow.day
        hour = 0
        minute = 0
        second = 0
        microsecond = 0
        self.midnight = datetime.datetime(year,month,day,hour,minute,second,microsecond,tzinfo=pytz.utc)
        self.TimeDiff = self.utcNow-self.midnight
    
    
    def updateConstants(self):
        ''' Update constants '''
        self.days = self.DateDiff.days
        self.hours = self.TimeDiff.seconds/60.0/60.0
        
        self.mu = -3.6 + 0.9856 * self.days
        self.phi = self.mu + 1.9 * math.sin( math.radians(self.mu) )
        self.omicron = self.phi + 102.9
        self.delta = 22.8 * math.sin( math.radians(self.omicron) ) + 0.6 * math.pow( math.sin( math.radians(self.omicron) ) , 3.0 )
        
        self.sigma = -15.0 * self.hours
    
    
    def calcPoint(self,degrees=0):
        ''' Calculate points along terminator '''
        loc = math.radians(degrees)
        
        x = -math.cos(math.radians(self.sigma))*math.sin(math.radians(self.delta))*math.sin(loc) \
                -math.sin(math.radians(self.sigma))*math.cos(loc)
        y = -math.sin(math.radians(self.sigma))*math.sin(math.radians(self.delta))*math.sin(loc) \
                +math.cos(math.radians(self.sigma))*math.cos(loc)
        
        # Calculate latitude and longitude
        B = math.degrees( math.asin( math.cos(math.radians(self.delta)) * math.sin(loc) ) )
        L = math.degrees( math.atan2( y , x ) )
        
        # Return (longitude,latitude)
        return (L,B)
    
    
    def calcPath(self,interval=0.5):
        ''' Calculate Terminator path '''
        terminator = []
        
        curPt = 0.0
        endPt = 360.0
        
        while curPt <= endPt:
            terminator.append( self.calcPoint(curPt) )
            curPt += interval
        
        # Return list of terminator points
        return terminator



class irradiation:
    '''
    Calculate irradiation mesh
    '''
    
    def __init__(self,now=None):
        '''
        Constructor
        '''
        self.setTime(now)
    
    
    def setTime(self,now=None):
        ''' Set the time for calculation '''
        # If when not specified or of wrong type, set to now
        if now == None or type(now) != datetime.datetime:
            self.utcNow = datetime.datetime.utcnow()
        else:
            self.utcNow = now
    
    
    def calcSunAltitudeAtPoint(self,lon=0,lat=0,fast=True):
        ''' Calculate irradiation at specified point '''
        # Calculate sun altitude depending on method requested
        if fast:
            sunAltitude = solar.GetAltitudeFast(lat, lon, self.utcNow)
        else:
            sunAltitude = solar.GetAltitude(lat, lon, self.utcNow)
        # Return altitude at specified point
        return sunAltitude
    
    
    def calcDaylightAtPoint(self,lon=0,lat=0,fast=True):
        ''' Calculate irradiation at specified point '''
        # Calculate sun altitude depending on method requested
        sunAltitude = self.calcSunAltitudeAtPoint(lon,lat,fast)
        # Calculate direct irradiation
        irradiation = solar.radiation.GetRadiationDirect(self.utcNow, sunAltitude)
        # Return irradiation at specified point
        return irradiation
    
    
    def calcDaylightMesh(self,lonInterval=10.,lonRange=360.,latInterval=10.,latRange=160.,lonOffset=0.,latOffset=0.,fast=True,arrayGrid=True):
        ''' Calculate Irradiation mesh '''
        # Generate mesh grid
        lats = [ math.degrees(math.atan(math.sinh( -latRange/2.0+x*latInterval )))\
                 for x in range(int(latOffset/latInterval),int((latRange+latOffset)/latInterval)+1)\
                 if -90. <= math.degrees(math.atan(math.sinh( -latRange/2.0+x*latInterval ))) <= 90. ]
        lons = [ -lonRange/2.0 + x*lonInterval\
                 for x in range(int(lonOffset/lonInterval),int((lonRange+lonOffset)/lonInterval)+1)\
                 if -180. <= -lonRange/2.0 + x*lonInterval <= 180. ]
        # Length of each axis
        x = len(lats)
        y = len(lons)
        # If array requested, return numpy array, otherwise just list of points
        if arrayGrid:
            # Create array of correct size
            irradiation = np.zeros((x, y, 4))
            # Loop through lats and lons
            for i in range(y):
                for j in range(x):
                    # Get irradiation at current lat/lon point (using specified method)
                    irradiation[j][i][3] = self.calcDaylightAtPoint(lons[i], lats[j], fast)
        else:
            # Get irradiation for each lat/lon point (using specified method)
            irradiation = [(lon, lat, self.calcDaylightAtPoint(lon, lat, fast)) for lat in lats for lon in lons]
        # Return lists of irradiation points
        return irradiation


