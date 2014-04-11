#!/usr/bin/python
"""
@author: David Newell
@license: MIT

Global Event Information System
  Calculate Earth terminator position
Copyright 2013 Newell Designs, David Newell.
"""

# Import correct division
from __future__ import division
# Import required modules
from itertools import product
import math, datetime, pytz
import Pysolar as solar
import numpy as np
from dateutil.relativedelta import relativedelta


class daylight:
    """ Daylight object for calculating daylight and terminator

    :param now: Current time
    :type now: datetime
    """
    def __init__(self, now=None):
        """ Create daylight object

        :param now: Current time
        :type now: datetime
        """
        self.set_time(now)

    def set_time(self, now=None):
        """ Set time for calculations

        :param now: Current time
        :type now: datetime
        """
        # If when not specified or of wrong type, set to now
        if now == None or type(now) != datetime.datetime:
            self.utcNow = datetime.datetime.utcnow()
        else:
            self.utcNow = now
        # Update constants
        self._update_constants()

    def _set_prev_dec(self):
        """ Find the most recent December 31 """
        if self.utcNow.month == 12 and self.utcNow.day == 31:
            year = self.utcNow.year
        else:
            year = self.utcNow.year-1
        self._prev_dec = datetime.datetime(self.utcNow.year, 12, 31, tzinfo=pytz.utc)
        self._date_diff = self.utcNow-self._prev_dec

    def _set_prev_midnight(self):
        """ Find the most recent midnight """
        self._midnight = datetime.datetime(self.utcNow.year, self.utcNow.month, self.utcNow.day, tzinfo=pytz.utc)
        self._time_diff = self.utcNow-self._midnight

    def _update_constants(self):
        """ Update terminator constants """
        # Set previous December 31
        self._set_prev_dec()
        # Set most recent midnight
        self._set_prev_midnight()

        self._days = self._date_diff.days
        self._hours = self._time_diff.seconds/60.0/60.0

        self._mu = -3.6 + 0.9856 * self._days
        self._phi = self._mu + 1.9 * math.sin(math.radians(self._mu))
        self._omicron = self._phi + 102.9
        self._delta = 22.8 * math.sin(math.radians(self._omicron)) + 0.6 * math.pow(math.sin(math.radians(self._omicron)), 3.0)

        self._sigma = -15.0 * self._hours

    def _cartesian(self, arrays, out=None):
        """
        Generate a cartesian product of input arrays.

        Parameters
        ----------
        arrays : list of array-like
            1-D arrays to form the cartesian product of.
        out : ndarray
            Array to place the cartesian product in.

        Returns
        -------
        out : ndarray
            2-D array of shape (M, len(arrays)) containing cartesian products
            formed of input arrays.

        Examples
        --------
        >>> cartesian(([1, 2, 3], [4, 5], [6, 7]))
        array([[1, 4, 6],
               [1, 4, 7],
               [1, 5, 6],
               [1, 5, 7],
               [2, 4, 6],
               [2, 4, 7],
               [2, 5, 6],
               [2, 5, 7],
               [3, 4, 6],
               [3, 4, 7],
               [3, 5, 6],
               [3, 5, 7]])

        """
        arrays = [np.asarray(x) if type(x) is not np.ndarray else x for x in arrays]
        dtype = arrays[0].dtype

        n = np.prod([x.size for x in arrays])
        if out is None:
            out = np.zeros([n, len(arrays)], dtype=dtype)

        m = n / arrays[0].size
        out[:,0] = np.repeat(arrays[0], m)
        if arrays[1:]:
            self._cartesian(arrays[1:], out=out[0:m,1:])
            for j in xrange(1, arrays[0].size):
                out[j*m:(j+1)*m,1:] = out[0:m,1:]
        return out

    def sun_alt_at_point(self, lon=0, lat=0, fast=True):
        ''' Calculate irradiation at specified point '''
        # Calculate sun altitude depending on method requested
        if fast:
            sunAltitude = solar.GetAltitudeFast(lat, lon, self.utcNow)
        else:
            sunAltitude = solar.GetAltitude(lat, lon, self.utcNow)
        # Return altitude at specified point
        return sunAltitude

    def daylight_at_point(self, lon=0, lat=0, fast=True):
        """ Calculate irradiation at specific point

        :param lon: Longitude of point
        :type lon: float
        :param lat: Latitude of point
        :type lat: float
        :param fast: Use fast method
        :type fast: boolean
        """
        # Calculate sun altitude depending on method requested
        sunAltitude = self.sun_alt_at_point(lon, lat, fast)
        # Calculate direct irradiation
        irradiation = solar.radiation.GetRadiationDirect(self.utcNow, sunAltitude)
        # Return irradiation at specified point
        return irradiation

    def daylight_mesh(self, resolution=(360, 180), extent=[-180, 180, -90, 90], fast=True):
        """ Calculate irradiation mesh. Returns a numpy array with shape (resolution[0], resolution[1], 4).

        :param resolution: Number of points in mesh - (lon, lat) or # to be used for each
        :type resolution: tuple or int
        :param extent: Map extent (min lon, max lon, min lat, max lat)
        :type extent: list
        :param fast: Use fast method
        :type fast: boolean
        """
        # Capture resolution as a tuple
        if type(resolution) is int:
            resolution = (resolution, resolution)
        # Generate points for daylight mesh grid
        lats = np.linspace(extent[2], extent[3], num=resolution[1])
        lons = np.linspace(extent[0], extent[1], num=resolution[0])
        # Create numpy array of shape (lat x lon x 4)
        irradiation = np.zeros((resolution[1], resolution[0], 4))
        # Loop through lats and lons
        for i, j in product(xrange(resolution[0]), xrange(resolution[1])):
            irradiation[j][i][3] = self.daylight_at_point(lons[i], lats[j], fast)
        # for i in xrange(resolution[0]):
        #     for j in xrange(resolution[1]):
        #         # Get irradiation at current lat/lon point (using specified method)
        #         irradiation[j][i][3] = self.daylight_at_point(lons[i], lats[j], fast)
        # Return lists of irradiation points
        return irradiation

    def _term_point(self, degrees=0):
        """ Calculate points along terminator """
        loc = math.radians(degrees)

        cos_loc = math.cos(loc)
        delta_sin = math.sin(math.radians(self._delta))*math.sin(loc)
        sigma_rad = math.radians(self._sigma)

        x = -math.cos(sigma_rad)*delta_sin-math.sin(sigma_rad)*cos_loc
        y = -math.sin(sigma_rad)*delta_sin+math.cos(sigma_rad)*cos_loc

        # Calculate latitude and longitude
        lat = math.degrees(math.asin(math.cos(math.radians(self._delta))*math.sin(loc)))
        lon = math.degrees(math.atan2(y, x))

        # Return point on terminator
        return (lon, lat)

    def terminator_position(self, resolution=360):
        """ Calculate terminator positon

        :param resolution: Number of points to plot along terminator
        :type resolution: int
        """
        return [self._term_point(i) for i in np.linspace(0, 360, num=int(resolution))]

    def new_terminator_position(self, date, resolution=100):
        """ Calculate terminator positon at given time

        :param date: Current date
        :type date: datetime
        :param resolution: Number of points to calculate
        :type resolution: int
        """
        jd = _to_julian(date)
        t = _julian_cent(jd)
        ut = (jd-(int(jd-0.5)+0.5))*1440
        zen = np.zeros((resolution, resolution))
        lats = np.linspace(-90, 90, num=resolution)
        lons = np.linspace(-180, 180, num=resolution)
        term = []
        for ilat in xrange(1, resolution+1):
            for ilon in xrange(resolution):
                az, el = _sun_azimuth_elevation(t, ut, lats[-ilat], lons[ilon], 0)
                zen[-ilat, ilon] = el
            a = 90-zen[-ilat, :]
            mins = np.r_[False, a[1:]*a[:-1] <= 0] | np.r_[a[1:]*a[:-1] <= 0, False]
            zmin = mins & np.r_[False, a[1:] < a[:-1]]
            if True in zmin:
                ll = np.interp(0, a[zmin][-1::-1], lons[zmin][-1::-1])
                term.append([lats[-ilat], ll])
            zmin = mins & np.r_[a[1:] < a[:-1], False]
            if True in zmin:
                ll = np.interp(0, a[zmin], lons[zmin])
                term.insert([lats[-ilat], ll])
        return np.array(term)


        def _to_julian(self, date):
            """ Calculate julian date for give date

            :param date: Current date
            :type date: datetime
            """
            if date.month < 2:
                date.replace(year=date.year-1)
                date += relativedelta(month=12)
            A = np.floor(date.year/100)
            B = 2-A+np.floor(A/4)
            jd = np.floor(365.25*(date.year+4716))+np.floor(30.6001*(date.month+1))+date.day+B-1524.5
            jd += date.hour/24 + date.minute/1440 + date.second/86400
            return jd

        def _julian_cent(self, jdate):
            """ Covert Julian date to centuries since J2000

            :param jdate: Julian date
            :type jdate: float
            """
            return (jd-2451545)/36525

        def _sun_azimuth_elevation(self, t, localtime, lat, lon, zone):
            """ Calculate sun azimuth and zenith angle

            :param t:
            :param localtime:
            :param lat:
            :param lon:
            :param zone:
            """
            eqTime = calcEquationOfTime(t)
            theta = calcSunDeclination(t)
            solarTimeFix = eqTime+4*lon-60*zone
            earthRadVec = calcSunRadVector(t)
