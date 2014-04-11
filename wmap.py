#!/usr/bin/python
"""
@author: David Newell
@license: MIT

Global Event Information System
Copyright 2012 Newell Designs, David Newell.
"""

# --------------------------------------------------------
#  Import modules
# --------------------------------------------------------

from __future__ import division

import matplotlib
matplotlib.use('Agg')

import os, datetime, time, json, pytz, math, logging
import numpy as np
import matplotlib.pyplot as plt
from matplotlib import cm
from matplotlib.backends.backend_agg import FigureCanvasAgg
import shapely.geometry as sgeom
import cartopy.crs as ccrs
from PIL import Image
import daylight


# --------------------------------------------------------
#  Plotting object
# --------------------------------------------------------

class Plot(object):
    """Create map with options of daylight and other points to set as wallpaper

    :param now: Current date and time
    :type now: datetime
    :param save_file: Filename for saved file
    :type save_file: str
    :param config_file: Configuration filename
    :type config_file: str
   """
    def __init__(self, current_date=None, save_file=None, config_file=None):
        """ Create a map plot

        :param now: Current date and time
        :type now: datetime
        :param save_file: Filename for saved file
        :type save_file: str
        :param config_file: Configuration filename
        :type config_file: str
        """
        # If no configuration file specified, raise error
        if config_file == None:
            raise Exception('No configuration file specified!')

        # Get current time, if not specified,  set to UTC now
        if current_date == None or type(current_date) != datetime.datetime:
            self._utc_now = datetime.datetime.utcnow()
        else:
            self._utc_now = current_date

        # Setup date & time
        self._utc_now_naive = self._utc_now
        self._utc_now = self._utc_now.replace(tzinfo=pytz.utc)
        self._local_now = datetime.datetime.now()

        # Daylight object
        self._daylight = daylight.daylight(now=self._utc_now)

        # Load configuration
        cfg = {}
        try:
            c = json.load(open(config_file))
            cfg = c["config"]
        except:
            pass
        # Set configuration variables (or defaults)
        self._dpi = cfg['dpi'] if 'dpi' in cfg else 96
        self._screen_size = cfg['screen_size'] if 'screen_size' in cfg else (1366, 768)
        # Calculate plot size
        self._plot_size = (self._screen_size[0]/self._dpi, self._screen_size[0]/self._dpi/2)
        # Set min/max longitude
        self._max_lon = 180
        self._min_lon = -180
        # Set min/max latitude
        self._max_lat = 90
        self._min_lat = -90
        # Map extent
        self._extent = [self._min_lon, self._max_lon, self._min_lat, self._max_lat]
        # Set darkness parameter for daylight
        self._darkness = cfg['darkness'] if 'darkness' in cfg else 0.8
        # Latitude and longitude ranges
        self._lat_range = self._max_lat - self._min_lat
        self._lon_range = self._max_lon - self._min_lon
        # Set filename variable
        self.save_file = save_file
        # Initialize map
        self._map = None
        # Initialize file save tracker
        self.saved = False
        # Satellite and tropical shell scripts
        self._sat_script = cfg['sat_script'] if 'sat_script' in cfg else './get_satellite.mac.sh'
        self._tropical_script = cfg['tropical_script'] if 'tropical_script' in cfg else './get_tropical.mac.sh'

    def plot_daylight(self, *args, **kwargs):
        """Plot daylight radiation using Pysolar calculations on LatLon grid"""
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Get daylight grid
        radiation = self._daylight.daylight_mesh(resolution=(540, 270), extent=self._extent, fast=True)
        # Normalize daylight
        radiation /= radiation.max()
        radiation[:, :, 3] = 1 - radiation[:, :, 3]
        # radiation = np.ma.masked_less(radiation, 0).filled(0)
        radiation = np.ma.masked_greater(radiation, self._darkness).filled(self._darkness)
        # Plot daylight on map
        self._map.imshow(radiation, interpolation='bicubic', extent=self._extent, transform=ccrs.PlateCarree(), *args, **kwargs)

    def plot_terminator(self, *args, **kwargs):
        """Plot terminator line on map"""
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Get terminator points
        term_points = self._daylight.terminator_position(resolution=1000)
        # Plot points
        for pt in term_points:
            self._map.plot(pt[0], pt[1], transform=ccrs.PlateCarree(), *args, **kwargs)

    def plot_point(self, lon=None, lat=None, *args, **kwargs):
        """Plot point on map"""
        # Check for lon and lat
        if lon == None or lat == None:
            return False
        # Send to plot_points
        self.plot_points(lons=[lon], lats=[lat], *args, **kwargs)

    def plot_points(self, lons=None, lats=None, *args, **kwargs):
        """Plot several points on map"""
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Check for lons and lats
        if lons == None or lats == None:
            return False
        # Plot specified point on map
        self._map.scatter(lons, lats, transform=ccrs.PlateCarree(), *args, **kwargs)

    def plot_great_circle(self, start, end, *args, **kwargs):
        """Draw a great circle path on map

        :param start: Starting point (lon, lat)
        :type start: tuple
        :param end: Ending point (lon, lat)
        :type end: tuple
        """
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Plot great circle
        self._map.plot(start, end, transform=ccrs.Geodetic(), *args, **kwargs)

    def add_text_to_map(self, text=None, lon=None, lat=None, *args, **kwargs):
        """Add text to map at specified geographical location"""
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Check for lon and lat
        if lon == None or lat == None or text == None:
            return False
        # Add text to map
        self._map.text(lon, lat, text, transform=ccrs.PlateCarree(), *args, **kwargs)

    def add_text_to_fig(self, text=None, x=None, y=None, *args, **kwargs):
        """Add text to map at specified figure relative location (0-1)"""
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Check for lon and lat
        if x == None or y == None or text == None:
            return False
        # Add text to figure
        self._figure.text(x, y, text, *args, **kwargs)

    def plot_worldtime(self, clockFile=None):
        """Plot world time at desired location specified in clock definition json file"""
        # If no map specified, raise error
        if self._map == None:
            raise Exception('Map not yet generated!')
        # Time format
        fmt = '%H:%M'
        # Initialize counter
        n = 0
        # Load clock definition json file
        if not clockFile == None:
            # Plot background
            self._map.fill([-180, -180, 180, 180], [-90, -64.25, -64.25, -90], transform=ccrs.PlateCarree(), alpha=0.5, color='white', zorder=2)
            self._map.fill([-180, -180, 180, 180], [-90, -64.25, -64.25, -90], transform=ccrs.PlateCarree(), alpha=0.4, color='wheat', zorder=3)
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
                    localTime = self._utc_now.astimezone(pytz.timezone(clocks[city]['tz']))
                    # Get sun altitude at location
                    sunAlt = self._daylight.sun_alt_at_point(clocks[city]['lon'], clocks[city]['lat'], True)
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
                    # Direction true means adjust text up
                    if not direction:
                        # City name & clock lat below reference point
                        cityPos = dlat - 4
                        clockPos = dlat - 6.9
                    else:
                        # City name & clock lat above reference point
                        cityPos = dlat + 4.8
                        clockPos = dlat + 2
                    # Add city and reference point to map
                    self.plot_point(lon=dlon, lat=clocks[city]['lat'], c=colors[n], **ptparams)
                    self.plot_point(lon=dlon, lat=dlat, c=colors[n], **ptparams)
                    # Add location and time text to map above reference point
                    self.add_text_to_map(lon=dlon, lat=cityPos, text=city, **txtparams)
                    self.add_text_to_map(lon=dlon, lat=clockPos, text=localTime.strftime(fmt), **txtparams)
                    # Update counter
                    n += 1
                    # Cycle through colors
                    if n >= len(colors):
                        n = 0

    def plot_tropical_wx(self, tropicalFile=None, txtX=0.968, txtY=0.015, **kwargs):
        """Plot tropical weather data from json provided by Weather Underground API"""
        # If no map specified, raise error
        if self._map == None or self._figure == None:
            raise Exception('Map not yet generated!')
        # Tropical icon mapping
        tropicalImgBasePath = 'ico/wx/tropical/small/'
        tropicalFutureImgBasePath = 'ico/wx/tropical/small/'
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
                'current': 1.0,
                'future': 0.7
            }
        imgSize = None
        tropicalText = {
                'size'      : 13,
                'ha'        : 'center',
                'va'        : 'baseline',
                'family'    : 'sans serif',
                'stretch'   : 'expanded',
                'alpha'     : 0.7,
                'bbox': {
                    'facecolor': 'wheat',
                    'alpha': 0.5,
                    'boxstyle': 'round'
                }
            }
        updateTextFmt = {
                'color': '#a60000',
                'family': 'serif',
                'weight': 'semibold',
                'alpha': 0.9,
                'size': 16,
                'zorder': 10,
                'va': 'center',
                'ha': 'right'
            }
        updateTextFmt.update(kwargs)
        # Get last update time
        lastUpdate = os.path.getmtime(tropicalFile)
        # If more than half an hour old, try to update
        if lastUpdate < time.time() - 1800:
            os.system(self._tropical_script)
        # Update last update time
        lastUpdate = os.path.getmtime(tropicalFile)
        # Load tropical data
        if not tropicalFile == None:
            try:
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
                            # Get storm center point relative to image
                            cx = (clon+self._lon_range/2)/self._lon_range*self._screen_size[0]
                            cy = (clat+self._lat_range/2)/self._lat_range*self._screen_size[1]
                            # Load icon and resize
                            if cat not in tropicalImg:
                                icon = Image.open(tropicalImgBasePath + tropicalImgPath[cat])
                                tropicalImg[cat] = np.asarray(icon.resize((int(icon.size[0]*resizeFactor['current']), int(icon.size[1]*resizeFactor['current']))))
                                if imgSize == None:
                                    imgSize = (int(icon.size[0]*resizeFactor['current']), int(icon.size[1]*resizeFactor['current']))
                                icon = None
                            # Get icon size
                            icoW = imgSize[0]
                            icoH = imgSize[1]
                            # Plot storm
                            self._figure.figimage(tropicalImg[cat], xo=cx-icoW/2, yo=cy-icoH/2, zorder=9, alpha=0.8)
                            self._figure.text(cx/self._screen_size[0], (cy-icoH-2)/self._screen_size[1], name, **tropicalText)
                            # Set previous point
                            prevPt = {
                                        'cat': cat,
                                        'xo': clon,
                                        'yo': clat,
                                        'lat': clat,
                                        'lon': clon
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
                                # Get storm center point relative to image
                                cx = (clon+self._lon_range/2)/self._lon_range*self._screen_size[0]
                                cy = (clat+self._lat_range/2)/self._lat_range*self._screen_size[1]
                                # Load icon and resize
                                if cat not in tropicalFutureImg:
                                    icon = Image.open(tropicalFutureImgBasePath + tropicalImgPath[cat])
                                    tropicalFutureImg[cat] = np.asarray(icon.resize((int(icon.size[0]*resizeFactor['future']), int(icon.size[1]*resizeFactor['future']))))
                                    if imgSize == None:
                                        imgSize = (int(icon.size[0]*resizeFactor['future']), int(icon.size[1]*resizeFactor['future']))
                                    icon = None
                                # Get icon size
                                icoW = imgSize[0]
                                icoH = imgSize[1]
                                # Add to forecast points
                                fcstPts[ftime] = {
                                        'cat': cat,
                                        'xo': cx-icoW/2,
                                        'yo': cy-icoH/2,
                                        'lat': clat,
                                        'lon': clon
                                    }
                            # Plot forecasted track
                            for i in sorted(fcstPts):
                                # Plot storm
                                self._figure.figimage(tropicalFutureImg[fcstPts[i]['cat']],
                                                                 xo=fcstPts[i]['xo'], yo=fcstPts[i]['yo'],
                                                                 zorder=8, alpha=0.4, transform=ccrs.PlateCarree())
                                # Update previous point
                                prevPt = fcstPts[i]
                    # Plot update time
                    updateText = 'Tropical Weather Updated:  {}'.format(time.strftime('%B %d, %Y  %I:%M%p', time.localtime(lastUpdate)))
                    self.add_text_to_fig(x=txtX, y=txtY, text=updateText, **updateTextFmt)
            except:
                logging.warning('Error loading tropical weather data...')

    def plot_daylight_update_time(self, x=0.032, y=0.015, *args, **kwargs):
        """Plot daylight update time"""
        pltArgs = {
                "color": '#a60000',
                "family": 'serif',
                "weight": 'semibold',
                "alpha": 0.95,
                "size": 16,
                "zorder": 10,
                "va": 'center'
            }
        pltArgs.update(kwargs)
        updateText = 'Daylight Updated:  {}'.format(self._local_now.strftime('%B %d, %Y  %I:%M%p'))
        self.add_text_to_fig(x=x, y=y, text=updateText, *args, **pltArgs)

    def update_satellite(self, imageFile=None):
        """Update satellite image based on time since last update"""
        # Raise error if figure,  map,  or filename do not exist
        if self._figure == None or self._map == None:
            raise Exception('Map not yet generated!')
        if imageFile == None:
            raise Exception('Image filename not specified.')
        # Get last update time
        lastUpdate = os.path.getmtime(imageFile)
        # If more than an hour old, try to update
        if lastUpdate < time.time() - 3600:
            os.system(self._sat_script)
        # Update last update time
        lastUpdate = os.path.getmtime(imageFile)

    def save_map(self):
        """Save map to file"""
        # Raise error if figure or filename do not exist
        if self._figure == None:
            raise Exception('Map not yet generated!')
        if self.save_file == None:
            raise Exception('File name not specified')
        # Save figure
        self._figure.savefig(self.save_file, bbox_inches='tight')
        # Update save tracker
        self.saved = True
        # Free up memory
        # TODO: determine why closing plot causes problems
        #plt.close()

    def create_map(self):
        """Create figure and initialize map"""
        # Create figure
        self._figure = plt.figure(figsize=self._plot_size, linewidth=0.0, dpi=self._dpi)
        # Create map object and clear surrounding whitespace
        self._map = self._figure.add_axes([0, 0, 1, 1], frameon=False, projection=ccrs.PlateCarree())
        # Set global zoom level
        self._map.set_global()
        # Plot stock image
        self._map.background_patch.set_visible(False)
        self._map.outline_patch.set_visible(False)

    def load_image(self, imageFile=None, replaceColor=None, *args, **kwargs):
        """Load an image and add to map"""
        # Raise error if figure,  map,  or filename do not exist
        if self._figure == None or self._map == None:
            raise Exception('Map not yet generated!')
        if imageFile == None:
            raise Exception('Image filename not specified.')
        # If transparency color not specified
        if replaceColor == None:
            # Read specified image and add to map
            self._map.imshow(plt.imread(imageFile), transform=ccrs.PlateCarree(), *args, **kwargs)
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
            self._map.imshow(img, transform=ccrs.PlateCarree(), *args, **kwargs)

    def set_wallpaper(self):
        """Save map as bmp format and set it as the wallpaper"""
        # Check if file has been saved and execute accordingly
        if self.saved:
            # Open image
            crop_img = Image.open(self.save_file)
        else:
            # Raise error if figure or filename do not exist
            if self._figure == None:
                raise Exception('Map not yet generated!')
            # Convert image
            canvas = plt.get_current_fig_manager().canvas
            agg = canvas.switch_backends(FigureCanvasAgg)
            agg.draw()
            crop_img = Image.fromstring('RGB', self._figure.canvas.get_width_height(), agg.tostring_rgb())
            #plt.close()
        # Determine crop coordinates
        w, h = crop_img.size
        wDiff = int((w-self._screen_size[0])*0.5)
        hDiff = int((h-self._screen_size[1])*0.5)
        cropBox = (wDiff, hDiff, self._screen_size[0]+wDiff, self._screen_size[1]+hDiff)
        # Crop image
        crop_img = crop_img.crop(cropBox)
        # Try to save temporary file,  otherwise convert to correct image format and save again
        try:
            crop_img.save(self.save_file, 'PNG')
        except:
            crop_img = crop_img.convert('RGB')
            crop_img.save(self.save_file, 'PNG')

