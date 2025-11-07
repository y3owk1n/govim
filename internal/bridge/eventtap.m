#import "eventtap.h"
#import <Carbon/Carbon.h>

typedef struct {
    CFMachPortRef eventTap;
    CFRunLoopSourceRef runLoopSource;
    EventTapCallback callback;
    void* userData;
    NSString** hotkeys;
    int hotkeyCount;
} EventTapContext;

// Helper function to check if current key combination matches a hotkey
BOOL isHotkeyMatch(CGKeyCode keyCode, CGEventFlags flags, NSString* hotkeyString) {
    if (!hotkeyString || [hotkeyString length] == 0) {
        return NO;
    }

    NSArray *parts = [hotkeyString componentsSeparatedByString:@"+"];
    NSString *mainKey = nil;
    BOOL needsCmd = NO, needsShift = NO, needsAlt = NO, needsCtrl = NO;

    // Parse hotkey string
    for (NSString *part in parts) {
        NSString *trimmed = [part stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceCharacterSet]];

        if ([trimmed isEqualToString:@"Cmd"] || [trimmed isEqualToString:@"Command"]) {
            needsCmd = YES;
        } else if ([trimmed isEqualToString:@"Shift"]) {
            needsShift = YES;
        } else if ([trimmed isEqualToString:@"Alt"] || [trimmed isEqualToString:@"Option"]) {
            needsAlt = YES;
        } else if ([trimmed isEqualToString:@"Ctrl"] || [trimmed isEqualToString:@"Control"]) {
            needsCtrl = YES;
        } else {
            mainKey = trimmed;
        }
    }

    if (!mainKey) return NO;

    // Check modifier flags
    BOOL hasCmd = (flags & kCGEventFlagMaskCommand) != 0;
    BOOL hasShift = (flags & kCGEventFlagMaskShift) != 0;
    BOOL hasAlt = (flags & kCGEventFlagMaskAlternate) != 0;
    BOOL hasCtrl = (flags & kCGEventFlagMaskControl) != 0;

    if (needsCmd != hasCmd || needsShift != hasShift || needsAlt != hasAlt || needsCtrl != hasCtrl) {
        return NO;
    }

    // Map key names to key codes (same as in hotkeys.m)
    NSDictionary *keyMap = @{
        @"Space": @(49),
        @"Return": @(36),
        @"Enter": @(36),
        @"Escape": @(53),
        @"Tab": @(48),
        @"Delete": @(51),
        @"Backspace": @(51),

        // Letters
        @"A": @(0), @"B": @(11), @"C": @(8), @"D": @(2), @"E": @(14),
        @"F": @(3), @"G": @(5), @"H": @(4), @"I": @(34), @"J": @(38),
        @"K": @(40), @"L": @(37), @"M": @(46), @"N": @(45), @"O": @(31),
        @"P": @(35), @"Q": @(12), @"R": @(15), @"S": @(1), @"T": @(17),
        @"U": @(32), @"V": @(9), @"W": @(13), @"X": @(7), @"Y": @(16),
        @"Z": @(6),

        // Numbers
        @"0": @(29), @"1": @(18), @"2": @(19), @"3": @(20), @"4": @(21),
        @"5": @(23), @"6": @(22), @"7": @(26), @"8": @(28), @"9": @(25),

        // Function keys
        @"F1": @(122), @"F2": @(120), @"F3": @(99), @"F4": @(118),
        @"F5": @(96), @"F6": @(97), @"F7": @(98), @"F8": @(100),
        @"F9": @(101), @"F10": @(109), @"F11": @(103), @"F12": @(111),

        // Arrow keys
        @"Left": @(123), @"Right": @(124), @"Down": @(125), @"Up": @(126),
    };

    NSNumber *expectedKeyCode = keyMap[mainKey];
    if (!expectedKeyCode) {
        // Try lowercase
        expectedKeyCode = keyMap[[mainKey lowercaseString]];
    }

    return expectedKeyCode && [expectedKeyCode intValue] == keyCode;
}

CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void* refcon) {
    EventTapContext* context = (EventTapContext*)refcon;

    if (type == kCGEventKeyDown) {
        CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        CGEventFlags flags = CGEventGetFlags(event);

        // Check if this key combination matches any of our hotkeys
        // If so, let it pass through to the system (return event instead of NULL)
        for (int i = 0; i < context->hotkeyCount; i++) {
            if (isHotkeyMatch(keyCode, flags, context->hotkeys[i])) {
                return event;
            }
        }

        // Special handling for delete/backspace key (keycode 51)
        if (keyCode == 51) {
            if (context->callback) {
                context->callback("\x7f", context->userData);
            }
            return NULL;
        }

        // Special handling for escape key (keycode 53)
        if (keyCode == 53) {
            if (context->callback) {
                context->callback("\x1b", context->userData);
            }
            return NULL;
        }

        // Always use US keyboard layout for consistent character mapping
        static TISInputSourceRef usKeyboard = NULL;
        if (!usKeyboard) {
            CFStringRef kbdLayoutName = CFSTR("com.apple.keylayout.US");
            CFArrayRef sourceList = TISCreateInputSourceList(
                (CFDictionaryRef)@{ (id)kTISPropertyInputSourceID : (id)kbdLayoutName },
                false);
            if (sourceList && CFArrayGetCount(sourceList) > 0) {
                usKeyboard = (TISInputSourceRef)CFArrayGetValueAtIndex(sourceList, 0);
                CFRetain(usKeyboard);
            }
            if (sourceList) CFRelease(sourceList);
        }

        CFDataRef layoutData = usKeyboard ?
            TISGetInputSourceProperty(usKeyboard, kTISPropertyUnicodeKeyLayoutData) :
            TISGetInputSourceProperty(TISCopyCurrentKeyboardInputSource(), kTISPropertyUnicodeKeyLayoutData);

        if (layoutData) {
            const UCKeyboardLayout* keyboardLayout = (const UCKeyboardLayout*)CFDataGetBytePtr(layoutData);
            UInt32 deadKeyState = 0;
            UniCharCount maxStringLength = 255;
            UniCharCount actualStringLength = 0;
            UniChar unicodeString[maxStringLength];

            // Get modifier flags
            UInt32 modifierKeyState = 0;
            if (flags & kCGEventFlagMaskShift) {
                modifierKeyState |= shiftKey >> 8;
            }
            if (flags & kCGEventFlagMaskControl) {
                modifierKeyState |= controlKey >> 8;
            }

            UCKeyTranslate(keyboardLayout,
                          keyCode,
                          kUCKeyActionDown,
                          modifierKeyState,
                          LMGetKbdType(),
                          kUCKeyTranslateNoDeadKeysMask,
                          &deadKeyState,
                          maxStringLength,
                          &actualStringLength,
                          unicodeString);

            if (actualStringLength > 0) {
                NSString* keyString = [NSString stringWithCharacters:unicodeString length:actualStringLength];
                const char* keyCString = [keyString UTF8String];

                if (context->callback && keyCString) {
                    context->callback(keyCString, context->userData);
                }
            }
        }

        // Consume the event (don't pass it through)
        return NULL;
    }

    return event;
}

EventTap createEventTap(EventTapCallback callback, void* userData) {
    EventTapContext* context = (EventTapContext*)malloc(sizeof(EventTapContext));
    if (!context) return NULL;

    context->callback = callback;
    context->userData = userData;

    // Initialize dynamic hotkeys to empty
    context->hotkeys = NULL;
    context->hotkeyCount = 0;

    // Set up event tap
    CGEventMask eventMask = (1 << kCGEventKeyDown);
    context->eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        eventMask,
        eventTapCallback,
        context
    );

    if (!context->eventTap) {
        free(context);
        return NULL;
    }

    context->runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, context->eventTap, 0);

    return (EventTap)context;
}

void setEventTapHotkeys(EventTap tap, const char** hotkeys, int count) {
    if (!tap) return;
    EventTapContext* context = (EventTapContext*)tap;

    // Free existing
    if (context->hotkeys) {
        for (int i = 0; i < context->hotkeyCount; i++) {
            [context->hotkeys[i] release];
        }
        free(context->hotkeys);
        context->hotkeys = nil;
    }

    context->hotkeyCount = count;
    context->hotkeys = (NSString**)malloc(sizeof(NSString*) * count);

    for (int i = 0; i < count; i++) {
        if (hotkeys[i] && strlen(hotkeys[i]) > 0) {
            context->hotkeys[i] = [[NSString stringWithUTF8String:hotkeys[i]] retain];
        } else {
            context->hotkeys[i] = nil;
        }
    }
}

void enableEventTap(EventTap tap) {
    if (!tap) return;

    EventTapContext* context = (EventTapContext*)tap;

    // Must run on main thread since we're modifying the main run loop
    dispatch_async(dispatch_get_main_queue(), ^{
        CFRunLoopAddSource(CFRunLoopGetMain(), context->runLoopSource, kCFRunLoopCommonModes);
        CGEventTapEnable(context->eventTap, true);
    });
}

void disableEventTap(EventTap tap) {
    if (!tap) return;

    EventTapContext* context = (EventTapContext*)tap;

    // Must run on main thread since we're modifying the main run loop
    dispatch_async(dispatch_get_main_queue(), ^{
        CGEventTapEnable(context->eventTap, false);
        CFRunLoopRemoveSource(CFRunLoopGetMain(), context->runLoopSource, kCFRunLoopCommonModes);
    });
}

void destroyEventTap(EventTap tap) {
    if (!tap) return;

    EventTapContext* context = (EventTapContext*)tap;

    if (context->eventTap) {
        CGEventTapEnable(context->eventTap, false);
        CFRelease(context->eventTap);
    }

    if (context->runLoopSource) {
        CFRelease(context->runLoopSource);
    }

    // Release hotkey strings
    if (context->hotkeys) {
        for (int i = 0; i < context->hotkeyCount; i++) {
            [context->hotkeys[i] release];
        }
        free(context->hotkeys);
    }

    free(context);
}
