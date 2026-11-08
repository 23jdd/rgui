package rgui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"rgui/draw"
)

type MediaPlayback uint8

const (
	MediaStopped MediaPlayback = iota
	MediaPlaying
	MediaPaused
)

// MediaBackend supplies playback and, optionally, a current frame texture.
// All methods run on the UI thread. Update returns true when the frame changed.
// Stop releases playback resources allocated for the mounted node.
type MediaBackend interface {
	Play()
	Pause()
	Stop()
	Update(time.Time) bool
	VideoTexture() rl.Texture2D
}

// MediaElement displays a media backend and play, pause, and stop controls.
func MediaElement(source MediaBackend, playback *Signal[MediaPlayback]) Node {
	if playback == nil {
		playback = NewSignal(MediaStopped)
	}
	surface := &mediaSurface{source: source, playback: playback}
	controls := ToolBar(
		Button("Play", func() { playback.Set(MediaPlaying) }),
		Button("Pause", func() { playback.Set(MediaPaused) }),
		Button("Stop", func() { playback.Set(MediaStopped) }),
	)
	return styleName(DockPanel(true, Docked{Edge: DockBottom, Child: controls}, Docked{Child: surface}), "mediaelement")
}

type mediaSurface struct {
	BaseNode
	source   MediaBackend
	playback *Signal[MediaPlayback]
	applied  MediaPlayback
}

func (m *mediaSurface) Measure(ctx MeasureContext) rl.Vector2 {
	return rl.Vector2{X: min(320, ctx.Available.X), Y: min(180, ctx.Available.Y)}
}

func (m *mediaSurface) tick(now time.Time, _ bool, _ *component) {
	if m.source == nil {
		return
	}
	request := m.playback.Get()
	if request != m.applied {
		switch request {
		case MediaPlaying:
			m.source.Play()
		case MediaPaused:
			m.source.Pause()
		default:
			m.source.Stop()
		}
		m.applied = request
		m.InvalidateVisual()
	}
	if request == MediaPlaying && m.source.Update(now) {
		m.InvalidateVisual()
	}
}

func (m *mediaSurface) release() {
	if m.source != nil && m.applied != MediaStopped {
		m.source.Stop()
	}
	m.applied = MediaStopped
}

func (m *mediaSurface) Primitives(ctx PaintContext) []Node {
	bounds := m.Bounds()
	var nodes []Node
	if m.source != nil {
		if texture := m.source.VideoTexture(); texture.ID != 0 && texture.Width > 0 && texture.Height > 0 {
			image := &imageNode{texture: texture, stretch: ImageUniform}
			image.Place(bounds)
			return append(nodes, image.Primitives(ctx)...)
		}
	}
	label := "Stopped"
	if m.source == nil {
		label = "No media"
	} else if m.applied == MediaPlaying {
		label = "Playing"
	} else if m.applied == MediaPaused {
		label = "Paused"
	}
	style := ctx.TextStyle()
	size := draw.MeasureText(label, style)
	return append(nodes, &TextNode{Position: rl.Vector2{X: bounds.X + (bounds.Width-size.X)/2, Y: bounds.Y + (bounds.Height-size.Y)/2}, Text: label, Style: style})
}

// MusicMedia adapts a caller-owned raylib music stream. Initialize the raylib
// audio device before using it and unload the stream after detaching the node.
func MusicMedia(music rl.Music) MediaBackend { return &musicMedia{music: music} }

type musicMedia struct {
	music           rl.Music
	paused, playing bool
}

func (m *musicMedia) Play() {
	if !rl.IsAudioDeviceReady() || !rl.IsMusicValid(m.music) {
		return
	}
	if m.paused {
		rl.ResumeMusicStream(m.music)
	} else {
		rl.PlayMusicStream(m.music)
	}
	m.paused, m.playing = false, true
}

func (m *musicMedia) Pause() {
	if m.playing {
		rl.PauseMusicStream(m.music)
		m.paused, m.playing = true, false
	}
}

func (m *musicMedia) Stop() {
	if m.paused || m.playing {
		rl.StopMusicStream(m.music)
	}
	m.paused, m.playing = false, false
}

func (m *musicMedia) Update(time.Time) bool {
	if m.playing {
		rl.UpdateMusicStream(m.music)
	}
	return false
}

func (*musicMedia) VideoTexture() rl.Texture2D { return rl.Texture2D{} }
