package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/password"
	"github.com/escribanoruben/tale/internal/profile"
)

// AppSettings are the user-visible generation options shown in the TUI.
type AppSettings struct {
	Language             string
	Phrases              int
	AllowAdjacentRepeats bool
}

// AppConfig configures the full interactive flow.
type AppConfig struct {
	Version              string
	Language             string
	Phrases              int
	Length               int
	Password             string
	AllowAdjacentRepeats bool
	ForceSetup           bool
	SaveSettings         func(AppSettings) error
	LoadProfile          func(language string) (profile.Profile, *dictionary.Dictionary, error)
}

type appPhase int

const (
	appPhaseSetup appPhase = iota
	appPhasePicker
)

// AppModel orchestrates setup and the word picker.
type AppModel struct {
	phase     appPhase
	cfg       AppConfig
	settings  AppSettings
	languages []string
	setup     setupModel
	picker    Model
	setupHits *hitTargets
	err       error
}

// RunApp starts setup (if needed) and the mnemonic picker.
func RunApp(cfg AppConfig) error {
	appVersion = cfg.Version
	languages, err := profile.AvailableLanguages()
	if err != nil {
		return err
	}
	if len(languages) == 0 {
		return fmt.Errorf("no language profiles available")
	}

	settings := AppSettings{
		Language:             cfg.Language,
		Phrases:              cfg.Phrases,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
	}
	if settings.Phrases <= 0 {
		settings.Phrases = 5
	}
	settings.Phrases = snapPhraseCount(settings.Phrases)

	m := AppModel{
		cfg:       cfg,
		settings:  settings,
		languages: languages,
		setup:     newSetupModel(languages, settings),
		setupHits: &hitTargets{},
	}
	if cfg.ForceSetup {
		m.phase = appPhaseSetup
	} else {
		session, err := m.newSession()
		if err != nil {
			return err
		}
		m.picker = NewModel(session, settings, languages)
		m.phase = appPhasePicker
	}

	final, err := tea.NewProgram(m, mouseProgramOpts()...).Run()
	if err != nil {
		return err
	}
	app := final.(AppModel)
	if app.err != nil {
		return app.err
	}
	if app.picker.session != nil && app.picker.session.Quitting && !app.picker.session.Done {
		return fmt.Errorf("cancelled")
	}
	if app.picker.session != nil && app.picker.session.Done {
		fmt.Print(app.picker.View())
	}
	return nil
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case appPhaseSetup:
		return m.updateSetup(msg)
	default:
		return m.updatePicker(msg)
	}
}

func (m AppModel) View() string {
	switch m.phase {
	case appPhaseSetup:
		return m.setup.View(m.setupHits)
	default:
		return m.picker.View()
	}
}

func (m AppModel) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		return m.handleSetupMouse(mouseMsg)
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if keyMsg.String() == "esc" {
		return m, tea.Quit
	}

	langBefore := m.setup.ctrl.draft.Language
	m.setup.update(keyMsg)
	if m.setup.ctrl.draft.Language != langBefore {
		m.settings.Language = m.setup.ctrl.draft.Language
	}

	switch keyMsg.String() {
	case "g", "ctrl+enter":
		return m.startPickerFromSetup()
	}
	return m, nil
}

func (m AppModel) startPickerFromSetup() (tea.Model, tea.Cmd) {
	m.settings = m.setup.toSettings()
	if err := m.saveSettings(); err != nil {
		m.err = err
		return m, tea.Quit
	}
	session, err := m.newSession()
	if err != nil {
		m.err = err
		return m, tea.Quit
	}
	m.picker = NewModel(session, m.settings, m.languages)
	m.phase = appPhasePicker
	return m, nil
}

func (m AppModel) handleSetupMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if d := wheelDelta(msg); d != 0 {
		key := "j"
		if d < 0 {
			key = "k"
		}
		m.setup.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		return m, nil
	}
	if !isLeftClick(msg) {
		return m, nil
	}
	zone, ok := m.setupHits.at(msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	switch zone.kind {
	case hitSettingsField:
		m.setup.ctrl.jumpToSection(zone.index, m.languages)
	case hitSettingsPick:
		m.setup.ctrl.selectOption(zone.extra, zone.index, m.languages)
		m.settings.Language = m.setup.ctrl.draft.Language
	}
	return m, nil
}

func (m AppModel) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mouseMsg, ok := msg.(tea.MouseMsg); ok {
		if m.picker.settingsFocus {
			return m.handleSettingsMouse(mouseMsg)
		}
		if isLeftClick(mouseMsg) {
			zone, ok := m.picker.hits.at(mouseMsg.X, mouseMsg.Y)
			if ok && (zone.kind == hitSettingsPanel || zone.kind == hitSettingsField) {
				field := 0
				if zone.kind == hitSettingsField {
					field = zone.index
				}
				m.focusSettings(field)
				return m, nil
			}
		}
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.picker.settingsFocus {
			return m.handleSettingsKeys(keyMsg)
		}
		if isSettingsFocusKey(keyMsg) {
			m.focusSettings(0)
			return m, nil
		}
		if field, ok := settingsFieldForKey(keyMsg); ok {
			m.focusSettings(field)
			return m, nil
		}
	}

	updated, cmd := m.picker.Update(msg)
	m.picker = updated.(Model)
	if m.picker.session != nil && (m.picker.session.Done || m.picker.session.Quitting) {
		if m.picker.session.Done {
			_ = m.saveSettings()
		}
		return m, cmd
	}
	return m, cmd
}

