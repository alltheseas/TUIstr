package posts

import "github.com/charmbracelet/bubbles/key"

type postsKeyMap struct {
	Home   key.Binding
	Search key.Binding
	Back   key.Binding
	Load   key.Binding
	New    key.Binding
	Copy   key.Binding
}

var postsKeys = postsKeyMap{
	Home: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "home")),
	Search: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "community search")),
	Back: key.NewBinding(
		key.WithKeys("bs"),
		key.WithHelp("bs", "back")),
	Load: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "load more posts")),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new post")),
	Copy: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy nevent")),
}

func (k postsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Home, k.Search, k.Load, k.New, k.Copy}
}

func (k postsKeyMap) FullHelp() []key.Binding {
	return []key.Binding{k.Home, k.Search, k.Back, k.Load, k.New, k.Copy}
}
