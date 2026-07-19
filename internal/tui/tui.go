package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/normalize"
	"github.com/escribanoruben/tale/internal/password"
	"github.com/escribanoruben/tale/internal/phrase"
	"github.com/escribanoruben/tale/internal/profile"
)

const visibleRadius = 2

var (
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	appTitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	activeStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Padding(0, 1)
	passwordValueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("236")).Padding(0, 1)
	dimStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	warnStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	panelStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	passwordPanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(0, 1)
)

// Session holds password mnemonic construction state.
type Session struct {
	Profile                profile.Profile
	Password               string
	Blocks                 []string
	Selections             [][]string
	Used                   map[string]bool
	PreventAdjacentRepeats bool
	BlockIndex             int
	SlotIndex              int
	Candidates             []string
	RollerIndex            int
	Done                   bool
	Quitting               bool
	Err                    error
	dict                   *dictionary.Dictionary
}

// NewSession builds a TUI session for the given profile and password.
func NewSession(p profile.Profile, password string, dict *dictionary.Dictionary, preventAdjacentRepeats bool) (*Session, error) {
	blocks, err := splitBlocks(password, p.BlockSize)
	if err != nil {
		return nil, err
	}
	if len(p.PhraseTemplate) != p.BlockSize {
		return nil, fmt.Errorf("phrase template has %d slots, block size is %d", len(p.PhraseTemplate), p.BlockSize)
	}
	s := &Session{
		Profile:                p,
		Password:               password,
		Blocks:                 blocks,
		Selections:             make([][]string, len(blocks)),
		Used:                   make(map[string]bool),
		PreventAdjacentRepeats: preventAdjacentRepeats,
		dict:                   dict,
	}
	for i := range s.Selections {
		s.Selections[i] = make([]string, len(p.PhraseTemplate))
	}
	if err := s.randomFill(); err != nil {
		return nil, err
	}
	if err := s.refreshCandidates(); err != nil {
		return nil, err
	}
	s.syncRollerToSelection()
	return s, nil
}

func (s *Session) randomFill() error {
	results, err := phrase.RandomForPassword(s.Profile, s.Password, s.dict, phrase.Options{AllowRepeat: false})
	if err != nil {
		return err
	}
	for i, r := range results {
		copy(s.Selections[i], r.Words)
		for _, w := range r.Words {
			s.Used[normalize.Key(w)] = true
		}
	}
	return nil
}

func splitBlocks(password string, blockSize int) ([]string, error) {
	if blockSize <= 0 {
		return nil, fmt.Errorf("invalid block size")
	}
	if len(password)%blockSize != 0 {
		return nil, fmt.Errorf("password length %d not divisible by block size %d", len(password), blockSize)
	}
	blocks := make([]string, 0, len(password)/blockSize)
	for i := 0; i < len(password); i += blockSize {
		blocks = append(blocks, password[i:i+blockSize])
	}
	return blocks, nil
}

func (s *Session) usedExceptCurrent() map[string]bool {
	return s.usedExceptSlot(s.BlockIndex, s.SlotIndex)
}

func (s *Session) usedExceptSlot(blockIndex, slotIndex int) map[string]bool {
	out := make(map[string]bool)
	for bi, sel := range s.Selections {
		for si, w := range sel {
			if bi == blockIndex && si == slotIndex {
				continue
			}
			if w != "" {
				out[normalize.Key(w)] = true
			}
		}
	}
	return out
}

func (s *Session) refreshCandidates() error {
	posName := s.Profile.PhraseTemplate[s.SlotIndex]
	pos, err := dictionary.ParsePOS(posName)
	if err != nil {
		return err
	}
	letter := string(s.Blocks[s.BlockIndex][s.SlotIndex])
	words, err := s.dict.WordsFor(s.Profile.Language, letter, pos)
	if err != nil {
		return err
	}
	words = dictionary.FilterAvailable(words, s.usedExceptCurrent())
	s.Candidates = words
	if len(s.Candidates) == 0 {
		s.RollerIndex = 0
		return nil
	}
	if s.RollerIndex >= len(s.Candidates) {
		s.RollerIndex = 0
	}
	return nil
}

