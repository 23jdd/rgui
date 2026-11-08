package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// SoundPlayerAction returns a handler that plays a caller-owned raylib sound.
// Pass it to Button or another event source. The audio device and sound must
// remain available while the action is in use.
func SoundPlayerAction(sound rl.Sound) func() {
	return func() {
		if rl.IsAudioDeviceReady() && rl.IsSoundValid(sound) {
			rl.PlaySound(sound)
		}
	}
}
