#import "accessibility.h"
#import <Cocoa/Cocoa.h>

// Check if accessibility permissions are granted
int checkAccessibilityPermissions() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean trusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    return trusted ? 1 : 0;
}

// Set arbitrary application attribute (e.g. AXManualAccessibility)
int setApplicationAttribute(int pid, const char* attribute, int value) {
    if (!attribute) return 0;

    AXUIElementRef appRef = AXUIElementCreateApplication(pid);
    if (!appRef) {
        return 0;
    }

    CFStringRef attrName = CFStringCreateWithCString(NULL, attribute, kCFStringEncodingUTF8);
    if (!attrName) {
        CFRelease(appRef);
        return 0;
    }

    CFBooleanRef boolValue = value ? kCFBooleanTrue : kCFBooleanFalse;
    AXError error = AXUIElementSetAttributeValue(appRef, attrName, boolValue);

    CFRelease(attrName);
    CFRelease(appRef);

    return (error == kAXErrorSuccess) ? 1 : 0;
}

void** getVisibleChildren(void* element, int* count) {
    if (!element || !count) return NULL;

    AXUIElementRef axElement = (AXUIElementRef)element;
    CFTypeRef visibleChildrenValue = NULL;

    if (AXUIElementCopyAttributeValue(axElement, kAXVisibleChildrenAttribute, &visibleChildrenValue) == kAXErrorSuccess && visibleChildrenValue) {
        if (CFGetTypeID(visibleChildrenValue) == CFArrayGetTypeID()) {
            CFArrayRef children = (CFArrayRef)visibleChildrenValue;
            CFIndex childCount = CFArrayGetCount(children);
            *count = (int)childCount;

            void** result = (void**)malloc(childCount * sizeof(void*));

            for (CFIndex i = 0; i < childCount; i++) {
                AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(children, i);
                CFRetain(child);
                result[i] = (void*)child;
            }

            CFRelease(visibleChildrenValue);
            return result;
        }

        CFRelease(visibleChildrenValue);
    }

    return getChildren(element, count);
}

// Get system-wide accessibility element
void* getSystemWideElement() {
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    return (void*)systemWide;
}

// Get focused application
void* getFocusedApplication() {
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    if (!systemWide) return NULL;
    
    AXUIElementRef focusedApp = NULL;
    AXError error = AXUIElementCopyAttributeValue(
        systemWide,
        kAXFocusedApplicationAttribute,
        (CFTypeRef*)&focusedApp
    );
    
    CFRelease(systemWide);
    
    if (error != kAXErrorSuccess) {
        return NULL;
    }
    
    return (void*)focusedApp;
}

// Get the menu bar element of an application
void* getMenuBar(void* app) {
    if (!app) return NULL;
    AXUIElementRef axApp = (AXUIElementRef)app;
    AXUIElementRef menubar = NULL;
    AXError error = AXUIElementCopyAttributeValue(axApp, kAXMenuBarAttribute, (CFTypeRef*)&menubar);
    if (error != kAXErrorSuccess) {
        return NULL;
    }
    return (void*)menubar;
}

// Get application by PID
void* getApplicationByPID(int pid) {
    AXUIElementRef app = AXUIElementCreateApplication(pid);
    return (void*)app;
}

// Get application by bundle identifier
void* getApplicationByBundleId(const char* bundle_id) {
    if (!bundle_id) return NULL;

    NSString* bundleIdStr = [NSString stringWithUTF8String:bundle_id];
    NSArray<NSRunningApplication*>* apps = [NSRunningApplication runningApplicationsWithBundleIdentifier:bundleIdStr];
    if (apps.count == 0) {
        return NULL;
    }
    NSRunningApplication* app = apps.firstObject;
    pid_t pid = app.processIdentifier;
    AXUIElementRef axApp = AXUIElementCreateApplication(pid);
    return (void*)axApp;
}

// Helper function to convert CFStringRef to C string
char* cfStringToCString(CFStringRef cfStr) {
    if (!cfStr) return NULL;
    
    CFIndex length = CFStringGetLength(cfStr);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
    char* buffer = (char*)malloc(maxSize);
    
    if (CFStringGetCString(cfStr, buffer, maxSize, kCFStringEncodingUTF8)) {
        return buffer;
    }
    
    free(buffer);
    return NULL;
}