func (s *Session) currentPOSLabel() string {
	return s.Profile.PhraseTemplate[s.SlotIndex]
}

func (s *Session) requiredLetter() string {
	return string(s.Blocks[s.BlockIndex][s.SlotIndex])
}

func (s *Session) flatIndex() int {
	return s.BlockIndex*len(s.Profile.PhraseTemplate) + s.SlotIndex
}

func (s *Session) totalSlots() int {
	return len(s.Blocks) * len(s.Profile.PhraseTemplate)
}

func (s *Session) syncRollerToVisible() {
	if len(s.Candidates) == 0 {
		s.RollerIndex = 0
		return
	}
	if s.RollerIndex >= len(s.Candidates) {
		s.RollerIndex = len(s.Candidates) - 1
	}
	current := s.Selections[s.BlockIndex][s.SlotIndex]
	if current == "" {
		return
	}
	for i, w := range s.Candidates {
		if normalize.Key(w) == normalize.Key(current) {
			s.RollerIndex = i
			return
		}
	}
}

func (s *Session) syncRollerToSelection() {
	s.syncRollerToVisible()
}

func (s *Session) applySelection() {
	if len(s.Candidates) == 0 {
		return
	}
	word := s.Candidates[s.RollerIndex]
	old := s.Selections[s.BlockIndex][s.SlotIndex]
	if old != "" {
		delete(s.Used, normalize.Key(old))
	}
	s.Selections[s.BlockIndex][s.SlotIndex] = word
	s.Used[normalize.Key(word)] = true
}

func (s *Session) previewRoller(delta int) {
	if len(s.Candidates) == 0 {
		return
	}
	s.RollerIndex = (s.RollerIndex + delta + len(s.Candidates)) % len(s.Candidates)
}

// cycleWord moves to the next/previous candidate and applies it immediately.
func (s *Session) cycleWord(delta int) {
	s.previewRoller(delta)
	s.applySelection()
}

func (s *Session) setRollerOffset(offset int) {
	if len(s.Candidates) == 0 {
		return
	}
	s.RollerIndex = (s.RollerIndex + offset + len(s.Candidates)) % len(s.Candidates)
}

func (s *Session) moveSlotInBlock(delta int) error {
	n := len(s.Profile.PhraseTemplate)
	if n == 0 || len(s.Blocks) == 0 {
		return nil
	}
	if delta == 0 {
		return nil
	}

	flat := s.BlockIndex*n + s.SlotIndex
	total := len(s.Blocks) * n
	flat = (flat + delta%total + total) % total
	s.BlockIndex = flat / n
	s.SlotIndex = flat % n

	if err := s.refreshCandidates(); err != nil {
		return err
	}
	s.syncRollerToSelection()
	return nil
}

func (s *Session) moveBlock(delta int) error {
	if len(s.Blocks) == 0 {
		return nil
	}
	s.BlockIndex = (s.BlockIndex + delta + len(s.Blocks)) % len(s.Blocks)
	if err := s.refreshCandidates(); err != nil {
		return err
	}
	s.syncRollerToSelection()
	return nil
}

func (s *Session) jumpToBlock(blockIndex int) error {
	return s.jumpToSlot(blockIndex, s.SlotIndex)
}

func (s *Session) jumpToSlot(blockIndex, slotIndex int) error {
	if blockIndex < 0 || blockIndex >= len(s.Blocks) {
		return nil
	}
	n := len(s.Profile.PhraseTemplate)
	if slotIndex < 0 || slotIndex >= n {
		return nil
	}
	s.BlockIndex = blockIndex
	s.SlotIndex = slotIndex
	if err := s.refreshCandidates(); err != nil {
		return err
	}
	s.syncRollerToSelection()
	return nil
}

// Model adapts Session to bubbletea.
type Model struct {
	session        *Session
	settings       AppSettings
	settingsCtrl   settingsController
	settingsFocus         bool
	settingsConfirmDiscard bool
	languages      []string
	keys           keyMap
	mode           uiMode
	width          int
	click          clickTracker
	hits           *hitTargets
}

