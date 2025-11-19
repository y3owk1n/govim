package accessibility

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	infra "github.com/y3owk1n/neru/internal/infra/accessibility"
	"go.uber.org/zap"
)

// Service handles high-level accessibility operations.
type Service struct {
	config *config.Config
	logger *zap.Logger
}

// NewService creates a new accessibility service.
func NewService(cfg *config.Config, log *zap.Logger) *Service {
	return &Service{
		config: cfg,
		logger: log,
	}
}

// UpdateRolesForCurrentApp updates the accessibility roles based on the currently focused application.
func (s *Service) UpdateRolesForCurrentApp() {
	focusedApp := infra.GetFocusedApplication()
	if focusedApp == nil {
		s.logger.Debug("No focused application, using global roles only")
		infra.SetClickableRoles(s.config.Hints.ClickableRoles)
		return
	}
	defer focusedApp.Release()

	bundleID := focusedApp.GetBundleIdentifier()
	if bundleID == "" {
		s.logger.Debug("Could not get bundle ID, using global roles only")
		infra.SetClickableRoles(s.config.Hints.ClickableRoles)
		return
	}

	clickableRoles := s.config.GetClickableRolesForApp(bundleID)

	s.logger.Debug("Updating roles for current app",
		zap.String("bundle_id", bundleID),
		zap.Int("clickable_count", len(clickableRoles)),
	)

	infra.SetClickableRoles(clickableRoles)
}

// GetFocusedBundleID retrieves the bundle identifier of the currently focused application.
// Returns an empty string if the bundle ID cannot be determined.
func (s *Service) GetFocusedBundleID() string {
	app := infra.GetFocusedApplication()
	if app == nil {
		return ""
	}
	defer app.Release()
	return app.GetBundleIdentifier()
}

// IsFocusedAppExcluded checks if the currently focused application is in the excluded apps list.
// Returns true if the app should be excluded from Neru functionality.
func (s *Service) IsFocusedAppExcluded() bool {
	bundleID := s.GetFocusedBundleID()
	if bundleID != "" && s.config.IsAppExcluded(bundleID) {
		s.logger.Debug("Current app is excluded; ignoring mode activation",
			zap.String("bundle_id", bundleID))
		return true
	}
	return false
}

// CollectElements collects UI elements based on the current mode.
func (s *Service) CollectElements() []*infra.TreeNode {
	var elements []*infra.TreeNode

	// Check if Mission Control is active - affects what we can scan
	missionControlActive := infra.IsMissionControlActive()

	elements = s.collectClickableElements(missionControlActive)

	elements = s.addSupplementaryElements(elements, missionControlActive)

	return elements
}

// PerformActionAtPoint executes the specified action at the given point.
func (s *Service) PerformActionAtPoint(action string, pt image.Point) error {
	actionName := domain.ActionName(action)
	switch actionName {
	case domain.ActionNameLeftClick:
		return infra.LeftClickAtPoint(pt, false)
	case domain.ActionNameRightClick:
		return infra.RightClickAtPoint(pt, false)
	case domain.ActionNameMiddleClick:
		return infra.MiddleClickAtPoint(pt, false)
	case domain.ActionNameMouseDown:
		return infra.LeftMouseDownAtPoint(pt)
	case domain.ActionNameMouseUp:
		return infra.LeftMouseUpAtPoint(pt)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// collectClickableElements collects clickable elements from the frontmost window.
func (s *Service) collectClickableElements(missionControlActive bool) []*infra.TreeNode {
	if missionControlActive {
		s.logger.Info("Mission Control is active, skipping frontmost window clickable elements")
		return nil
	}

	s.logger.Info("Scanning for clickable elements")
	roles := infra.GetClickableRoles()
	s.logger.Debug("Clickable roles", zap.Strings("roles", roles))

	clickableElements, err := infra.GetClickableElements()
	if err != nil {
		s.logger.Error("Failed to get clickable elements", zap.Error(err))
		return nil
	}

	s.logger.Info("Found clickable elements", zap.Int("count", len(clickableElements)))
	return clickableElements
}

// addSupplementaryElements adds menubar, dock, and notification center elements.
func (s *Service) addSupplementaryElements(
	elements []*infra.TreeNode,
	missionControlActive bool,
) []*infra.TreeNode {
	if !missionControlActive {
		elements = s.addMenubarElements(elements)
	} else {
		s.logger.Info("Mission Control is active, skipping menubar elements")
	}

	elements = s.addDockElements(elements)

	// Notification Center elements (only when Mission Control is active)
	if missionControlActive {
		elements = s.addNotificationCenterElements(elements)
	}

	return elements
}

// addMenubarElements adds menubar clickable elements.
func (s *Service) addMenubarElements(elements []*infra.TreeNode) []*infra.TreeNode {
	if !s.config.Hints.IncludeMenubarHints {
		return elements
	}

	s.logger.Info("Adding menubar elements")

	var mbElems []*infra.TreeNode
	var err error
	mbElems, err = infra.GetMenuBarClickableElements()
	if err == nil {
		elements = append(elements, mbElems...)
		s.logger.Debug("Included menubar elements", zap.Int("count", len(mbElems)))
	} else {
		s.logger.Warn("Failed to get menubar elements", zap.Error(err))
	}
	for _, bundleID := range s.config.Hints.AdditionalMenubarHintsTargets {
		var additionalElems []*infra.TreeNode
		var err error
		additionalElems, err = infra.GetClickableElementsFromBundleID(bundleID)
		if err == nil {
			elements = append(elements, additionalElems...)
			s.logger.Debug("Included additional menubar elements",
				zap.String("bundle_id", bundleID),
				zap.Int("count", len(additionalElems)))
		} else {
			s.logger.Warn("Failed to get additional menubar elements",
				zap.String("bundle_id", bundleID),
				zap.Error(err))
		}
	}

	return elements
}

// addDockElements adds dock clickable elements.
func (s *Service) addDockElements(elements []*infra.TreeNode) []*infra.TreeNode {
	if !s.config.Hints.IncludeDockHints {
		return elements
	}

	var dockElems []*infra.TreeNode
	var err error
	dockElems, err = infra.GetClickableElementsFromBundleID(domain.BundleIDDock)
	if err == nil {
		elements = append(elements, dockElems...)
		s.logger.Debug("Included dock elements", zap.Int("count", len(dockElems)))
	} else {
		s.logger.Warn("Failed to get dock elements", zap.Error(err))
	}

	return elements
}

// addNotificationCenterElements adds notification center clickable elements.
func (s *Service) addNotificationCenterElements(
	elements []*infra.TreeNode,
) []*infra.TreeNode {
	if !s.config.Hints.IncludeNCHints {
		return elements
	}

	s.logger.Info("Adding notification center elements")

	var ncElems []*infra.TreeNode
	var err error
	ncElems, err = infra.GetClickableElementsFromBundleID(domain.BundleIDNotificationCenter)
	if err == nil {
		elements = append(elements, ncElems...)
		s.logger.Debug("Included notification center elements", zap.Int("count", len(ncElems)))
	} else {
		s.logger.Warn("Failed to get notification center elements", zap.Error(err))
	}

	return elements
}
