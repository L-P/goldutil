package neat

import (
	"errors"
	"fmt"
	"goldutil/goldsrc"
	"goldutil/goldsrc/qmap/valve"
)

// Generic entity that can be used as a "master" in other entities.
// When targeting it:
//   - trigger_relay's targetstate is respected.
//   - All other callers are handled and will toggle by default, including
//     multi_manager, trigger_changetarget, path_track, and monsters using TriggerTarget.
//   - You can target <targetname>_on, <targetname>_off, or <targetname>_toggle
//     to set the corresponding state on the master without having to set up a
//     trigger_relay.
//
// The `target` and `globalstate` properties are written to the multisource.
//
// Internally, it works by setting up a func_button (<targetname>_proxy), a
// multisource (<targetname>), and three trigger_relays (<targetname>_on,
// <targetname>_off, and <targetname>_toggle).
// trigger_relays targeting <targetname> are redirected to <targetname>_proxy.
// All other entities targeting <targetname> will be rewritten to call <targetname>_toggle.
type NeatMaster struct {
	Classname *string `qmap:"classname,neat_master"`
	Origin    valve.Position

	GlobalState string `qmap:"globalstate"`
	Target      string
	TargetName  string `qmap:"targetname"`
}

func (ent NeatMaster) Validate() error {
	if ent.TargetName == "" {
		return errors.New("empty targetname on neat_master")
	}
	if ent.Origin == "" {
		return errors.New("empty origin on neat_master")
	}

	return nil
}

// Equivalent to env_message with two important differences:
//   - The message duration is read from titles.txt.
//   - The target isn't fired until the message ends.
//   - The delay is used as padding in addition to message length.
type NeatMessage struct {
	Classname *string `qmap:"classname,neat_message"`
	Origin    valve.Position

	TargetName   string `qmap:"targetname"`
	Target       string
	Message      string
	Delay        float32
	Flags        int                `qmap:"spawnflags"`
	Sound        string             `qmap:"messagesound"`
	Volume       string             `qmap:"messagevolume"`
	Attenuation  valve.Attenuation  `qmap:"messageattenuation"`
	TriggerState valve.TriggerState `qmap:"triggerstate"`
}

func (ent NeatMessage) Validate(titles map[string]goldsrc.Title) error {
	if ent.TargetName == "" {
		return errors.New("empty targetname")
	}

	if ent.Message == "" {
		return errors.New("empty message")
	}

	if _, ok := titles[ent.Message]; !ok {
		return fmt.Errorf("message name '%s' not found in titles.txt", ent.Message)
	}

	return nil
}
