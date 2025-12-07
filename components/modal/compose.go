package modal

import (
	"fmt"
	"strings"
	"tuistr/components/colors"
	"tuistr/components/messages"
	"tuistr/model"
	"tuistr/utils"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ComposeMode int

const (
	ComposePost ComposeMode = iota
	ComposeReply
)

type ComposeModal struct {
	textarea       textarea.Model
	communityInput textinput.Model
	mode           ComposeMode
	post           model.Post
	contextTitle   string
	errorMsg       string
	showCommunity  bool
	style          lipgloss.Style
	instructions   string
}

func NewComposeModal() ComposeModal {
	ta := textarea.New()
	ta.Placeholder = "Write your post..."
	ta.SetWidth(60)
	ta.SetHeight(8)
	ta.ShowLineNumbers = false
	ta.CharLimit = 1024

	ti := textinput.New()
	ti.Placeholder = "t:linux"
	ti.CharLimit = 50

	return ComposeModal{
		textarea:       ta,
		communityInput: ti,
		mode:           ComposePost,
		showCommunity:  true,
		style:          lipgloss.NewStyle(),
		instructions:   "ctrl+s to publish • esc to cancel",
	}
}

func (c ComposeModal) Init() tea.Cmd {
	return nil
}

func (c ComposeModal) Update(msg tea.Msg) (ComposeModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return c, messages.ExitModal
		case "ctrl+s":
			return c.submit()
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if c.showCommunity {
		c.communityInput, cmd = c.communityInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	c.textarea, cmd = c.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c ComposeModal) View() string {
	header := lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Text)).Bold(true).Render(c.contextTitle)
	bodyView := c.textarea.View()

	var communityView string
	if c.showCommunity {
		label := lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Text)).Render("community (topic id)")
		communityView = lipgloss.JoinVertical(lipgloss.Left, label, c.communityInput.View())
	}

	info := lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Subtext)).Render(c.instructions)

	errorView := ""
	if c.errorMsg != "" {
		errorView = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Red)).Render(c.errorMsg)
	}

	content := []string{header}
	if communityView != "" {
		content = append(content, communityView)
	}
	content = append(content, bodyView, info)
	if errorView != "" {
		content = append(content, errorView)
	}

	return c.style.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (c *ComposeModal) SetSize(w, h int) {
	usableW := w - c.style.GetHorizontalFrameSize()
	if usableW <= 0 {
		usableW = w
	}
	c.textarea.SetWidth(usableW - 2)
	c.communityInput.Width = usableW
}

func (c *ComposeModal) Focus() {
	c.textarea.Focus()
}

func (c *ComposeModal) Blur() {
	c.textarea.Blur()
	c.communityInput.Blur()
	c.textarea.SetValue("")
	c.communityInput.SetValue("")
	c.errorMsg = ""
}

func (c *ComposeModal) SetPostContext(community string) {
	c.mode = ComposePost
	c.post = model.Post{}
	c.showCommunity = true
	c.contextTitle = "New community post (topic only)"
	c.instructions = "ctrl+s to publish • esc to cancel"
	c.errorMsg = ""

	community = utils.NormalizeCommunity(community)
	if utils.ValidateTopic(community) {
		c.communityInput.SetValue(community)
	}
	c.textarea.SetValue("")
}

func (c *ComposeModal) SetReplyContext(post model.Post) {
	c.mode = ComposeReply
	c.post = post
	c.showCommunity = false
	c.contextTitle = fmt.Sprintf("Reply to %s", strings.TrimSpace(post.PostTitle))
	c.instructions = "ctrl+s to reply • esc to cancel"
	c.errorMsg = ""
	c.textarea.SetValue("")
}

func (c *ComposeModal) submit() (ComposeModal, tea.Cmd) {
	content := strings.TrimSpace(c.textarea.Value())
	if content == "" {
		c.errorMsg = "content is required"
		return *c, nil
	}

	if c.mode == ComposePost {
		community := strings.TrimSpace(c.communityInput.Value())
		if !utils.ValidateTopic(community) {
			c.errorMsg = "community must be a topic like t:linux"
			return *c, nil
		}
		return *c, func() tea.Msg {
			return messages.SubmitPostMsg{Community: community, Content: content}
		}
	}

	return *c, func() tea.Msg {
		return messages.SubmitReplyMsg{Post: c.post, Content: content}
	}
}
