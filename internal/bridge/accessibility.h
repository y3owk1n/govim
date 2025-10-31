#ifndef ACCESSIBILITY_H
#define ACCESSIBILITY_H

#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>

// Element information structure
typedef struct {
    CGPoint position;
    CGSize size;
    char* title;
    char* role;
    char* roleDescription;
    bool isEnabled;
    bool isFocused;
    int pid;
} ElementInfo;

// Function declarations
int checkAccessibilityPermissions();
void* getSystemWideElement();
void* getFocusedApplication();
void* getApplicationByPID(int pid);
void* getApplicationByBundleId(const char* bundle_id);
void* getMenuBar(void* app);
ElementInfo* getElementInfo(void* element);
void freeElementInfo(ElementInfo* info);
void* getElementAtPosition(CGPoint position);
int getChildrenCount(void* element);
void** getChildren(void* element, int* count);
void** getVisibleChildren(void* element, int* count);
int performClick(void* element);
int performRightClick(void* element);
int performDoubleClick(void* element);
int performMiddleClick(void* element);
int setFocus(void* element);
char* getElementAttribute(void* element, const char* attribute);
void freeString(char* str);
void releaseElement(void* element);

// Window and application functions
void** getAllWindows(int* count);
void* getFrontmostWindow();
char* getApplicationName(void* app);
char* getBundleIdentifier(void* app);
int setApplicationAttribute(int pid, const char* attribute, int value);

// Scroll functions
int isScrollable(void* element);
CGRect getScrollBounds(void* element);
int scrollElement(void* element, int deltaX, int deltaY);

// Mouse click fallback
int clickElementWithMouse(void* element);

#endif // ACCESSIBILITY_H
