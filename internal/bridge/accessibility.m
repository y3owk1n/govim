#import "accessibility.h"
#import <Cocoa/Cocoa.h>

// Check if accessibility permissions are granted
int checkAccessibilityPermissions() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean trusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    return trusted ? 1 : 0;
}

int setApplicationAttribute(int pid, const char* attribute, int value) {
    if (!attribute) return 0;
    AXUIElementRef appRef = AXUIElementCreateApplication(pid);
    if (!appRef) return 0;

    CFStringRef attrName = CFStringCreateWithCString(NULL, attribute, kCFStringEncodingUTF8);
    if (!attrName) {
        CFRelease(appRef);
        return 0;
    }

    // Check if attribute is settable
    Boolean isSettable = false;
    AXError checkError = AXUIElementIsAttributeSettable(appRef, attrName, &isSettable);
    if (checkError != kAXErrorSuccess || !isSettable) {
        CFRelease(attrName);
        CFRelease(appRef);
        return 0; // Attribute doesn't exist or isn't settable
    }

    CFBooleanRef boolValue = value ? kCFBooleanTrue : kCFBooleanFalse;
    AXError error = AXUIElementSetAttributeValue(appRef, attrName, boolValue);
    CFRelease(attrName);
    CFRelease(appRef);
    return (error == kAXErrorSuccess) ? 1 : 0;
}

// Helper function to check if a point is actually visible (not occluded by other windows)
static bool isPointVisible(CGPoint point, pid_t elementPid) {
    // Use AXUIElementCopyElementAtPosition to check what's at this point
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    if (!systemWide) return true;

    AXUIElementRef elementAtPoint = NULL;
    AXError error = AXUIElementCopyElementAtPosition(systemWide, point.x, point.y, &elementAtPoint);
    CFRelease(systemWide);

    if (error != kAXErrorSuccess || !elementAtPoint) {
        return true; // Assume visible if we can't check
    }

    // Get the PID of the element at this point
    pid_t pidAtPoint;
    bool isVisible = false;

    if (AXUIElementGetPid(elementAtPoint, &pidAtPoint) == kAXErrorSuccess) {
        // The point is visible if it belongs to the same application
        isVisible = (pidAtPoint == elementPid);
    }

    CFRelease(elementAtPoint);
    return isVisible;
}

// Helper function to check if an element is occluded by checking multiple sample points
static bool isElementOccluded(CGRect elementRect, pid_t elementPid) {
    // Sample 5 points: center and 4 corners (slightly inset)
    CGFloat inset = 2.0; // Inset from edges to avoid border issues

    CGPoint samplePoints[5] = {
        // Center
        CGPointMake(elementRect.origin.x + elementRect.size.width / 2,
                   elementRect.origin.y + elementRect.size.height / 2),
        // Top-left
        CGPointMake(elementRect.origin.x + inset,
                   elementRect.origin.y + inset),
        // Top-right
        CGPointMake(elementRect.origin.x + elementRect.size.width - inset,
                   elementRect.origin.y + inset),
        // Bottom-left
        CGPointMake(elementRect.origin.x + inset,
                   elementRect.origin.y + elementRect.size.height - inset),
        // Bottom-right
        CGPointMake(elementRect.origin.x + elementRect.size.width - inset,
                   elementRect.origin.y + elementRect.size.height - inset)
    };

    // Element is considered visible if at least 2 sample points are visible
    int visiblePoints = 0;
    for (int i = 0; i < 5; i++) {
        if (isPointVisible(samplePoints[i], elementPid)) {
            visiblePoints++;
            if (visiblePoints >= 2) {
                return false; // Not occluded
            }
        }
    }

    return true; // Occluded (less than 2 points visible)
}

