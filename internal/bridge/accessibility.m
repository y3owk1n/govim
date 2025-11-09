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

void** getVisibleRows(void* element, int* count) {
    if (!element || !count) return NULL;

    AXUIElementRef axElement = (AXUIElementRef)element;
    CFTypeRef rowsValue = NULL;

    // Try to fetch visible rows
    if (AXUIElementCopyAttributeValue(axElement, kAXVisibleRowsAttribute, &rowsValue) != kAXErrorSuccess) {
        *count = 0;
        return NULL;
    }

    // Ensure the result is an array
    if (CFGetTypeID(rowsValue) != CFArrayGetTypeID()) {
        CFRelease(rowsValue);
        *count = 0;
        return NULL;
    }

    CFArrayRef rows = (CFArrayRef)rowsValue;
    CFIndex rowCount = CFArrayGetCount(rows);
    *count = (int)rowCount;

    void** result = (void**)malloc(rowCount * sizeof(void*));
    if (!result) {
        CFRelease(rowsValue);
        *count = 0;
        return NULL;
    }

    for (CFIndex i = 0; i < rowCount; i++) {
        AXUIElementRef row = (AXUIElementRef)CFArrayGetValueAtIndex(rows, i);
        CFRetain(row);
        result[i] = (void*)row;
    }

    CFRelease(rowsValue);
    return result;
}

static CFStringRef kAXLinkRole            = CFSTR("AXLink");
static CFStringRef kAXCheckboxRole        = CFSTR("AXCheckBox");
static CFStringRef kAXFocusableAttribute  = CFSTR("AXFocusable");
static CFStringRef kAXVisibleAttribute    = CFSTR("AXVisible");

// Check if element has click action
int hasClickAction(void* element) {
    if (!element) return 0;
    AXUIElementRef axElement = (AXUIElementRef)element;

    // Ignore hidden or disabled elements early
    CFBooleanRef hidden = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXHiddenAttribute, (CFTypeRef *)&hidden) == kAXErrorSuccess && hidden) {
        if (CFBooleanGetValue(hidden)) {
            CFRelease(hidden);
            return 0;
        }
        CFRelease(hidden);
    }

    CFBooleanRef enabled = NULL;
    bool isEnabled = true; // default to true if attribute not present
    if (AXUIElementCopyAttributeValue(axElement, kAXEnabledAttribute, (CFTypeRef *)&enabled) == kAXErrorSuccess && enabled) {
        isEnabled = CFBooleanGetValue(enabled);
        CFRelease(enabled);
    }
    if (!isEnabled) return 0;

    // Get role for role-specific fallbacks
    CFStringRef role = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXRoleAttribute, (CFTypeRef *)&role) != kAXErrorSuccess) {
        role = NULL;
    }

    // Explicit actions are the strongest signal
    CFArrayRef actions = NULL;
    if (AXUIElementCopyActionNames(axElement, &actions) == kAXErrorSuccess && actions) {
        CFIndex count = CFArrayGetCount(actions);
        for (CFIndex i = 0; i < count; i++) {
            CFStringRef action = (CFStringRef)CFArrayGetValueAtIndex(actions, i);
            if (CFStringCompare(action, kAXPressAction, 0) == kCFCompareEqualTo ||
                CFStringCompare(action, CFSTR("AXShowMenu"), 0) == kCFCompareEqualTo ||
                CFStringCompare(action, CFSTR("AXConfirm"), 0) == kCFCompareEqualTo ||
                CFStringCompare(action, CFSTR("AXPick"), 0) == kCFCompareEqualTo ||
                CFStringCompare(action, CFSTR("AXRaise"), 0) == kCFCompareEqualTo) {
                CFRelease(actions);
                if (role) CFRelease(role);
                return 1;
            }
        }
        CFRelease(actions);
    }

    // Some elements support AXPress even if not listed in action names
    CFStringRef pressDesc = NULL;
    if (AXUIElementCopyActionDescription(axElement, kAXPressAction, &pressDesc) == kAXErrorSuccess && pressDesc) {
        CFRelease(pressDesc);
        if (role) CFRelease(role);
        return 1;
    }
    if (pressDesc) CFRelease(pressDesc);

    // Focusable and enabled controls are clickable (e.g., text fields)
    CFBooleanRef focusable = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXFocusableAttribute, (CFTypeRef *)&focusable) == kAXErrorSuccess && focusable) {
        if (CFBooleanGetValue(focusable)) {
            CFRelease(focusable);
            // Ensure the element has a hittable rect and is not occluded by other apps
            CGPoint center;
            pid_t pid;
            bool visible = true;
            if (getElementCenter((void*)axElement, &center) && AXUIElementGetPid(axElement, &pid) == kAXErrorSuccess) {
                visible = isPointVisible(center, pid);
            }
            if (role) CFRelease(role);
            return visible ? 1 : 0;
        }
        CFRelease(focusable);
    }

    // Role-specific fallback for links: Treat as clickable if they expose a URL
    if (role && CFStringCompare(role, kAXLinkRole, 0) == kCFCompareEqualTo) {
        CFTypeRef urlAttr = NULL;
        if (AXUIElementCopyAttributeValue(axElement, kAXURLAttribute, &urlAttr) == kAXErrorSuccess && urlAttr) {
            CFRelease(urlAttr);
            CFRelease(role);
            return 1;
        }
        if (urlAttr) CFRelease(urlAttr);
    }

    if (role) CFRelease(role);

    // final check, ensure a visible bounding box and not occluded
    CGPoint center;
    pid_t pid;
    if (getElementCenter((void*)axElement, &center) && AXUIElementGetPid(axElement, &pid) == kAXErrorSuccess) {
        return isPointVisible(center, pid) ? 1 : 0;
    }

    return 0;
}

