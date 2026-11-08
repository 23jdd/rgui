package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// NavigationWindow hosts a Frame and its history commands in a native window.
func NavigationWindow(config WindowConfig, nav *Navigator) error {
	if nav == nil {
		nav = NewNavigator(nil)
	}
	if config.Title == "" {
		config.Title = nav.Title()
		if config.Title == "" {
			config.Title = "Navigation"
		}
		callback := config.OnFrame
		previous := config.Title
		config.OnFrame = func(stats FrameStats) {
			if title := nav.Title(); title != "" && title != previous {
				rl.SetWindowTitle(title)
				previous = title
			}
			if callback != nil {
				callback(stats)
			}
		}
	}
	return Run(config, func() Node {
		return DockPanel(true,
			Docked{Edge: DockTop, Child: ToolBar(Button("Back", func() { nav.Back() }), Button("Forward", func() { nav.Forward() }))},
			Docked{Child: Frame(nav)})
	})
}