// Get children that are actually visible on screen
// checkOcclusion: if true, filters out elements covered by other windows
void** getVisibleChildren(void* element, int* count, int checkOcclusion) {
    if (!element || !count) return NULL;

    AXUIElementRef axElement = (AXUIElementRef)element;

    // Get the PID of the element's application
    pid_t elementPid;
    if (AXUIElementGetPid(axElement, &elementPid) != kAXErrorSuccess) {
        *count = 0;
        return NULL;
    }

    // Check if this is the Dock
    NSRunningApplication* app = [NSRunningApplication runningApplicationWithProcessIdentifier:elementPid];
    BOOL isDock = app && [[app bundleIdentifier] isEqualToString:@"com.apple.dock"];

    // First, get all children
    int totalCount = 0;
    void** allChildren = getChildren(element, &totalCount);

    if (!allChildren || totalCount == 0) {
        *count = 0;
        return NULL;
    }

    // For the Dock, skip all filtering and return all children
    if (isDock) {
        *count = totalCount;
        return allChildren;
    }

    // For non-Dock apps, continue with normal filtering...

    // Try to get the parent element's bounds (if available)
    CGRect parentBounds = CGRectZero;
    bool hasParentBounds = false;

    CFTypeRef posValue = NULL, sizeValue = NULL;

    if (AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &posValue) == kAXErrorSuccess && posValue) {
        if (AXValueGetValue(posValue, kAXValueCGPointType, &parentBounds.origin)) {
            if (AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeValue) == kAXErrorSuccess && sizeValue) {
                if (AXValueGetValue(sizeValue, kAXValueCGSizeType, &parentBounds.size)) {
                    hasParentBounds = true;
                }
                CFRelease(sizeValue);
            }
        }
        CFRelease(posValue);
    }

    // Get screen bounds for all displays
    uint32_t displayCount;
    CGDirectDisplayID displays[32];
    CGGetActiveDisplayList(32, displays, &displayCount);

    // Create a union of all display bounds
    CGRect allScreensBounds = CGRectZero;
    for (uint32_t i = 0; i < displayCount; i++) {
        CGRect displayBounds = CGDisplayBounds(displays[i]);
        if (i == 0) {
            allScreensBounds = displayBounds;
        } else {
            allScreensBounds = CGRectUnion(allScreensBounds, displayBounds);
        }
    }

    // Temporary array to hold visible children
    void** visibleChildren = (void**)malloc(totalCount * sizeof(void*));
    int visibleCount = 0;

    for (int i = 0; i < totalCount; i++) {
        AXUIElementRef child = (AXUIElementRef)allChildren[i];

        // Get child's position and size
        CGPoint childPos = CGPointZero;
        CGSize childSize = CGSizeZero;
        bool hasPosition = false, hasSize = false;

        CFTypeRef childPosValue = NULL;
        if (AXUIElementCopyAttributeValue(child, kAXPositionAttribute, &childPosValue) == kAXErrorSuccess && childPosValue) {
            if (AXValueGetValue(childPosValue, kAXValueCGPointType, &childPos)) {
                hasPosition = true;
            }
            CFRelease(childPosValue);
        }

        CFTypeRef childSizeValue = NULL;
        if (AXUIElementCopyAttributeValue(child, kAXSizeAttribute, &childSizeValue) == kAXErrorSuccess && childSizeValue) {
            if (AXValueGetValue(childSizeValue, kAXValueCGSizeType, &childSize)) {
                hasSize = true;
            }
            CFRelease(childSizeValue);
        }

        // If we can't get position/size, include it anyway (might be a container)
        if (!hasPosition && !hasSize) {
            visibleChildren[visibleCount++] = (void*)child;
            continue;
        }

        // Check if child has non-zero size (if size is available)
        if (hasSize && (childSize.width <= 0 || childSize.height <= 0)) {
            CFRelease(child);
            continue;
        }

        // If we have position, do spatial checks
        if (hasPosition && hasSize) {
            CGRect childRect = CGRectMake(childPos.x, childPos.y, childSize.width, childSize.height);

            // Check if child intersects with any screen bounds
            if (!CGRectIntersectsRect(childRect, allScreensBounds)) {
                CFRelease(child);
                continue;
            }

            // If parent has bounds, check if child intersects with parent
            if (hasParentBounds && !CGRectIntersectsRect(childRect, parentBounds)) {
                CFRelease(child);
                continue;
            }

            // Only check occlusion if requested
            if (checkOcclusion && isElementOccluded(childRect, elementPid)) {
                CFRelease(child);
                continue;
            }
        }

        // This child is visible, add it to our result
        visibleChildren[visibleCount++] = (void*)child;
    }

    free(allChildren);

    if (visibleCount == 0) {
        free(visibleChildren);
        *count = 0;
        return NULL;
    }

    // Resize array to actual visible count
    void** result = (void**)realloc(visibleChildren, visibleCount * sizeof(void*));
    *count = visibleCount;

    return result ? result : visibleChildren;
}

// Get system-wide accessibility element
void* getSystemWideElement() {
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    return (void*)systemWide;
}

