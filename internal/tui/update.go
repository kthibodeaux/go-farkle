package tui

import (
	"strconv"
	"time"

	"github.com/ascii-arcade/farkle/internal/score"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.isRolling {
			return m, nil
		}

		m.error = ""

		switch msg.String() {
		case "r":
			if len(m.poolHeld) > 0 {
				m.error = "cannot roll with held dice"
				return m, nil
			}
			m.isRolling = true
			m.justRolled = true
			m.tickCount = 0
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return tickMsg{}
			})
		case "1":
			m.handleNumber(1)
		case "2":
			m.handleNumber(2)
		case "3":
			m.handleNumber(3)
		case "4":
			m.handleNumber(4)
		case "5":
			m.handleNumber(5)
		case "6":
			m.handleNumber(6)
		case "n":
			m.bust()
		case "y":
			m.bank()
		case "l":
			m.lock()
		case "u":
			if len(m.poolHeld) > 0 {
				die := m.poolHeld[len(m.poolHeld)-1]
				m.poolRoll.add(die)
				m.poolHeld.remove(die)
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		if m.tickCount < rollFrames {
			m.tickCount++
			m.poolRoll.roll()
			return m, tea.Tick(rollInterval, func(time.Time) tea.Msg {
				return tickMsg{}
			})
		}
		m.isRolling = false
		m.log.add(m.styledPlayerName(m.currentPlayerIndex) + " rolled " + m.poolRoll.renderCharacters())

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *model) handleNumber(n int) {
	if m.poolRoll.contains(n) {
		m.justRolled = false
		m.poolRoll.remove(n)
		m.poolHeld.add(n)
	}
}

func (m *model) bust() {
	m.justRolled = false
	m.log.add(m.styledPlayerName(m.currentPlayerIndex) + lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Render(" busted"))
	m.nextTurn()
}

func (m *model) nextTurn() {
	m.lockedInScore = 0
	m.poolHeld = newDicePool(0)
	m.poolRoll = newDicePool(6)

	if m.currentPlayerIndex == len(m.players)-1 {
		m.currentPlayerIndex = 0
	} else {
		m.currentPlayerIndex++
	}
}

func (m *model) bank() {
	if m.justRolled {
		m.error = "cannot bank immediately after rolling"
		return
	}
	if len(m.poolHeld) > 0 {
		m.error = "must lock in held dice before banking"
		return
	}
	if m.lockedInScore == 0 {
		m.error = "cannot bank 0 points"
		return
	}
	if m.players[m.currentPlayerIndex].score == 0 && m.lockedInScore < 500 {
		m.error = "must bank at least 500 points on the first turn"
		return
	}

	m.log.add(m.styledPlayerName(m.currentPlayerIndex) + " banked " + strconv.Itoa(m.lockedInScore) + " points")
	m.players[m.currentPlayerIndex].score += m.lockedInScore
	m.nextTurn()
}

func (m *model) lock() {
	if m.justRolled {
		m.error = "cannot lock immediately after rolling"
		return
	}
	if len(m.poolHeld) == 0 {
		m.error = "cannot lock with 0 held dice"
		return
	} else {
		score, err := score.Calculate(m.poolHeld)
		if err != nil {
			m.error = err.Error()
			return
		}

		m.lockedInScore += score
		m.log.add(m.styledPlayerName(m.currentPlayerIndex) + " locked " + m.poolHeld.renderCharacters() + " (+" + strconv.Itoa(score) + ", " + strconv.Itoa(m.lockedInScore) + ")")
		m.poolHeld = newDicePool(0)

		if len(m.poolRoll) == 0 {
			m.poolRoll = newDicePool(6)
		}
	}
}