type keyMap struct {
	SlotPrev  key.Binding
	SlotNext  key.Binding
	BlockPrev key.Binding
	BlockNext key.Binding
	Enter     key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		SlotPrev: key.NewBinding(
			key.WithKeys("h", "left", "shift+tab"),
			key.WithHelp("h/←", "prev slot"),
		),
		SlotNext: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "next slot"),
		),
		BlockPrev: key.NewBinding(
			key.WithKeys("k", "up", "["),
			key.WithHelp("k/↑", "prev phrase"),
		),
		BlockNext: key.NewBinding(
			key.WithKeys("j", "down", "]"),
			key.WithHelp("j/↓", "next phrase"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show result"),
		),
	}
}

// NewModel creates the bubbletea model.
func NewModel(session *Session, settings AppSettings, languages []string) Model {
	return Model{
		session:      session,
		settings:     settings,
		settingsCtrl: newSettingsController(settings),
		languages:    languages,
		keys:         defaultKeyMap(),
		hits:         &hitTargets{},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == modeHelp {
			return m.handleHelp(msg)
		}
		if m.mode == modeRandomMenu {
			return m.handleRandomMenu(msg)
		}
		if m.mode == modeChangeLetter {
			return m.handleChangeLetter(msg)
		}
		if handled, quit := m.handlePhraseKeys(msg); handled {
			if quit {
				return m, tea.Quit
			}
			return m, nil
		}

		switch {
		case msg.Type == tea.KeyEsc:
			m.session.Quitting = true
			return m, tea.Quit
		case msg.String() == "?":
			m.mode = modeHelp
			m.session.Err = nil
		case msg.String() == "r":
			m.mode = modeRandomMenu
			m.session.Err = nil
		case msg.String() == "c":
			m.mode = modeChangeLetter
			m.session.Err = nil
		}
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeHelp {
		return m.handleHelpMouse(msg)
	}
	now := time.Now()
	if d := wheelDelta(msg); d != 0 {
		zone, ok := m.hits.at(msg.X, msg.Y)
		if ok {
			m.wheelCycleWord(zone, d)
		}
		return m, nil
	}
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if !isLeftClick(msg) {
		return m, nil
	}
	zone, ok := m.hits.at(msg.X, msg.Y)
	if !ok {
		m.click.ok = false
		return m, nil
	}
	double := m.click.note(zone, true, now)
	switch zone.kind {
	case hitSlot:
		if err := m.session.jumpToSlot(zone.index, zone.extra); err != nil {
			m.session.Err = err
		}
	case hitBlock:
		if err := m.session.jumpToBlock(zone.index); err != nil {
			m.session.Err = err
		}
	case hitRoller:
		m.session.setRollerOffset(zone.index)
		m.session.applySelection()
		_ = double
	case hitOverlay:
		if m.mode == modeRandomMenu {
			return m.clickRandomOverlay(zone.index)
		}
	}
	return m, nil
}

func (m *Model) wheelCycleWord(zone hitZone, delta int) {
	switch zone.kind {
	case hitSlot:
		if err := m.session.jumpToSlot(zone.index, zone.extra); err != nil {
			m.session.Err = err
			return
		}
	case hitBlock:
		if err := m.session.jumpToBlock(zone.index); err != nil {
			m.session.Err = err
			return
		}
	case hitRoller:
		m.session.cycleWord(delta)
		return
	default:
		return
	}
	m.session.cycleWord(delta)
}

func (m Model) clickRandomOverlay(index int) (tea.Model, tea.Cmd) {
	scopes := []string{"w", "p", "a"}
	if index >= 0 && index < len(scopes) {
		if err := m.session.randomizeScope(scopes[index]); err != nil {
			m.session.Err = err
		}
		m.mode = modeNormal
		return m, nil
	}
	m.mode = modeNormal
	m.session.Err = nil
	return m, nil
}

func (m *Model) handlePhraseKeys(msg tea.KeyMsg) (handled bool, quit bool) {
	switch {
	case key.Matches(msg, m.keys.Enter):
		m.session.Done = true
		return true, true
	case msg.String() == "n":
		m.session.cycleWord(1)
		return true, false
	case msg.String() == "p":
		m.session.cycleWord(-1)
		return true, false
	case key.Matches(msg, m.keys.SlotNext):
		if err := m.session.moveSlotInBlock(1); err != nil {
			m.session.Err = err
		}
		return true, false
	case key.Matches(msg, m.keys.SlotPrev):
		if err := m.session.moveSlotInBlock(-1); err != nil {
			m.session.Err = err
		}
		return true, false
	case key.Matches(msg, m.keys.BlockNext):
		if err := m.session.moveBlock(1); err != nil {
			m.session.Err = err
		}
		return true, false
	case key.Matches(msg, m.keys.BlockPrev):
		if err := m.session.moveBlock(-1); err != nil {
			m.session.Err = err
		}
		return true, false
	default:
		return false, false
	}
}

func (m Model) handleRandomMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "w", "f", "a":
		scope := msg.String()
		if scope == "f" {
			scope = "p"
		}
		if err := m.session.randomizeScope(scope); err != nil {
			m.session.Err = err
		}
		m.mode = modeNormal
	case "esc":
		m.mode = modeNormal
		m.session.Err = nil
	}
	return m, nil
}