// Get the center point of an element
int getElementCenter(void* element, CGPoint* outPoint) {
    if (!element || !outPoint) return 0;

    AXUIElementRef axElement = (AXUIElementRef)element;
    *outPoint = CGPointZero;

    CFTypeRef positionRef = NULL;
    AXError error = AXUIElementCopyAttributeValue(axElement, kAXPositionAttribute, &positionRef);

    if (error != kAXErrorSuccess || !positionRef) {
        return 0;
    }

    if (!AXValueGetValue((AXValueRef)positionRef, kAXValueCGPointType, outPoint)) {
        CFRelease(positionRef);
        return 0;
    }
    CFRelease(positionRef);

    // Get size and offset to center
    CFTypeRef sizeRef = NULL;
    if (AXUIElementCopyAttributeValue(axElement, kAXSizeAttribute, &sizeRef) == kAXErrorSuccess && sizeRef) {
        CGSize size;
        if (AXValueGetValue((AXValueRef)sizeRef, kAXValueCGSizeType, &size)) {
            outPoint->x += size.width / 2.0;
            outPoint->y += size.height / 2.0;
        }
        CFRelease(sizeRef);
    }

    return 1;
}

// Move mouse to position
void moveMouse(CGPoint position) {
    CGEventRef move = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, position, kCGMouseButtonLeft);
    if (move) {
        CGEventPost(kCGHIDEventTap, move);
        CFRelease(move);
        CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);
    }
}

// Generic click function
static int performClick(void* element, CGEventType downEvent, CGEventType upEvent, CGMouseButton button, bool restoreCursor) {
    if (!element) return 0;

    CGPoint clickPoint;
    if (!getElementCenter(element, &clickPoint)) {
        return 0;
    }

    // Save current cursor position if needed
    CGPoint originalPosition = CGPointZero;
    if (restoreCursor) {
        CGEventRef event = CGEventCreate(NULL);
        originalPosition = CGEventGetLocation(event);
        CFRelease(event);
    }

    moveMouse(clickPoint);

    CGEventRef down = CGEventCreateMouseEvent(NULL, downEvent, clickPoint, button);
    CGEventRef up = CGEventCreateMouseEvent(NULL, upEvent, clickPoint, button);

    if (!down || !up) {
        if (down) CFRelease(down);
        if (up) CFRelease(up);
        // Restore cursor position before returning if needed
        if (restoreCursor) {
            moveMouse(originalPosition);
        }
        return 0;
    }

    CGEventPost(kCGHIDEventTap, down);
    CGEventPost(kCGHIDEventTap, up);
    CFRelease(down);
    CFRelease(up);

    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);

    // Restore cursor to original position if needed
    if (restoreCursor) {
        moveMouse(originalPosition);
    }

    return 1;
}

// Perform left click
int performLeftClick(void* element, bool restoreCursor) {
    return performClick(element, kCGEventLeftMouseDown, kCGEventLeftMouseUp, kCGMouseButtonLeft, restoreCursor);
}

// Perform right click
int performRightClick(void* element, bool restoreCursor) {
    return performClick(element, kCGEventRightMouseDown, kCGEventRightMouseUp, kCGMouseButtonRight, restoreCursor);
}