func (m AppModel) handleSettingsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if d := wheelDelta(msg); d != 0 {
		key := "j"
		if d < 0 {
			key = "k"
		}
		return m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	}
	if !isLeftClick(msg) {
		return m, nil
	}
	zone, ok := m.picker.hits.at(msg.X, msg.Y)
	if !ok {
		return m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEsc})
	}
	switch zone.kind {
	case hitSettingsField:
		m.picker.settingsCtrl.jumpToSection(zone.index, m.languages)
	case hitSettingsPick:
		m.picker.settingsCtrl.selectOption(zone.extra, zone.index, m.languages)
	case hitSettingsConfirm:
		m.picker.settingsCtrl.focusActionRow(zone.index, m.languages, m.settings)
		return m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeySpace})
	}
	return m, nil
}

func (m AppModel) handleSettingsKeys(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.picker.settingsConfirmDiscard {
		return m.handleSettingsDiscardConfirm(keyMsg)
	}
	exit := handleSettingsKey(&m.picker.settingsCtrl, keyMsg, m.languages, m.settings)
	switch exit {
	case settingsExitQuitApp:
		m.picker.session.Quitting = true
		return m, tea.Quit
	case settingsExitConfirmDiscard:
		m.picker.settingsConfirmDiscard = true
	case settingsExitUnfocus:
		return m.discardSettingsAndUnfocus()
	case settingsExitApply:
		draft := m.picker.settingsCtrl.draft
		draft.Phrases = snapPhraseCount(draft.Phrases)
		if settingsChangeKind(m.settings, draft) != settingsChangeNone {
			m.settings = draft
			if err := m.applySettingsChanges(); err != nil {
				m.picker.session.Err = err
				return m, nil
			}
		}
		m.picker.settingsFocus = false
		m.picker.settingsConfirmDiscard = false
		m.picker.settingsCtrl = newSettingsController(m.settings)
		m.picker.session.Err = nil
	}
	return m, nil
}

func (m AppModel) handleSettingsDiscardConfirm(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.String() {
	case "y", "Y", "s", "S":
		return m.discardSettingsAndUnfocus()
	case "n", "N", "esc":
		m.picker.settingsConfirmDiscard = false
	}
	return m, nil
}

func (m AppModel) discardSettingsAndUnfocus() (tea.Model, tea.Cmd) {
	m.picker.settingsConfirmDiscard = false
	m.picker.settingsCtrl = newSettingsController(m.settings)
	m.picker.settingsFocus = false
	m.picker.session.Err = nil
	return m, nil
}

func (m *AppModel) applySettingsChanges() error {
	m.settings.Phrases = snapPhraseCount(m.settings.Phrases)
	if err := m.saveSettings(); err != nil {
		return err
	}
	session, err := m.newSession()
	if err != nil {
		return err
	}
	m.picker.session = session
	m.picker.settings = m.settings
	m.picker.settingsCtrl = newSettingsController(m.settings)
	return nil
}

func (m AppModel) saveSettings() error {
	if m.cfg.SaveSettings == nil {
		return nil
	}
	return m.cfg.SaveSettings(m.settings)
}

func (m AppModel) newSession() (*Session, error) {
	if m.cfg.LoadProfile == nil {
		return nil, fmt.Errorf("LoadProfile not configured")
	}
	p, dict, err := m.cfg.LoadProfile(m.settings.Language)
	if err != nil {
		return nil, err
	}
	if m.cfg.Length > 0 {
		if err := applyPasswordLength(&p, m.cfg.Length); err != nil {
			return nil, err
		}
	} else if m.settings.Phrases > 0 {
		p.PasswordLength = m.settings.Phrases * p.BlockSize
	}
	pass := m.cfg.Password
	if pass == "" {
		pass, err = password.Generate(p.Alphabet, p.PasswordLength, password.Options{
			PreventAdjacentRepeats: !m.settings.AllowAdjacentRepeats,
		})
		if err != nil {
			return nil, err
		}
	}
	return NewSession(p, pass, dict, !m.settings.AllowAdjacentRepeats)
}

func applyPasswordLength(p *profile.Profile, length int) error {
	if length <= 0 {
		return fmt.Errorf("length must be positive")
	}
	if length%p.BlockSize != 0 {
		return fmt.Errorf("length %d is not divisible by block size %d", length, p.BlockSize)
	}
	p.PasswordLength = length
	return nil
}

func uiForLang(lang string) uiText {
	if lang == "es" {
		return uiFor(profile.Profile{Language: "es"})
	}
	return uiFor(profile.Profile{Language: "en"})
}

func isSettingsFocusKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyTab:
		return true
	default:
		return msg.String() == "tab"
	}
}

func settingsFieldForKey(msg tea.KeyMsg) (int, bool) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return 0, false
	}
	switch msg.Runes[0] {
	case '1':
		return settingFieldLanguage, true
	case '2':
		return settingFieldNoAdjacent, true
	case '3':
		return settingFieldPhrases, true
	default:
		return 0, false
	}
}

func (m *AppModel) focusSettings(field int) {
	m.picker.settingsFocus = true
	m.picker.settingsConfirmDiscard = false
	m.picker.mode = modeNormal
	m.picker.settingsCtrl = newSettingsController(m.settings)
	m.picker.settingsCtrl.syncFocusFromDraft(m.languages)
	m.picker.settingsCtrl.jumpToSection(field, m.languages)
	m.picker.session.Err = nil
}

func boolLabel(text uiText, label string, yes bool) string {
	if yes {
		return label + ": " + text.yes
	}
	return label + ": " + text.no
}
