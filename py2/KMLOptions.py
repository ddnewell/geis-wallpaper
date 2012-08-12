'''
Created on Jan 27, 2011

@author: David Newell
'''

# Options dictionaries for use in KML applications

# Dictionary of altitude modes and markup
altitudeMode = {
        'relativeToGround' : '<altitudeMode>relativeToGround</altitudeMode>\n',
        'relativeToSeaFloor' : '<altitudeMode>relativeToGround</altitudeMode>\n<gx:altitudeMode>relativeToSeaFloor</gx:altitudeMode>\n',
        'clampToGround' : '<tessellate>1</tessellate>\n<altitudeMode>clampToGround</altitudeMode>\n',
        'clampToSeaFloor' : '<tessellate>1</tessellate>\n<altitudeMode>clampToGround</altitudeMode>\n<gx:altitudeMode>clampToSeaFloor</gx:altitudeMode>\n',
        'absolute' : '<altitudeMode>absolute</altitudeMode>\n'
    }

# Dictionary of extrude options and markup
extrude = {
        'yes' : '<extrude>1</extrude>\n',
        'no' : '<extrude>0</extrude>\n'
    }