#ifndef GEIS_WALLPAPER_DARWIN_H
#define GEIS_WALLPAPER_DARWIN_H

// Sets the desktop image for every screen's current Space. Returns the number
// of screens that failed (0 = all succeeded), or -1 if no screens were found.
int geis_set_wallpaper_all(const char *path);

// Sets the desktop image for a single screen (by index into [NSScreen screens]).
// Returns 0 on success, -1 on bad index, -2 on failure.
int geis_set_wallpaper_for_screen(int idx, const char *path);

// Returns the number of attached screens.
int geis_screen_count(void);

// Writes screen idx's backing-store (pixel) size into *w,*h. idx<0 means the
// main screen. Returns 0 on success, -1 if it could not be read.
int geis_screen_size(int idx, int *w, int *h);

#endif
