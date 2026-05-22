package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	cacheKey     = "slashdotMain.rss"
	cacheTimeKey = "slashdotMain.time"
)

var (
	pushed    = false
	lastFetch = time.Now().Add(-100 * time.Hour) // before trigger
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

	g.nav.OnBack = func() {
		pushed = false
		g.nav.Back()
	}
	if mob, ok := fyne.CurrentApp().Driver().(mobile.Driver); ok {
		w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
			if ev.Name == mobile.KeyBack {
				if pushed {
					pushed = false
					g.nav.Back()
					return
				}

				mob.GoBack()
			}
		})
	}

	cleanup := func() {
		d.Hide()
		a.Stop()
	}
	fyne.CurrentApp().Lifecycle().SetOnEnteredForeground(func() {
		if time.Now().Sub(lastFetch).Hours() < 1 {
			return
		}

		lastFetch = time.Now()
		a.Start()
		d.Show()
		go g.loadFeed(cleanup, w)
	})
}

func (g *gui) loadFeed(done func(), w fyne.Window) {
	cacheTimeStr := ""

	// if we have a cache...
	if fyne.CurrentApp().Cache().Exists(cacheKey) {
		read, err := fyne.CurrentApp().Cache().Read(cacheKey)
		if err != nil {
			log.Println("Failed to read cached feed") // not important
		} else {
			_ = g.loadFeedFromReader(read, done, w) // ignore
			_ = read.Close()

			cacheTimeStr = fyne.CurrentApp().Preferences().String(cacheTimeKey)
		}
	}

	// then read from the web
	resp, err := http.Get("https://rss.slashdot.org/Slashdot/slashdotMain")
	if err != nil {
		fyne.Do(func() {
			done()
			age := cacheRecency(cacheTimeStr)
			dialog.ShowInformation("Error", "Failed to connect to slashdot, "+age, w)
		})
		return
	}

	copier := &bytes.Buffer{}
	read := io.TeeReader(resp.Body, copier)
	err = g.loadFeedFromReader(read, done, w)
	if err != nil {
		fyne.Do(func() {
			done()
			age := cacheRecency(cacheTimeStr)
			dialog.ShowInformation("Error", "Failed to load feed, "+age, w)
		})
		return
	}
	_ = resp.Body.Close()

	write, err := fyne.CurrentApp().Cache().Write(cacheKey)

	_, err = io.Copy(write, copier)
	_ = write.Close()
	if err != nil {
		log.Println("Failed to save cached feed, removing") // not important
		fyne.CurrentApp().Cache().Remove(cacheKey)
	}

	now := time.Now().Format(time.DateTime)
	fyne.CurrentApp().Preferences().SetString(cacheTimeKey, now)
	done()
}

func (g *gui) loadFeedFromReader(r io.Reader, done func(), w fyne.Window) error {
	rss, err := readFeed(r)
	if err != nil {
		return err
	}

	g.feed.Length = func() int {
		return len(rss.Items)
	}
	g.feed.CreateItem = func() fyne.CanvasObject {
		l := widget.NewRichTextFromMarkdown("This is an item that will have content")
		l.Wrapping = fyne.TextWrapWord

		icon := &canvas.Image{}
		icon.FillMode = canvas.ImageFillContain
		icon.SetMinSize(fyne.NewSquareSize(32))
		return container.NewBorder(nil, nil, icon,
			widget.NewIcon(theme.MenuExpandIcon()),
			l)
	}
	g.feed.UpdateItem = func(id widget.ListItemID, o fyne.CanvasObject) {
		item := rss.Items[id]

		l := o.(*fyne.Container).Objects[0].(*widget.RichText)
		l.ParseMarkdown(item.Title)

		i := o.(*fyne.Container).Objects[1].(*canvas.Image)
		loadIcon(item, i)

		minHeight := l.MinSize().Height
		g.feed.SetItemHeight(id, minHeight)
	}
	g.feed.OnSelected = func(id widget.ListItemID) {
		item := rss.Items[id]

		go func() {
			g.showItem(item, g.nav, w)
			time.Sleep(canvas.DurationStandard)
			fyne.Do(func() {
				g.feed.Unselect(id)
			})
		}()
	}

	fyne.Do(func() {
		g.feed.Refresh()
		done()
	})
	return nil
}

func (g *gui) showItem(i Item, nav *container.Navigation, w fyne.Window) {
	pushed = true

	v := newViewGUI()
	ui := v.makeUI()

	v.title.Wrapping = fyne.TextWrapWord
	v.title.ParseMarkdown("# " + i.Title)
	v.time.Text = fmt.Sprintf("Posted by %s %s.", i.Creator, durationSince(i.Date))

	v.content.Wrapping = fyne.TextWrapWord
	v.content.ParseMarkdown(i.Description)

	loadIcon(i, v.section)

	v.open.OnTapped = func() {
		u, _ := url.Parse(i.Link)
		_ = fyne.CurrentApp().OpenURL(u)
	}
	v.share.OnTapped = func() {
		fyne.CurrentApp().Clipboard().SetContent(i.Link)
		dialog.ShowInformation("Copied...", "Link copied to clipboard", w)
	}

	fyne.Do(func() {
		if runtime.GOOS == "android" {
			res := fyne.NewStaticResource("share-android.svg", shareAndroidBytes)
			v.share.SetIcon(theme.NewThemedResource(res))
		} else {
			res := fyne.NewStaticResource("share.svg", shareBytes)
			v.share.SetIcon(theme.NewThemedResource(res))
		}

		nav.PushWithTitle(ui, i.Title)
	})
}

func loadIcon(i Item, img *canvas.Image) {
	go func() { // potentially slow on first load
		res, err := fyne.CacheResourceFromURLString(i.ImageURL())
		if err != nil {
			fyne.LogError("Failed to read section image", err)
		}

		fyne.Do(func() {
			img.Image = nil
			img.Resource = res
			img.Refresh()
		})
	}()
}

func cacheRecency(prevTime string) string {
	cacheTime, _ := time.Parse(time.DateTime, prevTime)
	diff := time.Now().Sub(cacheTime)

	prefix := "using feed downloaded "
	if diff < time.Hour {
		return prefix + "recently"
	} else if diff < time.Hour*24 {
		return prefix + strconv.Itoa(int(diff.Hours())) + " hours ago"
	} else if diff < time.Hour*24*7 {
		return prefix + strconv.Itoa(int(diff.Hours()/24)) + " days ago"
	} else {
		return prefix + strconv.Itoa(int(diff.Hours()/24/7)) + " weeks ago"
	}
}
