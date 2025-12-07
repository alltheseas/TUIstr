package components

import (
	"fmt"
	"log/slog"
	"tuistr/client"
	"tuistr/components/comments"
	"tuistr/components/messages"
	"tuistr/components/modal"
	"tuistr/components/posts"
	"tuistr/config"
	"tuistr/model"
	"tuistr/utils"

	tea "github.com/charmbracelet/bubbletea"
)

const defaultLoadingMessage = "connecting to nostr relays..."

type (
	pageType int
)

const (
	HomePage pageType = iota
	CommunityPage
	CommentsPage
)

type CommunitiesTui struct {
	nostrClient   *client.NostrClient
	homePage      posts.PostsPage
	communityPage posts.PostsPage
	commentsPage  comments.CommentsPage
	modalManager  modal.ModalManager
	popup         bool
	initializing  bool
	page          pageType
	prevPage      pageType
	loadingPage   pageType
	startCmd      tea.Cmd
}

func NewCommunitiesTui(configuration config.Config, communityArg, postID string) (CommunitiesTui, error) {
	nostrClient, err := client.NewNostrClient(configuration)
	if err != nil {
		return CommunitiesTui{}, err
	}

	homePage := posts.NewPostsPage(nostrClient, true)
	communityPage := posts.NewPostsPage(nostrClient, false)
	commentsPage := comments.NewCommentsPage(nostrClient)

	modalManager := modal.NewModalManager()

	startCmd := initialCommand(nostrClient, configuration.Communities, communityArg, postID)

	return CommunitiesTui{
		nostrClient:   nostrClient,
		homePage:      homePage,
		communityPage: communityPage,
		commentsPage:  commentsPage,
		modalManager:  modalManager,
		initializing:  true,
		startCmd:      startCmd,
	}, nil
}

func initialCommand(client *client.NostrClient, communities config.CommunitiesConfig, communityArg, postID string) tea.Cmd {
	switch {
	case communityArg != "":
		return messages.LoadCommunity(communityArg)
	case postID != "":
		return func() tea.Msg {
			post, err := client.GetPostByID(postID)
			if err != nil {
				slog.Error("Could not load event", "id", postID, "error", err)
				return messages.ShowErrorModalMsg{ErrorMsg: fmt.Sprintf("Could not load event %s", postID)}
			}
			return messages.LoadThreadMsg(post)
		}
	case communities.Default != "":
		return messages.LoadCommunity(communities.Default)
	default:
		return messages.LoadHome
	}
}

func (r CommunitiesTui) Init() tea.Cmd {
	return r.startCmd
}

