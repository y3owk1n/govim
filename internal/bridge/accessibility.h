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
void** getVisibleRows(void* element, int* count);

int getElementCenter(void* element, CGPoint* outPoint);

void moveMouse(CGPoint position);

// Click functions - perform click actions on accessibility elements and restore cursor position
int performLeftClick(void* element, bool restoreCursor);
int performRightClick(void* element, bool restoreCursor);
int performDoubleClick(void* element, bool restoreCursor);
int performMiddleClick(void* element, bool restoreCursor);

int performMoveMouseToPosition(void* element);

int hasClickAction(void* element);
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

#endif // ACCESSIBILITY_H