// Get element information
ElementInfo* getElementInfo(void* element) {
    if (!element) return NULL;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    ElementInfo* info = (ElementInfo*)calloc(1, sizeof(ElementInfo));
    
    // Get position
    CFTypeRef positionValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionValue) == kAXErrorSuccess) {
        CGPoint point;
        if (AXValueGetValue(positionValue, kAXValueCGPointType, &point)) {
            info->position = point;
        }
        CFRelease(positionValue);
    }
    
    // Get size
    CFTypeRef sizeValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeValue) == kAXErrorSuccess) {
        CGSize size;
        if (AXValueGetValue(sizeValue, kAXValueCGSizeType, &size)) {
            info->size = size;
        }
        CFRelease(sizeValue);
    }
    
    // Get title
    CFTypeRef titleValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXTitleAttribute, &titleValue) == kAXErrorSuccess) {
        if (CFGetTypeID(titleValue) == CFStringGetTypeID()) {
            info->title = cfStringToCString((CFStringRef)titleValue);
        }
        CFRelease(titleValue);
    }
    
    // Get role
    CFTypeRef roleValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXRoleAttribute, &roleValue) == kAXErrorSuccess) {
        if (CFGetTypeID(roleValue) == CFStringGetTypeID()) {
            info->role = cfStringToCString((CFStringRef)roleValue);
        }
        CFRelease(roleValue);
    }
    
    // Get role description
    CFTypeRef roleDescValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXRoleDescriptionAttribute, &roleDescValue) == kAXErrorSuccess) {
        if (CFGetTypeID(roleDescValue) == CFStringGetTypeID()) {
            info->roleDescription = cfStringToCString((CFStringRef)roleDescValue);
        }
        CFRelease(roleDescValue);
    }
    
    // Get enabled state
    CFTypeRef enabledValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXEnabledAttribute, &enabledValue) == kAXErrorSuccess) {
        if (CFGetTypeID(enabledValue) == CFBooleanGetTypeID()) {
            info->isEnabled = CFBooleanGetValue((CFBooleanRef)enabledValue);
        }
        CFRelease(enabledValue);
    }
    
    // Get focused state
    CFTypeRef focusedValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXFocusedAttribute, &focusedValue) == kAXErrorSuccess) {
        if (CFGetTypeID(focusedValue) == CFBooleanGetTypeID()) {
            info->isFocused = CFBooleanGetValue((CFBooleanRef)focusedValue);
        }
        CFRelease(focusedValue);
    }
    
    // Get PID
    pid_t pid;
    if (AXUIElementGetPid(axElement, &pid) == kAXErrorSuccess) {
        info->pid = pid;
    }
    
    return info;
}

// Free element info
void freeElementInfo(ElementInfo* info) {
    if (!info) return;
    
    if (info->title) free(info->title);
    if (info->role) free(info->role);
    if (info->roleDescription) free(info->roleDescription);
    free(info);
}

// Get element at position
void* getElementAtPosition(CGPoint position) {
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    if (!systemWide) return NULL;
    
    AXUIElementRef element = NULL;
    AXError error = AXUIElementCopyElementAtPosition(systemWide, position.x, position.y, &element);
    
    CFRelease(systemWide);
    
    if (error != kAXErrorSuccess) {
        return NULL;
    }
    
    return (void*)element;
}

// Get children count
int getChildrenCount(void* element) {
    if (!element) return 0;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    CFTypeRef childrenValue = NULL;
    
    if (AXUIElementCopyAttributeValue(axElement, kAXChildrenAttribute, &childrenValue) != kAXErrorSuccess) {
        return 0;
    }
    
    if (CFGetTypeID(childrenValue) != CFArrayGetTypeID()) {
        CFRelease(childrenValue);
        return 0;
    }
    
    CFIndex count = CFArrayGetCount((CFArrayRef)childrenValue);
    CFRelease(childrenValue);
    
    return (int)count;
}

// Get children
void** getChildren(void* element, int* count) {
    if (!element || !count) return NULL;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    CFTypeRef childrenValue = NULL;
    
    if (AXUIElementCopyAttributeValue(axElement, kAXChildrenAttribute, &childrenValue) != kAXErrorSuccess) {
        *count = 0;
        return NULL;
    }
    
    if (CFGetTypeID(childrenValue) != CFArrayGetTypeID()) {
        CFRelease(childrenValue);
        *count = 0;
        return NULL;
    }
    
    CFArrayRef children = (CFArrayRef)childrenValue;
    CFIndex childCount = CFArrayGetCount(children);
    *count = (int)childCount;
    
    void** result = (void**)malloc(childCount * sizeof(void*));
    
    for (CFIndex i = 0; i < childCount; i++) {
        AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex(children, i);
        CFRetain(child);
        result[i] = (void*)child;
    }
    
    CFRelease(childrenValue);
    return result;
}

// Perform click
int performClick(void* element) {
    if (!element) return 0;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    AXError error = AXUIElementPerformAction(axElement, kAXPressAction);
    
    return (error == kAXErrorSuccess) ? 1 : 0;
}

