#!/bin/sh
# DreamHost FastCGI dispatcher for geis.weldedanvil.com.
# Apache (mod_fcgid) execs this file; `exec` replaces it with the geisd binary,
# which inherits the FastCGI socket on fd 0 and serves via net/http/fcgi.
# The binary, config (with API keys), and data dir all live OUTSIDE the web root.
exec /home/dh_r7wsei/geis/geisd -fcgi -config /home/dh_r7wsei/geis/geisd-config.json