func (r CommunitiesTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case messages.ShowErrorModalMsg:
		if r.initializing && msg.OnClose == nil {
			slog.Error("Error during initialization")
			if r.loadingPage == HomePage {
				errorMsg := "Could not initialize tuistr. Check the logfile for details."
				return r, messages.ShowErrorModalWithCallback(errorMsg, tea.Quit)
			}

			var errorMsg string
			if r.loadingPage == CommunityPage {
				errorMsg = "Error loading community. Returning to home page..."
			} else {
				errorMsg = "Error loading thread. Returning to home page..."
			}

			return r, messages.ShowErrorModalWithCallback(errorMsg, messages.LoadHome)
		}

	case messages.OpenModalMsg:
		r.focusModal()
		return r, nil

	case messages.LoadingCompleteMsg:
		cmd = r.completeLoading()
		return r, cmd

	case messages.ExitModalMsg:
		r.popup = false
		r.focusActivePage()
		cmd = r.modalManager.Blur()
		return r, cmd

	case messages.GoBackMsg:
		r.goBack()
		return r, nil

	case messages.LoadHomeMsg:
		if r.page == HomePage && !r.initializing {
			return r, r.modalManager.Blur()
		}

		r.focusModal()
		r.loadingPage = HomePage

		cmd = r.modalManager.SetLoading(defaultLoadingMessage)
		cmds = append(cmds, cmd)

	case messages.LoadCommunityMsg:
		community := string(msg)
		r.focusModal()
		r.loadingPage = CommunityPage

		loadingMsg := fmt.Sprintf("loading %s...", utils.NormalizeCommunity(community))
		cmd = r.modalManager.SetLoading(loadingMsg)
		cmds = append(cmds, cmd)

	case messages.LoadMorePostsMsg:
		r.focusModal()
		r.loadingPage = r.page

		cmd = r.modalManager.SetLoading("loading posts...")
		cmds = append(cmds, cmd)

	case messages.LoadThreadMsg:
		r.focusModal()
		r.loadingPage = CommentsPage

		cmd = r.modalManager.SetLoading("loading thread...")
		cmds = append(cmds, cmd)

	case messages.ShowComposePostMsg:
		r.focusModal()
		return r, r.modalManager.SetComposePost(msg.Community)

	case messages.ShowReplyModalMsg:
		r.focusModal()
		return r, r.modalManager.SetComposeReply(model.Post(msg))

	case messages.SubmitPostMsg:
		r.focusModal()
		r.loadingPage = r.page
		cmds = append(cmds, r.modalManager.SetLoading("publishing post..."), publishPost(r.nostrClient, msg))
		return r, tea.Batch(cmds...)

	case messages.SubmitReplyMsg:
		r.focusModal()
		r.loadingPage = CommentsPage
		cmds = append(cmds, r.modalManager.SetLoading("publishing reply..."), publishReply(r.nostrClient, msg))
		return r, tea.Batch(cmds...)

	case messages.PostPublishedMsg:
		r.popup = false
		cmds = append(cmds, r.modalManager.Blur())
		target := utils.NormalizeCommunity(model.Post(msg).Community)
		if target != "" {
			cmds = append(cmds, messages.LoadCommunity(target))
		} else {
			cmds = append(cmds, messages.LoadHome)
		}
		return r, tea.Batch(cmds...)

	case messages.ReplyPublishedMsg:
		r.popup = false
		cmds = append(cmds, r.modalManager.Blur(), messages.LoadThread(model.Post(msg.Post)))
		return r, tea.Batch(cmds...)

	case messages.PublishErrorMsg:
		return r, r.modalManager.SetError(msg.ErrorMsg)

	case messages.CopyNeventMsg:
		nevent, err := r.nostrClient.EncodeNevent(msg.Post)
		if err != nil {
			return r, r.modalManager.SetError(err.Error())
		}
		if err := utils.CopyToClipboard(nevent); err != nil {
			return r, r.modalManager.SetError(fmt.Sprintf("Could not copy nevent: %v", err))
		}
		slog.Info("copied nevent", "id", msg.Post.ID)
		return r, nil

	case messages.OpenUrlMsg:
		url := string(msg)
		if err := utils.OpenUrl(url); err != nil {
			slog.Error("Error opening url in browser", "url", url, "error", err.Error())
			cmd = r.modalManager.SetError(fmt.Sprintf("Could not open url %s in browser", url))
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		r.homePage.SetSize(msg.Width, msg.Height)
		r.communityPage.SetSize(msg.Width, msg.Height)
		r.commentsPage.SetSize(msg.Width, msg.Height)
		r.modalManager.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		}
	}

	r.modalManager, cmd = r.modalManager.Update(msg)
	cmds = append(cmds, cmd)

	r.homePage, cmd = r.homePage.Update(msg)
	cmds = append(cmds, cmd)

	r.communityPage, cmd = r.communityPage.Update(msg)
	cmds = append(cmds, cmd)

	r.commentsPage, cmd = r.commentsPage.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r CommunitiesTui) View() string {
	if r.popup {
		switch r.page {
		case HomePage:
			return r.modalManager.View(r.homePage)
		case CommunityPage:
			return r.modalManager.View(r.communityPage)
		case CommentsPage:
			return r.modalManager.View(r.commentsPage)
		}
	}

	switch r.page {
	case HomePage:
		return r.homePage.View()
	case CommunityPage:
		return r.communityPage.View()
	case CommentsPage:
		return r.commentsPage.View()
	}

	return ""
}

func (r *CommunitiesTui) goBack() {
	switch r.page {
	case CommentsPage:
		if r.prevPage == HomePage {
			r.setPage(HomePage)
		} else {
			r.setPage(CommunityPage)
		}
	default:
		r.setPage(HomePage)
	}

	r.focusActivePage()
}

func (r *CommunitiesTui) setPage(page pageType) {
	r.page, r.prevPage = page, r.page
}

func (r *CommunitiesTui) completeLoading() tea.Cmd {
	r.initializing = false
	r.popup = false
	r.setPage(r.loadingPage)
	r.focusActivePage()

	return r.modalManager.Blur()
}

func (r *CommunitiesTui) focusModal() {
	r.popup = true
	r.homePage.Blur()
	r.communityPage.Blur()
	r.commentsPage.Blur()
}

func (r *CommunitiesTui) focusActivePage() {
	switch r.page {
	case HomePage:
		r.homePage.Focus()
		r.communityPage.Blur()
		r.commentsPage.Blur()
	case CommunityPage:
		r.homePage.Blur()
		r.communityPage.Focus()
		r.commentsPage.Blur()
	case CommentsPage:
		r.homePage.Blur()
		r.communityPage.Blur()
		r.commentsPage.Focus()
	}
}

func publishPost(client *client.NostrClient, msg messages.SubmitPostMsg) tea.Cmd {
	return func() tea.Msg {
		post, err := client.PublishPost(msg.Community, msg.Content)
		if err != nil {
			return messages.PublishErrorMsg{ErrorMsg: err.Error()}
		}
		return messages.PostPublishedMsg(post)
	}
}

func publishReply(client *client.NostrClient, msg messages.SubmitReplyMsg) tea.Cmd {
	return func() tea.Msg {
		comment, err := client.PublishReply(msg.Post, msg.Content)
		if err != nil {
			return messages.PublishErrorMsg{ErrorMsg: err.Error()}
		}
		return messages.ReplyPublishedMsg{Post: msg.Post, Comment: comment}
	}
}
