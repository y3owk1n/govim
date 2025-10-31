#import "eventtap.h"
#import <Carbon/Carbon.h>

typedef struct {
    CFMachPortRef eventTap;
    CFRunLoopSourceRef runLoopSource;
    EventTapCallback callback;
    void* userData;
} EventTapContext;

CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void* refcon) {
    EventTapContext* context = (EventTapContext*)refcon;
    
    if (type == kCGEventKeyDown) {
        CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        CGEventFlags flags = CGEventGetFlags(event);
        
        // Convert keycode to character
        TISInputSourceRef currentKeyboard = TISCopyCurrentKeyboardInputSource();
        CFDataRef layoutData = TISGetInputSourceProperty(currentKeyboard, kTISPropertyUnicodeKeyLayoutData);
        
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
        
        if (currentKeyboard) CFRelease(currentKeyboard);
        
        // Consume the event (don't pass it through)
        return NULL;
    }
    
    return event;
}

EventTap createEventTap(EventTapCallback callback, void* userData) {
    EventTapContext* context = (EventTapContext*)malloc(sizeof(EventTapContext));
    context->callback = callback;
    context->userData = userData;
    
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

void enableEventTap(EventTap tap) {
    if (!tap) return;
    
    EventTapContext* context = (EventTapContext*)tap;
    CFRunLoopAddSource(CFRunLoopGetCurrent(), context->runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(context->eventTap, true);
}

void disableEventTap(EventTap tap) {
    if (!tap) return;
    
    EventTapContext* context = (EventTapContext*)tap;
    CGEventTapEnable(context->eventTap, false);
    CFRunLoopRemoveSource(CFRunLoopGetCurrent(), context->runLoopSource, kCFRunLoopCommonModes);
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
    
    free(context);
}
