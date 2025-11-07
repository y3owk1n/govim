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
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

// Element represents an accessibility UI element
type Element struct {
	ref unsafe.Pointer
}

var (
	clickableRoles   = make(map[string]struct{})
	clickableRolesMu sync.RWMutex

	scrollableRoles   = make(map[string]struct{})
	scrollableRolesMu sync.RWMutex
)

// SetClickableRoles configures which accessibility roles are treated as clickable.
func SetClickableRoles(roles []string) {
	clickableRolesMu.Lock()
	defer clickableRolesMu.Unlock()

	clickableRoles = make(map[string]struct{}, len(roles))
	for _, role := range roles {
		trimmed := strings.TrimSpace(role)
		if trimmed == "" {
			continue
		}
		clickableRoles[trimmed] = struct{}{}
	}

	logger.Debug("Updated clickable roles",
		zap.Int("count", len(clickableRoles)),
		zap.Strings("roles", roles))
}

// SetScrollableRoles configures which accessibility roles are treated as scrollable.
func SetScrollableRoles(roles []string) {
	scrollableRolesMu.Lock()
	defer scrollableRolesMu.Unlock()

	scrollableRoles = make(map[string]struct{}, len(roles))
	for _, role := range roles {
		trimmed := strings.TrimSpace(role)
		if trimmed == "" {
			continue
		}
		scrollableRoles[trimmed] = struct{}{}
	}

	logger.Debug("Updated scrollable roles",
		zap.Int("count", len(scrollableRoles)),
		zap.Strings("roles", roles))
}

// GetClickableRoles returns the configured clickable roles.
func GetClickableRoles() []string {
	clickableRolesMu.RLock()
	defer clickableRolesMu.RUnlock()

	roles := make([]string, 0, len(clickableRoles))
	for role := range clickableRoles {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}

// GetScrollableRoles returns the configured scrollable roles.
func GetScrollableRoles() []string {
	scrollableRolesMu.RLock()
	defer scrollableRolesMu.RUnlock()

	roles := make([]string, 0, len(scrollableRoles))
	for role := range scrollableRoles {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
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

// GetApplicationByBundleID returns an application element by bundle identifier
func GetApplicationByBundleID(bundleID string) *Element {
	cBundle := C.CString(bundleID)
	defer C.free(unsafe.Pointer(cBundle))

	ref := C.getApplicationByBundleId(cBundle)
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

// GetChildren returns all child elements with optional occlusion checking
func (e *Element) GetChildren() ([]*Element, error) {
	if e.ref == nil {
		return nil, fmt.Errorf("element is nil")
	}

	var count C.int
	var rawChildren unsafe.Pointer

	info := globalCache.Get(e)
	if info == nil {
		var err error
		info, err = e.GetInfo()
		if err != nil {
			return nil, nil
		}
		globalCache.Set(e, info)
	}

	if info != nil {
		switch info.Role {
		case "AXList", "AXTable", "AXOutline":
			ptr := unsafe.Pointer(C.getVisibleRows(e.ref, &count))
			if ptr != nil {
				rawChildren = ptr
			} else {
				rawChildren = unsafe.Pointer(C.getChildren(e.ref, &count))
			}
		default:
			rawChildren = unsafe.Pointer(C.getChildren(e.ref, &count))
		}
	}

	if rawChildren == nil || count == 0 {
		return nil, nil
	}
	defer C.free(rawChildren)

	childSlice := (*[1 << 30]unsafe.Pointer)(rawChildren)[:count:count]
	children := make([]*Element, count)
	for i := range children {
		children[i] = &Element{ref: childSlice[i]}
	}

	return children, nil
}

// LeftClick performs a click action on the element
func (e *Element) LeftClick(restoreCursor bool) error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performLeftClick(e.ref, C.bool(restoreCursor))
	if result == 1 {
		return nil
	}

	return fmt.Errorf("left-click action failed")
}

// RightClick performs a right-click action on the element
func (e *Element) RightClick(restoreCursor bool) error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performRightClick(e.ref, C.bool(restoreCursor))
	if result == 0 {
		return fmt.Errorf("right-click action failed")
	}
	return nil
}

// DoubleClick performs a double-click action on the element
func (e *Element) DoubleClick(restoreCursor bool) error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performDoubleClick(e.ref, C.bool(restoreCursor))
	if result == 0 {
		return fmt.Errorf("double-click action failed")
	}
	return nil
}

// MiddleClick performs a middle-click action on the element
func (e *Element) MiddleClick(restoreCursor bool) error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performMiddleClick(e.ref, C.bool(restoreCursor))
	if result == 0 {
		return fmt.Errorf("middle-click action failed")
	}
	return nil
}

// GoToPosition performs a move mouse action to the element
func (e *Element) GoToPosition() error {
	if e.ref == nil {
		return fmt.Errorf("element is nil")
	}

	result := C.performMoveMouseToPosition(e.ref)
	if result == 1 {
		return nil
	}

	return fmt.Errorf("left-click action failed")
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

// ReleaseAll releases all elements in a slice
func ReleaseAll(elements []*Element) {
	for _, elem := range elements {
		if elem != nil {
			elem.Release()
		}
	}
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

// GetMenuBar returns the menu bar element for the given application element
func (e *Element) GetMenuBar() *Element {
	if e.ref == nil {
		return nil
	}
	ref := C.getMenuBar(e.ref)
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
		return fmt.Errorf("failed to scroll element")
	}
	return nil
}

// IsClickable checks if the element is clickable
func (e *Element) IsClickable() bool {
	if e.ref == nil {
		return false
	}

	info := globalCache.Get(e)
	if info == nil {
		var err error
		info, err = e.GetInfo()
		if err != nil {
			return false
		}
		globalCache.Set(e, info)
	}

	// First check if the role is in the clickable roles list
	clickableRolesMu.RLock()
	_, ok := clickableRoles[info.Role]
	clickableRolesMu.RUnlock()

	if ok {
		// Also verify it actually has scroll capability
		result := C.hasClickAction(e.ref)
		return result == 1
	}

	return false
}

// IsScrollable checks if the element is scrollable
func (e *Element) IsScrollable() bool {
	if e.ref == nil {
		return false
	}

	info := globalCache.Get(e)
	if info == nil {
		var err error
		info, err = e.GetInfo()
		if err != nil {
			return false
		}
		globalCache.Set(e, info)
	}

	// Check if the role is in the scrollable roles list
	scrollableRolesMu.RLock()
	_, ok := scrollableRoles[info.Role]
	scrollableRolesMu.RUnlock()

	if ok {
		// Also verify it actually has scroll capability
		result := C.isScrollable(e.ref)
		return result == 1
	}

	return false
}