// Get focused application
void* getFocusedApplication() {
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    if (systemWide) {
        AXUIElementRef focusedApp = NULL;
        AXError error = AXUIElementCopyAttributeValue(
            systemWide,
            kAXFocusedApplicationAttribute,
            (CFTypeRef*)&focusedApp
        );

        CFRelease(systemWide);

        if (error == kAXErrorSuccess && focusedApp) {
            return (void*)focusedApp;
        }
    }

    // Fallback: use NSWorkspace frontmostApplication to obtain an AXUIElement
    NSRunningApplication* front = [NSWorkspace sharedWorkspace].frontmostApplication;
    if (!front) return NULL;
    pid_t pid = front.processIdentifier;
    AXUIElementRef axApp = AXUIElementCreateApplication(pid);
    if (!axApp) return NULL;
    return (void*)axApp;
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

// Check if element has click action
int hasClickAction(void* element) {
    if (!element) return 0;

    AXUIElementRef axElement = (AXUIElementRef)element;
    CFArrayRef actions = NULL;

    AXError error = AXUIElementCopyActionNames(axElement, &actions);
    if (error != kAXErrorSuccess || !actions) {
        return 0;
    }

    CFIndex count = CFArrayGetCount(actions);
    int hasPress = 0;

    for (CFIndex i = 0; i < count; i++) {
        CFStringRef action = (CFStringRef)CFArrayGetValueAtIndex(actions, i);
        if (CFStringCompare(action, kAXPressAction, 0) == kCFCompareEqualTo) {
            hasPress = 1;
            break;
        }
    }

    CFRelease(actions);
    return hasPress;
}

// Perform click
int performClick(void* element) {
    if (!element) return 0;

    AXUIElementRef axElement = (AXUIElementRef)element;

    // Try accessibility action first
    AXError error = AXUIElementPerformAction(axElement, kAXPressAction);
    if (error == kAXErrorSuccess) {
        return 1;
    }

    // Fallback to mouse simulation
    return clickElementWithMouse(element);
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

    AXUIElementRef axElement = (AXUIElementRef)element;

    // Get the element's position
    CFTypeRef positionValue = NULL;
    AXError error = AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionValue);

    if (error != kAXErrorSuccess || !positionValue) {
        return 0;
    }

    CGPoint position;
    if (!AXValueGetValue(positionValue, kAXValueCGPointType, &position)) {
        CFRelease(positionValue);
        return 0;
    }
    CFRelease(positionValue);

    // Get the element's size to click in the center
    CFTypeRef sizeValue = NULL;
    error = AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeValue);

    if (error == kAXErrorSuccess && sizeValue) {
        CGSize size;
        if (AXValueGetValue(sizeValue, kAXValueCGSizeType, &size)) {
            position.x += size.width / 2;
            position.y += size.height / 2;
        }
        CFRelease(sizeValue);
    }

    // Move mouse to position first
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, position, kCGMouseButtonLeft);
    if (move) {
        CGEventPost(kCGHIDEventTap, move);
        CFRelease(move);
        CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false); // Process events
    }

    // Create double-click event
    CGEventRef down1 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, position, kCGMouseButtonLeft);
    CGEventRef up1 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, position, kCGMouseButtonLeft);
    CGEventRef down2 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, position, kCGMouseButtonLeft);
    CGEventRef up2 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, position, kCGMouseButtonLeft);

    if (!down1 || !up1 || !down2 || !up2) {
        if (down1) CFRelease(down1);
        if (up1) CFRelease(up1);
        if (down2) CFRelease(down2);
        if (up2) CFRelease(up2);
        return 0;
    }

    // Set click count to 2 for double-click
    CGEventSetIntegerValueField(down2, kCGMouseEventClickState, 2);
    CGEventSetIntegerValueField(up2, kCGMouseEventClickState, 2);

    // Post the events
    CGEventPost(kCGHIDEventTap, down1);
    CGEventPost(kCGHIDEventTap, up1);
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false); // 50ms between clicks
    CGEventPost(kCGHIDEventTap, down2);
    CGEventPost(kCGHIDEventTap, up2);

    CFRelease(down1);
    CFRelease(up1);
    CFRelease(down2);
    CFRelease(up2);

    // Process events immediately
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);

    return 1;
}

