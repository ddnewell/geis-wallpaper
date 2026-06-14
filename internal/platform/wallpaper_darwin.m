#import <AppKit/AppKit.h>
#import "wallpaper_darwin.h"

static NSDictionary *wallpaperOptions(void) {
    return @{
        NSWorkspaceDesktopImageScalingKey: @(NSImageScaleProportionallyUpOrDown),
        NSWorkspaceDesktopImageAllowClippingKey: @YES
    };
}

int geis_set_wallpaper_all(const char *path) {
    @autoreleasepool {
        NSString *p = [NSString stringWithUTF8String:path];
        if (p == nil) {
            return -1;
        }
        NSURL *url = [NSURL fileURLWithPath:p];
        NSWorkspace *ws = [NSWorkspace sharedWorkspace];
        NSArray<NSScreen *> *screens = [NSScreen screens];
        if (screens.count == 0) {
            return -1;
        }
        int failures = 0;
        for (NSScreen *screen in screens) {
            @autoreleasepool {
                NSError *err = nil;
                if (![ws setDesktopImageURL:url forScreen:screen options:wallpaperOptions() error:&err]) {
                    failures++;
                }
            }
        }
        return failures;
    }
}

int geis_set_wallpaper_for_screen(int idx, const char *path) {
    @autoreleasepool {
        NSArray<NSScreen *> *screens = [NSScreen screens];
        if (idx < 0 || idx >= (int)screens.count) {
            return -1;
        }
        NSString *p = [NSString stringWithUTF8String:path];
        if (p == nil) {
            return -2;
        }
        NSURL *url = [NSURL fileURLWithPath:p];
        NSError *err = nil;
        BOOL ok = [[NSWorkspace sharedWorkspace] setDesktopImageURL:url
                                                          forScreen:screens[idx]
                                                            options:wallpaperOptions()
                                                              error:&err];
        return ok ? 0 : -2;
    }
}

int geis_screen_count(void) {
    @autoreleasepool {
        return (int)[[NSScreen screens] count];
    }
}

int geis_screen_size(int idx, int *w, int *h) {
    @autoreleasepool {
        NSScreen *s = nil;
        if (idx < 0) {
            s = [NSScreen mainScreen];
        } else {
            NSArray<NSScreen *> *screens = [NSScreen screens];
            if (idx >= (int)screens.count) {
                return -1;
            }
            s = screens[idx];
        }
        if (s == nil) {
            return -1;
        }
        NSRect fr = [s frame];
        CGFloat scale = [s backingScaleFactor];
        if (scale <= 0) {
            scale = 1.0;
        }
        *w = (int)(fr.size.width * scale);
        *h = (int)(fr.size.height * scale);
        return 0;
    }
}