// Perform middle click
int performMiddleClick(void* element, bool restoreCursor) {
    return performClick(element, kCGEventOtherMouseDown, kCGEventOtherMouseUp, kCGMouseButtonCenter, restoreCursor);
}

// Perform double click
int performDoubleClick(void* element, bool restoreCursor) {
    if (!element) return 0;

    CGPoint pos;
    if (!getElementCenter(element, &pos)) return 0;

    CGPoint originalPos = CGPointZero;
    if (restoreCursor) {
        CGEventRef ev = CGEventCreate(NULL);
        originalPos = CGEventGetLocation(ev);
        CFRelease(ev);
    }

    moveMouse(pos);

    for (int i = 1; i <= 2; ++i) {
        CGEventRef dn = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pos, kCGMouseButtonLeft);
        CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   pos, kCGMouseButtonLeft);
        if (i == 2) {
            CGEventSetIntegerValueField(dn, kCGMouseEventClickState, 2);
            CGEventSetIntegerValueField(up, kCGMouseEventClickState, 2);
        }
        CGEventPost(kCGHIDEventTap, dn);
        CGEventPost(kCGHIDEventTap, up);
        CFRelease(dn);
        CFRelease(up);
        if (i < 2) CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false);
    }

    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);
    if (restoreCursor) moveMouse(originalPos);
    return 1;
}

// Triple-click
int performTripleClick(void* element, bool restoreCursor) {
    if (!element) return 0;

    CGPoint pos;
    if (!getElementCenter(element, &pos)) return 0;

    CGPoint originalPos = CGPointZero;
    if (restoreCursor) {
        CGEventRef ev = CGEventCreate(NULL);
        originalPos = CGEventGetLocation(ev);
        CFRelease(ev);
    }

    moveMouse(pos);

    for (int i = 1; i <= 3; ++i) {
        CGEventRef dn = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pos, kCGMouseButtonLeft);
        CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   pos, kCGMouseButtonLeft);
        if (i == 3) {
            CGEventSetIntegerValueField(dn, kCGMouseEventClickState, 3);
            CGEventSetIntegerValueField(up, kCGMouseEventClickState, 3);
        }
        CGEventPost(kCGHIDEventTap, dn);
        CGEventPost(kCGHIDEventTap, up);
        CFRelease(dn);
        CFRelease(up);
        if (i < 3) CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false);
    }

    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.01, false);
    if (restoreCursor) moveMouse(originalPos);
    return 1;
}

// Press the left button down (start drag)
int performLeftMouseDown(void *element)
{
    if (!element) return 0;

    CGPoint pos;
    if (!getElementCenter(element, &pos)) return 0;

    moveMouse(pos);
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false);

    CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown,
                                              pos, kCGMouseButtonLeft);
    if (!down) return 0;
    CGEventPost(kCGHIDEventTap, down);
    CFRelease(down);

    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false);
    return 1;
}

// Release the left button (end drag)
int performLeftMouseUp(void *element)
{
    if (!element) return 0;

    CGPoint pos;
    if (!getElementCenter(element, &pos)) return 0;

    moveMouse(pos);
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false);

    CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,
                                            pos, kCGMouseButtonLeft);
    if (!up) return 0;
    CGEventPost(kCGHIDEventTap, up);
    CFRelease(up);

    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, false);
    return 1;
}

