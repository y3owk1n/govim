#import "hotkeys.h"
#import <Carbon/Carbon.h>
#import <Cocoa/Cocoa.h>

static NSMutableDictionary *hotkeyRefs = nil;
static NSMutableDictionary *hotkeyCallbacks = nil;

// Initialize storage
static void initializeStorage() {
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        hotkeyRefs = [NSMutableDictionary dictionary];
        hotkeyCallbacks = [NSMutableDictionary dictionary];
    });
}

// Hotkey handler
OSStatus hotkeyHandler(EventHandlerCallRef nextHandler, EventRef event, void *userData) {
    EventHotKeyID hotkeyID;
    GetEventParameter(event, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof(hotkeyID), NULL, &hotkeyID);
    
    int hotkeyId = (int)hotkeyID.id;
    NSNumber *key = @(hotkeyId);
    
    NSDictionary *callbackInfo = hotkeyCallbacks[key];
    if (callbackInfo) {
        HotkeyCallback callback = [callbackInfo[@"callback"] pointerValue];
        void *userData = [callbackInfo[@"userData"] pointerValue];
        
        if (callback) {
            callback(hotkeyId, userData);
        }
    }
    
    return noErr;
}

// Register hotkey
int registerHotkey(int keyCode, int modifiers, int hotkeyId, HotkeyCallback callback, void* userData) {
    initializeStorage();
    
    // Convert modifiers
    UInt32 carbonModifiers = 0;
    if (modifiers & ModifierCmd) carbonModifiers |= cmdKey;
    if (modifiers & ModifierShift) carbonModifiers |= shiftKey;
    if (modifiers & ModifierAlt) carbonModifiers |= optionKey;
    if (modifiers & ModifierCtrl) carbonModifiers |= controlKey;
    
    // Create hotkey ID
    EventHotKeyID hotkeyID;
    hotkeyID.signature = 'gvim';
    hotkeyID.id = hotkeyId;
    
    // Register hotkey
    EventHotKeyRef hotkeyRef;
    EventTypeSpec eventType;
    eventType.eventClass = kEventClassKeyboard;
    eventType.eventKind = kEventHotKeyPressed;
    
    InstallApplicationEventHandler(&hotkeyHandler, 1, &eventType, NULL, NULL);
    
    OSStatus status = RegisterEventHotKey(keyCode, carbonModifiers, hotkeyID,
                                         GetApplicationEventTarget(), 0, &hotkeyRef);
    
    if (status != noErr) {
        return 0;
    }
    
    // Store reference and callback
    NSNumber *key = @(hotkeyId);
    hotkeyRefs[key] = [NSValue valueWithPointer:hotkeyRef];
    
    NSDictionary *callbackInfo = @{
        @"callback": [NSValue valueWithPointer:callback],
        @"userData": [NSValue valueWithPointer:userData]
    };
    hotkeyCallbacks[key] = callbackInfo;
    
    return 1;
}

// Unregister hotkey
void unregisterHotkey(int hotkeyId) {
    if (!hotkeyRefs) return;
    
    NSNumber *key = @(hotkeyId);
    NSValue *refValue = hotkeyRefs[key];
    
    if (refValue) {
        EventHotKeyRef hotkeyRef = [refValue pointerValue];
        UnregisterEventHotKey(hotkeyRef);
        
        [hotkeyRefs removeObjectForKey:key];
        [hotkeyCallbacks removeObjectForKey:key];
    }
}

// Unregister all hotkeys
void unregisterAllHotkeys() {
    if (!hotkeyRefs) return;
    
    for (NSNumber *key in [hotkeyRefs allKeys]) {
        NSValue *refValue = hotkeyRefs[key];
        EventHotKeyRef hotkeyRef = [refValue pointerValue];
        UnregisterEventHotKey(hotkeyRef);
    }
    
    [hotkeyRefs removeAllObjects];
    [hotkeyCallbacks removeAllObjects];
}

// Parse key string (e.g., "Cmd+Shift+Space")
int parseKeyString(const char* keyString, int* keyCode, int* modifiers) {
    if (!keyString || !keyCode || !modifiers) return 0;
    
    NSString *keyStr = @(keyString);
    NSArray *parts = [keyStr componentsSeparatedByString:@"+"];
    
    *modifiers = ModifierNone;
    NSString *mainKey = nil;
    
    for (NSString *part in parts) {
        NSString *trimmed = [part stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceCharacterSet]];
        
        if ([trimmed isEqualToString:@"Cmd"] || [trimmed isEqualToString:@"Command"]) {
            *modifiers |= ModifierCmd;
        } else if ([trimmed isEqualToString:@"Shift"]) {
            *modifiers |= ModifierShift;
        } else if ([trimmed isEqualToString:@"Alt"] || [trimmed isEqualToString:@"Option"]) {
            *modifiers |= ModifierAlt;
        } else if ([trimmed isEqualToString:@"Ctrl"] || [trimmed isEqualToString:@"Control"]) {
            *modifiers |= ModifierCtrl;
        } else {
            mainKey = trimmed;
        }
    }
    
    if (!mainKey) return 0;
    
    // Map key names to key codes
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
    
    NSNumber *keyCodeNum = keyMap[mainKey];
    if (!keyCodeNum) {
        // Try lowercase
        keyCodeNum = keyMap[[mainKey lowercaseString]];
    }
    
    if (!keyCodeNum) return 0;
    
    *keyCode = [keyCodeNum intValue];
    return 1;
}
