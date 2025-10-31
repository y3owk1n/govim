package accessibility

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/accessibility.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"image"
	"unsafe"
)

// Element represents an accessibility UI element
type Element struct {
	ref unsafe.Pointer
}

// ElementInfo contains information about a UI element
type ElementInfo struct {
	Position        image.Point
	Size            image.Point
	Title           string
	Role            string
	RoleDescription string
	IsEnabled       bool
	IsFocused       bool
	PID             int
}

// CheckAccessibilityPermissions checks if the app has accessibility permissions
func CheckAccessibilityPermissions() bool {
	result := C.checkAccessibilityPermissions()
	return result == 1
}

// GetSystemWideElement returns the system-wide accessibility element
func GetSystemWideElement() *Element {
	ref := C.getSystemWideElement()
	if ref == nil {
		return nil
	}
	return &Element{ref: ref}
}

// GetFocusedApplication returns the currently focused application
func GetFocusedApplication() *Element {
	ref := C.getFocusedApplication()
	if ref == nil {
		return nil
	}
	return &Element{ref: ref}
}

// GetApplicationByPID returns an application element by its process ID
func GetApplicationByPID(pid int) *Element {
	ref := C.getApplicationByPID(C.int(pid))
	if ref == nil {
		return nil
	}
	return &Element{ref: ref}
}

// GetElementAtPosition returns the element at the specified screen position
func GetElementAtPosition(x, y int) *Element {
	pos := C.CGPoint{x: C.double(x), y: C.double(y)}
	ref := C.getElementAtPosition(pos)
	if ref == nil {
		return nil
	}
	return &Element{ref: ref}
}

// GetInfo returns information about the element
func (e *Element) GetInfo() (*ElementInfo, error) {
	if e.ref == nil {
		return nil, fmt.Errorf("element is nil")
	}

	cInfo := C.getElementInfo(e.ref)
	if cInfo == nil {
		return nil, fmt.Errorf("failed to get element info")
	}
	defer C.freeElementInfo(cInfo)

	info := &ElementInfo{
		Position: image.Point{
			X: int(cInfo.position.x),
			Y: int(cInfo.position.y),
		},
		Size: image.Point{
			X: int(cInfo.size.width),
			Y: int(cInfo.size.height),
		},
		IsEnabled: bool(cInfo.isEnabled),
		IsFocused: bool(cInfo.isFocused),
		PID:       int(cInfo.pid),
	}

	if cInfo.title != nil {
		info.Title = C.GoString(cInfo.title)
	}
	if cInfo.role != nil {
		info.Role = C.GoString(cInfo.role)
	}
	if cInfo.roleDescription != nil {
		info.RoleDescription = C.GoString(cInfo.roleDescription)
	}

	return info, nil
}

// GetChildren returns all child elements
func (e *Element) GetChildren() ([]*Element, error) {
	if e.ref == nil {
		return nil, fmt.Errorf("element is nil")
	}

	var count C.int
	children := C.getChildren(e.ref, &count)
	if children == nil || count == 0 {
		return []*Element{}, nil
	}
	defer C.free(unsafe.Pointer(children))

	// Convert C array to Go slice
	childSlice := (*[1 << 30]unsafe.Pointer)(unsafe.Pointer(children))[:count:count]
	result := make([]*Element, count)

	for i := 0; i < int(count); i++ {
		result[i] = &Element{ref: childSlice[i]}
	}

	return result, nil
}

// GetChildrenCount returns the number of child elements
func (e *Element) GetChildrenCount() int {
	if e.ref == nil {
		return 0
	}
	return int(C.getChildrenCount(e.ref))
}

// Click performs a click action on the element
func (e *Element) Click() error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performClick(e.ref)
	if result == 0 {
		return fmt.Errorf("click action failed")
	}
	return nil
}

// RightClick performs a right-click action on the element
func (e *Element) RightClick() error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performRightClick(e.ref)
	if result == 0 {
		return fmt.Errorf("right-click action failed")
	}
	return nil
}