// Perform middle click
int performMoveMouseToPosition(void* element) {
    if (!element) return 0;

    CGPoint clickPoint;

    if (!getElementCenter(element, &clickPoint)) {
        return 0;
    }

    moveMouse(clickPoint);

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
    else if (CFGetTypeID(value) == AXValueGetTypeID()) {
        // Handle AXValue types (like AXFrame, AXPosition, AXSize)
        AXValueType valueType = AXValueGetType((AXValueRef)value);

        if (valueType == kAXValueCGRectType) {
            CGRect rect;
            if (AXValueGetValue((AXValueRef)value, kAXValueCGRectType, &rect)) {
                // Format as "{{x, y}, {width, height}}"
                result = malloc(128);
                snprintf(result, 128, "{{%.1f, %.1f}, {%.1f, %.1f}}",
                        rect.origin.x, rect.origin.y,
                        rect.size.width, rect.size.height);
            }
        }
        else if (valueType == kAXValueCGPointType) {
            CGPoint point;
            if (AXValueGetValue((AXValueRef)value, kAXValueCGPointType, &point)) {
                result = malloc(64);
                snprintf(result, 64, "{%.1f, %.1f}", point.x, point.y);
            }
        }
        else if (valueType == kAXValueCGSizeType) {
            CGSize size;
            if (AXValueGetValue((AXValueRef)value, kAXValueCGSizeType, &size)) {
                result = malloc(64);
                snprintf(result, 64, "{%.1f, %.1f}", size.width, size.height);
            }
        }
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

// Scroll at current cursor position only
int scrollAtCursor(int deltaX, int deltaY) {
    CGEventRef event = CGEventCreate(NULL);
    CGPoint cursorPos = CGEventGetLocation(event);
    CFRelease(event);

    CGEventRef scrollEvent = CGEventCreateScrollWheelEvent(
        NULL,
        kCGScrollEventUnitPixel,
        2,
        deltaY,
        deltaX
    );

    if (scrollEvent) {
        CGEventSetLocation(scrollEvent, cursorPos);
        CGEventPost(kCGHIDEventTap, scrollEvent);
        CFRelease(scrollEvent);
        return 1;
    }

    return 0;
}

// Try to detect if Mission Control is active
// Works on Sequoia 15.1 (Tahoe)
// Maybe works on older versions as per stackoverflow result (https://stackoverflow.com/questions/12683225/osx-how-to-detect-if-mission-control-is-running)
bool isMissionControlActive() {
    bool result = false;

    CFArrayRef windowList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionAll,
        kCGNullWindowID
    );

    if (!windowList) {
        return false;
    }

    // Get screen size
    NSScreen *mainScreen = [NSScreen mainScreen];
    if (!mainScreen) {
        CFRelease(windowList);
        return false;
    }

    CGSize screenSize = mainScreen.frame.size;
    CFIndex count = CFArrayGetCount(windowList);
    int fullscreenDockWindows = 0;
    int highLayerDockWindows = 0;

    for (CFIndex i = 0; i < count; i++) {
        CFDictionaryRef windowInfo = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);
        if (!windowInfo) continue;

        CFStringRef ownerName = (CFStringRef)CFDictionaryGetValue(
            windowInfo,
            kCGWindowOwnerName
        );

        if (ownerName && CFStringCompare(ownerName, CFSTR("Dock"), 0) == kCFCompareEqualTo) {
            CFStringRef windowName = (CFStringRef)CFDictionaryGetValue(
                windowInfo,
                kCGWindowName
            );

            CFDictionaryRef bounds = (CFDictionaryRef)CFDictionaryGetValue(
                windowInfo,
                kCGWindowBounds
            );

            CFNumberRef windowLayer = (CFNumberRef)CFDictionaryGetValue(
                windowInfo,
                kCGWindowLayer
            );

            if (bounds) {
                double x = 0, y = 0, w = 0, h = 0;

                CFNumberRef xValue = (CFNumberRef)CFDictionaryGetValue(bounds, CFSTR("X"));
                CFNumberRef yValue = (CFNumberRef)CFDictionaryGetValue(bounds, CFSTR("Y"));
                CFNumberRef wValue = (CFNumberRef)CFDictionaryGetValue(bounds, CFSTR("Width"));
                CFNumberRef hValue = (CFNumberRef)CFDictionaryGetValue(bounds, CFSTR("Height"));

                if (xValue) CFNumberGetValue(xValue, kCFNumberDoubleType, &x);
                if (yValue) CFNumberGetValue(yValue, kCFNumberDoubleType, &y);
                if (wValue) CFNumberGetValue(wValue, kCFNumberDoubleType, &w);
                if (hValue) CFNumberGetValue(hValue, kCFNumberDoubleType, &h);

                // check for Y < 0 (works on older macOS... maybe?)
                if (y < 0) {
                    result = true;
                    break;
                }

                // count fullscreen Dock windows (works on Sequoia 15.1)
                bool isFullscreen = (w >= screenSize.width * 0.95 && h >= screenSize.height * 0.95);
                bool hasNoName = (!windowName || CFStringGetLength(windowName) == 0);

                if (isFullscreen && hasNoName && windowLayer) {
                    int layer = 0;
                    CFNumberGetValue(windowLayer, kCFNumberIntType, &layer);

                    fullscreenDockWindows++;
                    if (layer >= 18 && layer <= 20) {
                        highLayerDockWindows++;
                    }
                }
            }
        }
    }

    CFRelease(windowList);

    // Try to return the results from old or new OS method
    if (!result && fullscreenDockWindows >= 2 && highLayerDockWindows >= 2) {
        result = true;
    }

    return result;
}
