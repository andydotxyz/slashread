package main

import (
	"image/color"
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	loadTheme(a)

	g := newGUI()
	w := g.makeWindow(a)

	g.setupActions(w)
	w.ShowAndRun()
}

// here you can add some button / callbacks code using widget IDs
func (g *gui) setupActions(w fyne.Window) {
	g.feed.Length = func() int {
		return 0
	}
	g.feed.Refresh()

	a := widget.NewActivity()
	prop := canvas.NewRectangle(color.Transparent)
	prop.SetMinSize(fyne.NewSquareSize(64))
	d := dialog.NewCustomWithoutButtons("Loading",
		container.NewStack(prop, a), w)

	a.Start()
	d.Show()

	cleanup := func() {
		d.Hide()
		a.Stop()
	}
	go g.loadFeed(cleanup, w)
}

func (g *gui) loadFeed(done func(), w fyne.Window) {
	rss, err := readFeed("https://rss.slashdot.org/Slashdot/slashdotMain")
	if err != nil {
		fyne.Do(func() {
			done()
			dialog.ShowError(err, w)
		})
		return
	}

	g.feed.Length = func() int {
		return len(rss.Items)
	}
	g.feed.CreateItem = func() fyne.CanvasObject {
		l := widget.NewLabel("This is an item that will have content")
		l.Wrapping = fyne.TextWrapWord

		return container.NewBorder(nil, nil, nil,
			widget.NewIcon(theme.MenuExpandIcon()),
			l)
	}
	g.feed.UpdateItem = func(id widget.ListItemID, o fyne.CanvasObject) {
		item := rss.Items[id]

		l := o.(*fyne.Container).Objects[0].(*widget.Label)
		l.SetText(item.Title)

		minHeight := l.MinSize().Height
		g.feed.SetItemHeight(id, minHeight)
	}
	g.feed.OnSelected = func(id widget.ListItemID) {
		item := rss.Items[id]

		g.showItem(item, g.nav, w)
	}

	fyne.Do(func() {
		g.feed.Refresh()
		done()
	})
}

func (g *gui) showItem(i Item, nav *container.Navigation, w fyne.Window) {
	v := newViewGUI()
	ui := v.makeUI()

	v.title.Wrapping = fyne.TextWrapWord
	v.title.ParseMarkdown("# " + i.Title)
	v.time.Text = durationSince(i.Date)

	v.content.Wrapping = fyne.TextWrapWord
	v.content.ParseMarkdown(i.Description)

	v.open.OnTapped = func() {
		u, _ := url.Parse(i.Link)
		_ = fyne.CurrentApp().OpenURL(u)
	}

	if runtime.GOOS == "android" {
		res := fyne.NewStaticResource("share-android.svg", shareAndroidBytes)
		v.share.SetIcon(theme.NewThemedResource(res))
	} else {
		res := fyne.NewStaticResource("share.svg", shareBytes)
		v.share.SetIcon(theme.NewThemedResource(res))
	}
	v.share.OnTapped = func() {
		fyne.CurrentApp().Clipboard().SetContent(i.Link)
		dialog.ShowInformation("Copied...", "Link copied to clipboard", w)
	}

	nav.PushWithTitle(ui, i.Title)
}