func (m Model) handleHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEsc, key.Matches(msg, m.keys.Enter), msg.String() == "?":
		m.mode = modeNormal
		m.session.Err = nil
	}
	return m, nil
}

func (m Model) handleHelpMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && isLeftClick(msg) {
		m.mode = modeNormal
		m.session.Err = nil
	}
	return m, nil
}

func (m Model) handleChangeLetter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.mode = modeNormal
		m.session.Err = nil
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1:
		letter := msg.Runes[0]
		if letter >= 'A' && letter <= 'Z' {
			letter += 'a' - 'A'
		}
		if err := m.session.changeLetter(letter); err != nil {
			m.session.Err = err
		} else {
			m.session.Err = nil
		}
		m.mode = modeNormal
	}
	return m, nil
}

func (m Model) View() string {
	s := m.session
	text := uiFor(s.Profile)
	if s.Done {
		return renderSummary(s, text)
	}
	if s.Quitting {
		return text.cancelled + "\n"
	}

	if m.hits != nil {
		m.hits.reset()
	}

	var b strings.Builder

	if m.settingsFocus {
		view, hits := renderSettingsView(m.settings, m.languages, m.settingsCtrl, m.width)
		b.WriteString(view)
		if m.settingsConfirmDiscard {
			text := uiFor(s.Profile)
			innerW := viewContentWidth(m.width)
			b.WriteString(renderSettingsDiscardConfirm(text, innerW))
		}
		if m.hits != nil {
			m.hits.addZones(hits)
		}
	} else if m.mode == modeHelp {
		b.WriteString(renderHelpView(text, m.width))
	} else {
		hitBase := viewLineY(&b)
		b.WriteString(m.buildPickerBody(text, true, hitBase))

		b.WriteString("\n")
		bar, barHits := renderSettingsSummaryBar(m.settings)
		barY := viewLineY(&b)
		b.WriteString(bar)
		if m.hits != nil {
			m.hits.addOffset(barY, barHits)
		}
	}

	if s.Err != nil {
		b.WriteString("\n")
		b.WriteString(warnStyle.Render(s.Err.Error()))
	}
	return b.String()
}

func (m Model) buildPickerChrome(text uiText) string {
	s := m.session
	passPanel := renderPasswordPanel(text.password, renderPasswordHighlight(s.Password, s.flatIndex(), s.Profile.BlockSize))
	entropyLine := helpStyle.Render(fmt.Sprintf("%s: %s", text.entropy, s.entropyLabel()))

	var b strings.Builder
	b.WriteString(renderAppHeader(text))
	b.WriteString("\n")
	b.WriteString(joinPasswordRow(m.width, passPanel, entropyLine))
	return b.String()
}

