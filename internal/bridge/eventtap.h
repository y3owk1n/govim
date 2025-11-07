#ifndef EVENTTAP_H
#define EVENTTAP_H

#import <Foundation/Foundation.h>

// Event tap callback type
typedef void (*EventTapCallback)(const char* key, void* userData);

// Event tap handle
typedef void* EventTap;

// Function declarations
EventTap createEventTap(EventTapCallback callback, void* userData);
void enableEventTap(EventTap tap);
void disableEventTap(EventTap tap);
void destroyEventTap(EventTap tap);
void setEventTapHotkeys(EventTap tap, const char** hotkeys, int count);

#endif // EVENTTAP_H
