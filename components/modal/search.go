package modal

import (
	"reddittui/components/colors"
	"reddittui/components/messages"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	searchHelpText     = "Choose a community (NIP-73 identifier):"
	searchPlaceholder  = "community id (e.g. t:linux)"
	defaultSearchWidth = 35
)

var (
	searchHelpStyle  = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Text)).Italic(true)
	searchModelStyle = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Purple))
)

type CommunitySearchModal struct {
	textinput.Model
	style lipgloss.Style
}

func NewCommunitySearchModal() CommunitySearchModal {
	searchTextInput := textinput.New()
	searchTextInput.Placeholder = searchPlaceholder
	searchTextInput.ShowSuggestions = true
	searchTextInput.SetSuggestions(communitySuggestions)
	searchTextInput.CharLimit = 30

	return CommunitySearchModal{
		Model: searchTextInput,
		style: lipgloss.NewStyle(),
	}
}

func (s CommunitySearchModal) Init() tea.Cmd {
	return nil
}

func (s CommunitySearchModal) Update(msg tea.Msg) (CommunitySearchModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return s, messages.LoadCommunity(s.Value())
		case "esc":
			return s, messages.ExitModal
		}
	}

	var cmd tea.Cmd
	s.Model, cmd = s.Model.Update(msg)
	return s, cmd
}

func (s CommunitySearchModal) View() string {
	titleView := searchHelpStyle.Render(searchHelpText)
	modelView := searchModelStyle.Render(s.Model.View())
	joined := lipgloss.JoinVertical(lipgloss.Left, titleView, modelView)
	return s.style.Render(joined)
}

func (s *CommunitySearchModal) SetSize(w, h int) {
	searchW := min(w-s.style.GetHorizontalFrameSize(), defaultSearchWidth)
	s.style = s.style.Width(searchW)
}

func (s *CommunitySearchModal) Blur() {
	s.Model.Blur()
	s.Reset()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
