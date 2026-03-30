package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/harmonica"
)

// progressModel displays an animated progress bar during package loading.
type progressModel struct {
	progress progress.Model
	spring   harmonica.Spring
	velocity float64
	percent  float64
	target   float64
	styles   Styles
}

func newProgressModel(styles Styles) progressModel {
	p := progress.New(
		progress.WithColors(
			lipgloss.Color(ColorProgressA),
			lipgloss.Color(ColorProgressB),
		),
		progress.WithScaled(true),
	)
	p.SetWidth(50)

	return progressModel{
		progress: p,
		spring:   harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.0),
		styles:   styles,
	}
}

// progressTickMsg drives the spring animation.
type progressTickMsg struct{}

func progressTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(_ time.Time) tea.Msg {
		return progressTickMsg{}
	})
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg.(type) {
	case progressTickMsg:
		m.percent, m.velocity = m.spring.Update(m.percent, m.velocity, m.target)
		// Clamp
		if m.percent > 1.0 {
			m.percent = 1.0
		}
		if m.percent < 0 {
			m.percent = 0
		}
		return m, progressTick()
	}

	pm, cmd := m.progress.Update(msg)
	m.progress = pm
	return m, cmd
}

// SetTarget sets the target percentage (0.0–1.0) for the spring animation.
func (m *progressModel) SetTarget(t float64) {
	m.target = t
}

// View renders the progress bar.
func (m progressModel) View() string {
	return m.progress.ViewAs(m.percent)
}

// ProgressInfo renders the progress bar with detector info.
func (m progressModel) ProgressInfo(done, total int, currentDetector string, pkgCount int) string {
	bar := m.View()

	info := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim)).Render(
		fmt.Sprintf("  %d/%d detectors  ·  %d packages found", done, total, pkgCount),
	)

	current := ""
	if currentDetector != "" {
		current = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccent)).Render(
			fmt.Sprintf("  scanning %s…", currentDetector),
		)
	}

	return fmt.Sprintf("%s\n%s%s", bar, info, current)
}
