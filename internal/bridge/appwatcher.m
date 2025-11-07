#import "appwatcher.h"
#import <Cocoa/Cocoa.h>

extern void handleAppLaunch(const char* appName, const char* bundleID);
extern void handleAppTerminate(const char* appName, const char* bundleID);
extern void handleAppActivate(const char* appName, const char* bundleID);
extern void handleAppDeactivate(const char* appName, const char* bundleID);

@interface AppWatcherDelegate : NSObject
@end

@implementation AppWatcherDelegate

- (void)applicationDidLaunch:(NSNotification *)notification {
    NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
    const char *appName = [app.localizedName UTF8String];
    const char *bundleID = [app.bundleIdentifier UTF8String];
    handleAppLaunch(appName, bundleID);
}

- (void)applicationDidTerminate:(NSNotification *)notification {
    NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
    const char *appName = [app.localizedName UTF8String];
    const char *bundleID = [app.bundleIdentifier UTF8String];
    handleAppTerminate(appName, bundleID);
}

- (void)applicationDidActivate:(NSNotification *)notification {
    NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
    const char *appName = [app.localizedName UTF8String];
    const char *bundleID = [app.bundleIdentifier UTF8String];
    handleAppActivate(appName, bundleID);
}

- (void)applicationDidDeactivate:(NSNotification *)notification {
    NSRunningApplication *app = notification.userInfo[NSWorkspaceApplicationKey];
    const char *appName = [app.localizedName UTF8String];
    const char *bundleID = [app.bundleIdentifier UTF8String];
    handleAppDeactivate(appName, bundleID);
}

@end

static AppWatcherDelegate *delegate = nil;

void startAppWatcher(void) {
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
    }
}

void stopAppWatcher(void) {
    if (delegate != nil) {
        NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
        NSNotificationCenter *center = [workspace notificationCenter];
        [center removeObserver:delegate];
        delegate = nil;
    }
}