// Perform right click
int performRightClick(void* element) {
    if (!element) return 0;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    AXError error = AXUIElementPerformAction(axElement, kAXShowMenuAction);
    
    return (error == kAXErrorSuccess) ? 1 : 0;
}

// Perform double click
int performDoubleClick(void* element) {
    if (!element) return 0;
    
    // Double click is typically just two single clicks
    int result = performClick(element);
    if (result) {
        usleep(50000); // 50ms delay
        result = performClick(element);
    }
    
    return result;
}

// Set focus
int setFocus(void* element) {
    if (!element) return 0;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    CFBooleanRef trueValue = kCFBooleanTrue;
    AXError error = AXUIElementSetAttributeValue(axElement, kAXFocusedAttribute, trueValue);
    
    return (error == kAXErrorSuccess) ? 1 : 0;
}

// Get element attribute
char* getElementAttribute(void* element, const char* attribute) {
    if (!element || !attribute) return NULL;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    CFStringRef attrName = CFStringCreateWithCString(NULL, attribute, kCFStringEncodingUTF8);
    
    CFTypeRef value = NULL;
    AXError error = AXUIElementCopyAttributeValue(axElement, attrName, &value);
    CFRelease(attrName);
    
    if (error != kAXErrorSuccess || !value) {
        return NULL;
    }
    
    char* result = NULL;
    if (CFGetTypeID(value) == CFStringGetTypeID()) {
        result = cfStringToCString((CFStringRef)value);
    }
    
    CFRelease(value);
    return result;
}

// Free string
void freeString(char* str) {
    if (str) free(str);
}

// Release element
void releaseElement(void* element) {
    if (element) {
        CFRelease((AXUIElementRef)element);
    }
}

// Get all windows
void** getAllWindows(int* count) {
    if (!count) return NULL;
    
    AXUIElementRef focusedApp = (AXUIElementRef)getFocusedApplication();
    if (!focusedApp) {
        *count = 0;
        return NULL;
    }
    
    CFTypeRef windowsValue = NULL;
    if (AXUIElementCopyAttributeValue(focusedApp, kAXWindowsAttribute, &windowsValue) != kAXErrorSuccess) {
        CFRelease(focusedApp);
        *count = 0;
        return NULL;
    }
    
    if (CFGetTypeID(windowsValue) != CFArrayGetTypeID()) {
        CFRelease(windowsValue);
        CFRelease(focusedApp);
        *count = 0;
        return NULL;
    }
    
    CFArrayRef windows = (CFArrayRef)windowsValue;
    CFIndex windowCount = CFArrayGetCount(windows);
    *count = (int)windowCount;
    
    void** result = (void**)malloc(windowCount * sizeof(void*));
    
    for (CFIndex i = 0; i < windowCount; i++) {
        AXUIElementRef window = (AXUIElementRef)CFArrayGetValueAtIndex(windows, i);
        CFRetain(window);
        result[i] = (void*)window;
    }
    
    CFRelease(windowsValue);
    CFRelease(focusedApp);
    return result;
}

// Get frontmost window
void* getFrontmostWindow() {
    AXUIElementRef focusedApp = (AXUIElementRef)getFocusedApplication();
    if (!focusedApp) return NULL;
    
    AXUIElementRef window = NULL;
    AXError error = AXUIElementCopyAttributeValue(focusedApp, kAXFocusedWindowAttribute, (CFTypeRef*)&window);
    
    CFRelease(focusedApp);
    
    if (error != kAXErrorSuccess) {
        return NULL;
    }
    
    return (void*)window;
}

// Get application name
char* getApplicationName(void* app) {
    if (!app) return NULL;
    
    AXUIElementRef axApp = (AXUIElementRef)app;
    CFTypeRef titleValue = NULL;
    
    if (AXUIElementCopyAttributeValue(axApp, kAXTitleAttribute, &titleValue) != kAXErrorSuccess) {
        return NULL;
    }
    
    char* result = NULL;
    if (CFGetTypeID(titleValue) == CFStringGetTypeID()) {
        result = cfStringToCString((CFStringRef)titleValue);
    }
    
    CFRelease(titleValue);
    return result;
}

// Get bundle identifier
char* getBundleIdentifier(void* app) {
    if (!app) return NULL;
    
    pid_t pid;
    if (AXUIElementGetPid((AXUIElementRef)app, &pid) != kAXErrorSuccess) {
        return NULL;
    }
    
    NSRunningApplication* runningApp = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
    if (!runningApp) return NULL;
    
    NSString* bundleId = [runningApp bundleIdentifier];
    if (!bundleId) return NULL;
    
    return cfStringToCString((__bridge CFStringRef)bundleId);
}

