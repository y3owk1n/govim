#ifndef HOTKEYS_H
#define HOTKEYS_H

#import <Foundation/Foundation.h>

// Hotkey callback type
typedef void (*HotkeyCallback)(int hotkeyId, void* userData);

// Modifier keys
typedef enum {
    ModifierNone = 0,
    ModifierCmd = 1 << 0,
    ModifierShift = 1 << 1,
    ModifierAlt = 1 << 2,
    ModifierCtrl = 1 << 3
} ModifierKey;

// Function declarations
int registerHotkey(int keyCode, int modifiers, int hotkeyId, HotkeyCallback callback, void* userData);
void unregisterHotkey(int hotkeyId);
void unregisterAllHotkeys();
int parseKeyString(const char* keyString, int* keyCode, int* modifiers);

#endif // HOTKEYS_H