// Perform middle click
int performMiddleClick(void* element) {
    if (!element) return 0;

    AXUIElementRef axElement = (AXUIElementRef)element;

    // Get the element's position
    CFTypeRef positionValue = NULL;
    AXError error = AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionValue);

    if (error != kAXErrorSuccess || !positionValue) {
        return 0;
    }

    CGPoint position;
    if (!AXValueGetValue(positionValue, kAXValueCGPointType, &position)) {
        CFRelease(positionValue);
        return 0;
    }
    CFRelease(positionValue);

    // Get the element's size to click in the center
    CFTypeRef sizeValue = NULL;
    error = AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeValue);

    if (error == kAXErrorSuccess && sizeValue) {
        CGSize size;
        if (AXValueGetValue(sizeValue, kAXValueCGSizeType, &size)) {
            position.x += size.width / 2;
            position.y += size.height / 2;
        }
        CFRelease(sizeValue);
    }

    // Move mouse to position first
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, position, kCGMouseButtonLeft);
    if (move) {
        CGEventPost(kCGHIDEventTap, move);
        CFRelease(move);
        CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false); // Process events
    }

    // Create middle mouse button down event
    CGEventRef mouseDown = CGEventCreateMouseEvent(NULL, kCGEventOtherMouseDown, position, kCGMouseButtonCenter);
    if (!mouseDown) {
        return 0;
    }

    // Create middle mouse button up event
    CGEventRef mouseUp = CGEventCreateMouseEvent(NULL, kCGEventOtherMouseUp, position, kCGMouseButtonCenter);
    if (!mouseUp) {
        CFRelease(mouseDown);
        return 0;
    }

    // Post the events
    CGEventPost(kCGHIDEventTap, mouseDown);
    CGEventPost(kCGHIDEventTap, mouseUp);

    CFRelease(mouseDown);
    CFRelease(mouseUp);

    // Process events immediately
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);

    return 1;
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
    // Try to get the focused application first
    AXUIElementRef focusedApp = (AXUIElementRef)getFocusedApplication();
    AXUIElementRef appRef = focusedApp;
    bool shouldReleaseAppRef = false;

    if (!appRef) {
        // If focused application couldn't be obtained via AX, fall back to NSWorkspace
        NSRunningApplication* front = [NSWorkspace sharedWorkspace].frontmostApplication;
        if (!front) return NULL;
        pid_t pid = front.processIdentifier;
        appRef = AXUIElementCreateApplication(pid);
        if (!appRef) return NULL;
        shouldReleaseAppRef = true;
    }

    AXUIElementRef window = NULL;
    AXError error = AXUIElementCopyAttributeValue(appRef, kAXFocusedWindowAttribute, (CFTypeRef*)&window);

    if (shouldReleaseAppRef && appRef) {
        CFRelease(appRef);
    }

    if (error == kAXErrorSuccess && window) {
        if (focusedApp) {
            CFRelease(focusedApp);
        }
        return (void*)window;
    }

    // As a further fallback, try to get the application's windows list and return the main/first window
    if (focusedApp) {
        pid_t pid;
        if (AXUIElementGetPid(focusedApp, &pid) == kAXErrorSuccess) {
            AXUIElementRef appFromPid = AXUIElementCreateApplication(pid);
            if (appFromPid) {
                CFTypeRef windowsValue = NULL;
                if (AXUIElementCopyAttributeValue(appFromPid, kAXWindowsAttribute, &windowsValue) == kAXErrorSuccess && windowsValue && CFGetTypeID(windowsValue) == CFArrayGetTypeID()) {
                    CFArrayRef windows = (CFArrayRef)windowsValue;
                    if (CFArrayGetCount(windows) > 0) {
                        AXUIElementRef firstWindow = (AXUIElementRef)CFArrayGetValueAtIndex(windows, 0);
                        CFRetain(firstWindow);
                        CFRelease(windowsValue);
                        CFRelease(appFromPid);
                        CFRelease(focusedApp);
                        return (void*)firstWindow;
                    }
                }
                if (windowsValue) CFRelease(windowsValue);
                CFRelease(appFromPid);
            }
        }
        CFRelease(focusedApp);
    }

    return NULL;
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
// This checks if the element actually has scroll capability by looking for scroll bars
int isScrollable(void* element) {
    if (!element) return 0;

    AXUIElementRef axElement = (AXUIElementRef)element;

    // Check if element has horizontal or vertical scroll bars
    CFTypeRef horizontalScrollBar = NULL;
    CFTypeRef verticalScrollBar = NULL;

    int hasScrollBar = 0;

    if (AXUIElementCopyAttributeValue(axElement, kAXHorizontalScrollBarAttribute, &horizontalScrollBar) == kAXErrorSuccess) {
        if (horizontalScrollBar != NULL) {
            hasScrollBar = 1;
            CFRelease(horizontalScrollBar);
        }
    }

    if (!hasScrollBar && AXUIElementCopyAttributeValue(axElement, kAXVerticalScrollBarAttribute, &verticalScrollBar) == kAXErrorSuccess) {
        if (verticalScrollBar != NULL) {
            hasScrollBar = 1;
            CFRelease(verticalScrollBar);
        }
    }

    // If we found scroll bars, it's scrollable
    if (hasScrollBar) {
        return 1;
    }

    // Fallback: check if it's a scroll area role (for compatibility)
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

    // Create and post mouse move event first
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, clickPoint, kCGMouseButtonLeft);
    if (move) {
        CGEventPost(kCGHIDEventTap, move);
        CFRelease(move);
        CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false); // Process events
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

    // Process events immediately
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);

    return 1;
}