// DoubleClick performs a double-click action on the element
func (e *Element) DoubleClick() error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performDoubleClick(e.ref)
	if result == 0 {
		return fmt.Errorf("double-click action failed")
	}
	return nil
}

// SetFocus sets focus to the element
func (e *Element) SetFocus() error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.setFocus(e.ref)
	if result == 0 {
		return fmt.Errorf("set focus failed")
	}
	return nil
}

// GetAttribute gets a custom attribute value
func (e *Element) GetAttribute(name string) (string, error) {
	if e.ref == nil {
		return "", fmt.Errorf("element is nil")
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cValue := C.getElementAttribute(e.ref, cName)
	if cValue == nil {
		return "", fmt.Errorf("attribute not found")
	}
	defer C.freeString(cValue)

	return C.GoString(cValue), nil
}

// Release releases the element reference
func (e *Element) Release() {
	if e.ref != nil {
		C.releaseElement(e.ref)
		e.ref = nil
	}
}

// IsClickable checks if the element is clickable
func (e *Element) IsClickable() bool {
	info, err := e.GetInfo()
	if err != nil {
		return false
	}

	if !info.IsEnabled {
		return false
	}

	// Check if role is clickable
	clickableRoles := map[string]bool{
		"AXButton":          true,
		"AXCheckBox":        true,
		"AXRadioButton":     true,
		"AXPopUpButton":     true,
		"AXMenuItem":        true,
		"AXLink":            true,
		"AXTextField":       true,
		"AXTextArea":        true,
		"AXStaticText":      false,
		"AXImage":           false,
	}

	return clickableRoles[info.Role]
}

// GetAllWindows returns all windows of the focused application
func GetAllWindows() ([]*Element, error) {
	var count C.int
	windows := C.getAllWindows(&count)
	if windows == nil || count == 0 {
		return []*Element{}, nil
	}
	defer C.free(unsafe.Pointer(windows))

	windowSlice := (*[1 << 30]unsafe.Pointer)(unsafe.Pointer(windows))[:count:count]
	result := make([]*Element, count)

	for i := 0; i < int(count); i++ {
		result[i] = &Element{ref: windowSlice[i]}
	}

	return result, nil
}

// GetFrontmostWindow returns the frontmost window
func GetFrontmostWindow() *Element {
	ref := C.getFrontmostWindow()
	if ref == nil {
		return nil
	}
	return &Element{ref: ref}
}

// GetApplicationName returns the application name
func (e *Element) GetApplicationName() string {
	if e.ref == nil {
		return ""
	}

	cName := C.getApplicationName(e.ref)
	if cName == nil {
		return ""
	}
	defer C.freeString(cName)

	return C.GoString(cName)
}

// GetBundleIdentifier returns the bundle identifier
func (e *Element) GetBundleIdentifier() string {
	if e.ref == nil {
		return ""
	}

	cBundleID := C.getBundleIdentifier(e.ref)
	if cBundleID == nil {
		return ""
	}
	defer C.freeString(cBundleID)

	return C.GoString(cBundleID)
}

// IsScrollable checks if the element is scrollable
func (e *Element) IsScrollable() bool {
	if e.ref == nil {
		return false
	}

	result := C.isScrollable(e.ref)
	return result == 1
}

// GetScrollBounds returns the scroll area bounds
func (e *Element) GetScrollBounds() image.Rectangle {
	if e.ref == nil {
		return image.Rectangle{}
	}

	rect := C.getScrollBounds(e.ref)
	return image.Rectangle{
		Min: image.Point{
			X: int(rect.origin.x),
			Y: int(rect.origin.y),
		},
		Max: image.Point{
			X: int(rect.origin.x + rect.size.width),
			Y: int(rect.origin.y + rect.size.height),
		},
	}
}

// Scroll scrolls the element by the specified delta
func (e *Element) Scroll(deltaX, deltaY int) error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.scrollElement(e.ref, C.int(deltaX), C.int(deltaY))
	if result == 0 {
		return fmt.Errorf("scroll failed")
	}
	return nil
}