// Check if element is scrollable
int isScrollable(void* element) {
    if (!element) return 0;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    
    // Check if element has scroll bars or is a scroll area
    CFTypeRef roleValue = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXRoleAttribute, &roleValue) == kAXErrorSuccess) {
        if (CFGetTypeID(roleValue) == CFStringGetTypeID()) {
            CFStringRef role = (CFStringRef)roleValue;
            if (CFStringCompare(role, kAXScrollAreaRole, 0) == kCFCompareEqualTo) {
                CFRelease(roleValue);
                return 1;
            }
        }
        CFRelease(roleValue);
    }
    
    return 0;
}

// Get scroll bounds
CGRect getScrollBounds(void* element) {
    CGRect rect = CGRectZero;
    if (!element) return rect;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    
    CFTypeRef positionValue = NULL;
    CFTypeRef sizeValue = NULL;
    
    if (AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionValue) == kAXErrorSuccess) {
        CGPoint point;
        if (AXValueGetValue(positionValue, kAXValueCGPointType, &point)) {
            rect.origin = point;
        }
        CFRelease(positionValue);
    }
    
    if (AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeValue) == kAXErrorSuccess) {
        CGSize size;
        if (AXValueGetValue(sizeValue, kAXValueCGSizeType, &size)) {
            rect.size = size;
        }
        CFRelease(sizeValue);
    }
    
    return rect;
}

// Scroll element by simulating scroll wheel events (like Vimac does)
int scrollElement(void* element, int deltaX, int deltaY) {
    if (!element) return 0;
    
    AXUIElementRef axElement = (AXUIElementRef)element;
    
    // Get the position of the element to scroll at that location
    CFTypeRef positionRef = NULL;
    CGPoint scrollPoint = CGPointZero;
    
    if (AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionRef) == kAXErrorSuccess) {
        if (positionRef) {
            AXValueGetValue((AXValueRef)positionRef, kAXValueCGPointType, &scrollPoint);
            CFRelease(positionRef);
            
            // Offset to center of element
            CFTypeRef sizeRef = NULL;
            if (AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeRef) == kAXErrorSuccess) {
                if (sizeRef) {
                    CGSize size;
                    AXValueGetValue((AXValueRef)sizeRef, kAXValueCGSizeType, &size);
                    scrollPoint.x += size.width / 2;
                    scrollPoint.y += size.height / 2;
                    CFRelease(sizeRef);
                }
            }
        }
    }
    
    // Create scroll wheel event using pixel units for more reliable scrolling
    CGEventRef scrollEvent = CGEventCreateScrollWheelEvent(
        NULL,
        kCGScrollEventUnitPixel,  // Use pixels instead of lines
        2, // number of wheels (vertical and horizontal)
        deltaY,
        deltaX
    );
    
    if (scrollEvent) {
        // Set the event location to the scroll area center
        CGEventSetLocation(scrollEvent, scrollPoint);
        
        // Post the event to the HID event tap
        CGEventPost(kCGHIDEventTap, scrollEvent);
        CFRelease(scrollEvent);
        
        return 1;
    }
    
    return 0;
}

// Perform a real mouse left-click at the center of the element
int clickElementWithMouse(void* element) {
    if (!element) return 0;

    AXUIElementRef axElement = (AXUIElementRef)element;

    // Determine the center point of the element
    CGPoint clickPoint = CGPointZero;

    CFTypeRef positionRef = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionRef) == kAXErrorSuccess && positionRef) {
        AXValueGetValue((AXValueRef)positionRef, kAXValueCGPointType, &clickPoint);
        CFRelease(positionRef);

        CFTypeRef sizeRef = NULL;
        if (AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeRef) == kAXErrorSuccess && sizeRef) {
            CGSize size;
            AXValueGetValue((AXValueRef)sizeRef, kAXValueCGSizeType, &size);
            clickPoint.x += size.width / 2.0;
            clickPoint.y += size.height / 2.0;
            CFRelease(sizeRef);
        }
    }

    // Create and post mouse move, down, and up events
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, clickPoint, kCGMouseButtonLeft);
    if (move) {
        CGEventPost(kCGHIDEventTap, move);
        CFRelease(move);
    }

    CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, clickPoint, kCGMouseButtonLeft);
    CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, clickPoint, kCGMouseButtonLeft);

    if (!down || !up) {
        if (down) CFRelease(down);
        if (up) CFRelease(up);
        return 0;
    }

    CGEventPost(kCGHIDEventTap, down);
    CGEventPost(kCGHIDEventTap, up);
    CFRelease(down);
    CFRelease(up);

    return 1;
}
