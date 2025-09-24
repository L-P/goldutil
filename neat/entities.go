package neat

import (
	"errors"
	"goldutil/goldsrc/typedmap/valve"
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

	Origin     valve.Position
	Target     string
	TargetName string `qmap:"targetname"`
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
