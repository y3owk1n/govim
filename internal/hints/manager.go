package hints

import (
	"strings"
)

// Manager handles common hint input and filtering logic
type Manager struct {
	currentInput string
	currentHints *HintCollection
	onHintUpdate func([]*Hint)
}

// NewManager creates a new hint manager
func NewManager(onHintUpdate func([]*Hint)) *Manager {
	return &Manager{
		onHintUpdate: onHintUpdate,
	}
}

// SetHints sets the current hint collection
func (m *Manager) SetHints(hints *HintCollection) {
	m.currentHints = hints
	m.currentInput = ""
	m.updateHints()
}

// Reset resets the current input and redraws all hints
func (m *Manager) Reset() {
	m.currentInput = ""
	m.updateHints()
}

// HandleInput processes an input character and returns true if an exact match was found
func (m *Manager) HandleInput(key string) (exactMatch *Hint, ok bool) {
	if m.currentHints == nil {
		return nil, false
	}

	// Handle backspace
	if key == "\x7f" || key == "delete" || key == "backspace" {
		if len(m.currentInput) > 0 {
			m.currentInput = m.currentInput[:len(m.currentInput)-1]
			m.updateHints()
		} else {
			m.Reset()
		}
		return nil, false
	}

	// Ignore non-letter keys
	if len(key) != 1 || !isLetter(key[0]) {
		return nil, false
	}

	// Accumulate input (convert to uppercase to match hints)
	m.currentInput += strings.ToUpper(key)

	// Filter and update hints
	filtered := m.currentHints.FilterByPrefix(m.currentInput)

	if len(filtered) == 0 {
		// No matches - reset
		m.currentInput = ""
		return nil, false
	}

	// Update matched prefix for filtered hints
	for _, hint := range filtered {
		hint.MatchedPrefix = m.currentInput
	}

	// Notify of hint updates
	m.onHintUpdate(filtered)

	// Check for exact match
	if len(filtered) == 1 && filtered[0].Label == m.currentInput {
		return filtered[0], true
	}

	return nil, false
}

// GetInput returns the current input string
func (m *Manager) GetInput() string {
	return m.currentInput
}

func (m *Manager) updateHints() {
	var filtered []*Hint
	if m.currentInput == "" {
		filtered = m.currentHints.GetHints()
	} else {
		filtered = m.currentHints.FilterByPrefix(m.currentInput)
	}

	// Update matched prefix for filtered hints
	for _, hint := range filtered {
		hint.MatchedPrefix = m.currentInput
	}

	// Notify of hint updates
	m.onHintUpdate(filtered)
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