func (m Model) buildPickerBody(text uiText, registerHits bool, hitBase int) string {
	s := m.session
	var b strings.Builder

	b.WriteString(m.buildPickerChrome(text))
	b.WriteString("\n")

	rowStart := hitBase + viewLineY(&b)
	rollerPanel, rollerHits := renderRollerPanel(s, text)
	progressWide, progHitsWide := renderProgress(s, rowStart)

	if fitsSideBySide(m.width, rollerPanel, progressWide) {
		combined, rollerWidth, gapWidth, rollerYOff, phrasesYOff := joinSideBySide(rollerPanel, progressWide)
		b.WriteString(combined)
		if registerHits && m.hits != nil {
			m.hits.addZones(offsetHitZonesColumn(rollerHits, rowStart+rollerYOff, 0, rollerWidth))
			phrasesWidth := lipgloss.Width(progressWide)
			m.hits.addZones(offsetHitZonesColumn(progHitsWide, rowStart+phrasesYOff, rollerWidth+gapWidth, phrasesWidth))
		}
	} else {
		b.WriteString(rollerPanel)
		b.WriteString("\n\n")
		phrasesStart := hitBase + viewLineY(&b)
		progress, progHits := renderProgress(s, phrasesStart)
		b.WriteString(progress)
		if registerHits && m.hits != nil {
			m.hits.addOffset(rowStart, rollerHits)
			m.hits.addZones(progHits)
		}
	}

	m.appendPickerFooter(&b, text, registerHits, hitBase)
	return b.String()
}

func (m Model) appendPickerFooter(b *strings.Builder, text uiText, registerHits bool, hitBase int) {
	s := m.session

	if len(s.Candidates) == 0 {
		b.WriteString("\n")
		b.WriteString(warnStyle.Render(text.noWordsHint))
	}

	if m.mode == modeRandomMenu {
		b.WriteString("\n\n")
		overlayStart := hitBase + viewLineY(b)
		overlay := renderOverlay(text.randomTitle, []string{text.randomWord, text.randomPhrase, text.randomAll, text.randomCancel})
		b.WriteString(overlay)
		if registerHits && m.hits != nil {
			for i := 0; i < 4; i++ {
				row := overlayStart + 1 + 1 + i
				m.hits.add(row, row, hitOverlay, i)
			}
		}
	}
	if m.mode == modeChangeLetter {
		b.WriteString("\n\n")
		b.WriteString(renderOverlay(text.changeLetter, []string{text.changeCancel}))
	}
}

func (s *Session) entropyLabel() string {
	text := uiFor(s.Profile)
	opts := password.Options{PreventAdjacentRepeats: s.PreventAdjacentRepeats}
	bits := password.EntropyBits(len(s.Profile.Alphabet), len(s.Password), opts)
	if opts.PreventAdjacentRepeats && len(s.Password) > 1 && len(s.Profile.Alphabet) >= 2 {
		return fmt.Sprintf("~%.1f %s", bits, text.entropyBitsUnit)
	}
	return fmt.Sprintf(text.entropyFormula, bits, len(s.Password), len(s.Profile.Alphabet))
}

func renderRoller(words []string, index, blockIdx int) string {
	p := blockPaletteFor(blockIdx)
	chev := blockTextStyle(p, false).Render
	var lines []string
	for offset := -visibleRadius; offset <= visibleRadius; offset++ {
		pos := (index + offset + len(words)) % len(words)
		word := words[pos]
		switch {
		case offset == -2, offset == 2:
			lines = append(lines, "     "+dimStyle.Render(fmt.Sprintf("%-18s", word)))
		case offset == -1, offset == 1:
			lines = append(lines, "     "+labelStyle.Render(fmt.Sprintf("%-18s", word)))
		case offset == 0:
			lines = append(lines, "  "+chev("▸")+" "+blockSlotStyle(p).Render(word)+" "+chev("◂"))
		}
	}
	return strings.Join(lines, "\n")
}

func renderPasswordPanel(label, value string) string {
	return passwordPanelStyle.Render(labelStyle.Render(label+":") + " " + value)
}

