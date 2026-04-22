package neat

import (
	"errors"
	"fmt"

	"github.com/L-P/goldutil/goldsrc"
	"github.com/L-P/goldutil/goldsrc/qmap/valve"
)

type Master struct {
	Classname *string `qmap:"classname,neat_master"`
	Origin    valve.Position

	GlobalState string `qmap:"globalstate"`
	Target      string
	TargetName  string `qmap:"targetname"`
}

func (ent Master) Validate() error {
	if ent.TargetName == "" {
		return errors.New("empty targetname on neat_master")
	}
	if ent.Origin == "" {
		return errors.New("empty origin on neat_master")
	}

	return nil
}

type Message struct {
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

func (ent Message) Validate(titles map[string]goldsrc.Title) error {
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
