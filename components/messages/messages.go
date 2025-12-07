package messages

import (
	"reddittui/model"

	tea "github.com/charmbracelet/bubbletea"
)

type ErrorModalMsg struct {
	ErrorMsg string
	OnClose  tea.Cmd
}

type (
	GoBackMsg          struct{}
	LoadThreadMsg      model.Post
	LoadHomeMsg        struct{}
	LoadMorePostsMsg   bool
	LoadCommunityMsg   string
	UpdateCommentsMsg  model.Comments
	UpdatePostsMsg     model.Posts
	AddMorePostsMsg    model.Posts
	LoadingCompleteMsg struct{}
	ShowComposePostMsg struct {
		Community string
	}
	ShowReplyModalMsg model.Post
	SubmitPostMsg     struct {
		Community string
		Content   string
	}
	SubmitReplyMsg struct {
		Post    model.Post
		Content string
	}
	PostPublishedMsg  model.Post
	ReplyPublishedMsg struct {
		Post    model.Post
		Comment model.Comment
	}
	PublishErrorMsg struct {
		ErrorMsg string
	}
	CopyNeventMsg struct {
		Post model.Post
	}

	OpenModalMsg        struct{}
	ExitModalMsg        struct{}
	ShowSpinnerModalMsg string

	ShowErrorModalMsg ErrorModalMsg

	OpenUrlMsg string
)

func GoBack() tea.Msg {
	return GoBackMsg{}
}

func LoadHome() tea.Msg {
	return LoadHomeMsg{}
}

func LoadMorePosts(home bool) tea.Cmd {
	return func() tea.Msg {
		return LoadMorePostsMsg(home)
	}
}

func LoadCommunity(community string) tea.Cmd {
	return func() tea.Msg {
		return LoadCommunityMsg(community)
	}
}

func LoadThread(post model.Post) tea.Cmd {
	return func() tea.Msg {
		return LoadThreadMsg(post)
	}
}

func ShowComposePost(community string) tea.Cmd {
	return func() tea.Msg {
		return ShowComposePostMsg{Community: community}
	}
}

func ShowReplyModal(post model.Post) tea.Cmd {
	return func() tea.Msg {
		return ShowReplyModalMsg(post)
	}
}

func LoadingComplete() tea.Msg {
	return LoadingCompleteMsg{}
}

func OpenModal() tea.Msg {
	return OpenModalMsg{}
}

func ExitModal() tea.Msg {
	return ExitModalMsg{}
}

func ShowSpinnerModal(loadingMsg string) tea.Cmd {
	return func() tea.Msg {
		return ShowSpinnerModalMsg(loadingMsg)
	}
}

func ShowErrorModal(errorMsg string) tea.Cmd {
	return func() tea.Msg {
		return ShowErrorModalMsg{ErrorMsg: errorMsg}
	}
}

func ShowErrorModalWithCallback(errorMsg string, callback tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return ShowErrorModalMsg{ErrorMsg: errorMsg, OnClose: callback}
	}
}

func HideSpinnerModal() tea.Msg {
	return ExitModalMsg{}
}

func OpenUrl(url string) tea.Cmd {
	return func() tea.Msg {
		return OpenUrlMsg(url)
	}
}