func renderPasswordHighlight(password string, cursor, blockSize int) string {
	if blockSize <= 0 {
		blockSize = 4
	}
	var parts []string
	for i, r := range password {
		ch := string(r)
		blockIdx := i / blockSize
		p := blockPaletteFor(blockIdx)
		phraseActive := blockIdx == cursor/blockSize
		switch {
		case i == cursor:
			p := blockPaletteFor(blockIdx)
			parts = append(parts, blockCursorStyle(p).Render(ch))
		case phraseActive:
			parts = append(parts, blockTextStyle(p, true).Render(ch))
		default:
			parts = append(parts, blockTextStyle(p, false).Render(ch))
		}
	}
	return strings.Join(parts, "")
}

func renderPasswordColored(password string, blockSize int) string {
	if blockSize <= 0 {
		blockSize = 4
	}
	var parts []string
	for i, r := range password {
		ch := string(r)
		p := blockPaletteFor(i / blockSize)
		parts = append(parts, blockTextStyle(p, true).Render(ch))
	}
	return strings.Join(parts, "")
}

func renderOverlay(title string, lines []string) string {
	body := title
	if len(lines) > 0 {
		body += "\n" + strings.Join(lines, "\n")
	}
	return panelStyle.Render(body)
}

func renderProgress(s *Session, startY int) (string, []hitZone) {
	var parts []string
	var hits []hitZone
	for bi, block := range s.Blocks {
		line, lineHits := renderBlockLine(s, bi, block, bi == s.BlockIndex, startY+bi)
		parts = append(parts, line)
		hits = append(hits, lineHits...)
	}
	return strings.Join(parts, "\n"), hits
}

func renderBlockLine(s *Session, bi int, block string, active bool, y int) (string, []hitZone) {
	p := blockPaletteFor(bi)
	style := blockTextStyle(p, active)
	marker := "  "
	if active {
		marker = blockTextStyle(p, true).Render("▸") + " "
	}
	prefix := marker + style.Render(block+" → ")
	x := lipgloss.Width(prefix)
	words := make([]string, len(s.Selections[bi]))
	var hits []hitZone
	hits = append(hits, hitZone{
		y0: y, y1: y,
		x0: 0, x1: lipgloss.Width(prefix) - 1,
		kind: hitBlock, index: bi,
	})
	for si, word := range s.Selections[bi] {
		if active && si == s.SlotIndex {
			words[si] = blockSlotStyle(p).Render(word)
		} else {
			words[si] = style.Render(word)
		}
		w := lipgloss.Width(words[si])
		if w > 0 {
			hits = append(hits, hitZone{
				y0: y, y1: y,
				x0: x, x1: x + w - 1,
				kind: hitSlot, index: bi, extra: si,
			})
		}
		x += w
		if si < len(s.Selections[bi])-1 {
			x++ // separator space
		}
	}
	return prefix + strings.Join(words, " "), hits
}

func renderSummary(s *Session, text uiText) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(text.summaryTitle))
	b.WriteString("\n\n")
	b.WriteString(renderPasswordPanel(text.password, renderPasswordColored(s.Password, s.Profile.BlockSize)))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s: %s\n\n", text.entropy, s.entropyLabel()))
	for i, block := range s.Blocks {
		phrase := strings.Join(s.Selections[i], " ")
		p := blockPaletteFor(i)
		line := blockTextStyle(p, true).Render(fmt.Sprintf("[%s] %s", block, phrase))
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(text.recovery))
	b.WriteString("\n")
	return b.String()
}

// Run starts the interactive picker (used by tests).
func Run(session *Session) error {
	settings := AppSettings{
		Language:             session.Profile.Language,
		Phrases:              len(session.Blocks),
		AllowAdjacentRepeats: !session.PreventAdjacentRepeats,
	}
	m := NewModel(session, settings, []string{session.Profile.Language})
	final, err := tea.NewProgram(m, mouseProgramOpts()...).Run()
	if err != nil {
		return err
	}
	model := final.(Model)
	if model.session.Err != nil && !model.session.Done {
		return model.session.Err
	}
	if model.session.Quitting && !model.session.Done {
		return fmt.Errorf("cancelled")
	}
	if model.session.Done {
		fmt.Print(model.View())
	}
	return nil
}
