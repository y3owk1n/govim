#import "appwatcher.h"
#import <Cocoa/Cocoa.h>

extern void handleAppLaunch(const char* appName, const char* bundleID);
extern void handleAppTerminate(const char* appName, const char* bundleID);
extern void handleAppActivate(const char* appName, const char* bundleID);
extern void handleAppDeactivate(const char* appName, const char* bundleID);
extern void handleScreenParametersChanged(void);

@interface AppWatcherDelegate : NSObject
@end

@implementation AppWatcherDelegate

- (void)applicationDidLaunch:(NSNotification *)notification {
    @autoreleasepool {
        NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
        if (!app) return;

        // Copy strings to prevent dangling pointers
        NSString *appName = app.localizedName ?: @"Unknown";
        NSString *bundleID = app.bundleIdentifier ?: @"unknown.bundle";

        char *appNameCopy = strdup([appName UTF8String]);
        char *bundleIDCopy = strdup([bundleID UTF8String]);

        // Call Go callback on main thread for consistency
        dispatch_async(dispatch_get_main_queue(), ^{
            if (appNameCopy && bundleIDCopy) {
                handleAppLaunch(appNameCopy, bundleIDCopy);
            }
            free(appNameCopy);
            free(bundleIDCopy);
        });
    }
}

- (void)applicationDidTerminate:(NSNotification *)notification {
    @autoreleasepool {
        NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
        if (!app) return;

        NSString *appName = app.localizedName ?: @"Unknown";
        NSString *bundleID = app.bundleIdentifier ?: @"unknown.bundle";

        char *appNameCopy = strdup([appName UTF8String]);
        char *bundleIDCopy = strdup([bundleID UTF8String]);

        dispatch_async(dispatch_get_main_queue(), ^{
            if (appNameCopy && bundleIDCopy) {
                handleAppTerminate(appNameCopy, bundleIDCopy);
            }
            free(appNameCopy);
            free(bundleIDCopy);
        });
    }
}

- (void)applicationDidActivate:(NSNotification *)notification {
    @autoreleasepool {
        NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
        if (!app) return;

        NSString *appName = app.localizedName ?: @"Unknown";
        NSString *bundleID = app.bundleIdentifier ?: @"unknown.bundle";

        // Allocate copies that Go can safely use
        char *appNameCopy = strdup([appName UTF8String]);
        char *bundleIDCopy = strdup([bundleID UTF8String]);

        if ([NSThread isMainThread]) {
            // Already on main thread, call directly
            if (appNameCopy && bundleIDCopy) {
                handleAppActivate(appNameCopy, bundleIDCopy);
            }
            free(appNameCopy);
            free(bundleIDCopy);
        } else {
            // Not on main thread, dispatch
            dispatch_async(dispatch_get_main_queue(), ^{
                if (appNameCopy && bundleIDCopy) {
                    handleAppActivate(appNameCopy, bundleIDCopy);
                }
                free(appNameCopy);
                free(bundleIDCopy);
            });
        }
    }
}

- (void)applicationDidDeactivate:(NSNotification *)notification {
    @autoreleasepool {
        NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
        if (!app) return;

        NSString *appName = app.localizedName ?: @"Unknown";
        NSString *bundleID = app.bundleIdentifier ?: @"unknown.bundle";

        char *appNameCopy = strdup([appName UTF8String]);
        char *bundleIDCopy = strdup([bundleID UTF8String]);

        dispatch_async(dispatch_get_main_queue(), ^{
            if (appNameCopy && bundleIDCopy) {
                handleAppDeactivate(appNameCopy, bundleIDCopy);
            }
            free(appNameCopy);
            free(bundleIDCopy);
        });
    }
}

- (void)screenParametersDidChange:(NSNotification *)notification {
    @autoreleasepool {
        // Debounce to allow system to settle, then invoke Go handler on watcherQueue (not main thread)
        dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.1 * NSEC_PER_SEC)), dispatch_get_global_queue(QOS_CLASS_UTILITY, 0), ^{
            handleScreenParametersChanged();
        });
    }
}

@end

static AppWatcherDelegate *delegate = nil;
static dispatch_queue_t watcherQueue = nil;

void startAppWatcher(void) {
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        watcherQueue = dispatch_queue_create("com.neru.appwatcher", DISPATCH_QUEUE_SERIAL);
    });

    dispatch_sync(watcherQueue, ^{
        if (delegate == nil) {
            delegate = [[AppWatcherDelegate alloc] init];

            NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
            NSNotificationCenter *center = [workspace notificationCenter];

            [center addObserver:delegate
                      selector:@selector(applicationDidLaunch:)
                          name:NSWorkspaceDidLaunchApplicationNotification
                        object:nil];

            [center addObserver:delegate
                      selector:@selector(applicationDidTerminate:)
                          name:NSWorkspaceDidTerminateApplicationNotification
                        object:nil];

            [center addObserver:delegate
                      selector:@selector(applicationDidActivate:)
                          name:NSWorkspaceDidActivateApplicationNotification
                        object:nil];

            [center addObserver:delegate
                      selector:@selector(applicationDidDeactivate:)
                          name:NSWorkspaceDidDeactivateApplicationNotification
                        object:nil];

            // Observe screen parameter changes (display add/remove, resolution changes)
            [[NSNotificationCenter defaultCenter] addObserver:delegate
                                                     selector:@selector(screenParametersDidChange:)
                                                         name:NSApplicationDidChangeScreenParametersNotification
                                                       object:nil];
        }
    });
}

void stopAppWatcher(void) {
    dispatch_sync(watcherQueue, ^{
        if (delegate != nil) {
            NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
            NSNotificationCenter *center = [workspace notificationCenter];
            [center removeObserver:delegate];
            delegate = nil;  // ARC will handle deallocation
        }
    });
}
