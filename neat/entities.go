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
	KillTarget   string `qmap:"killtarget"`
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

type Sentence struct {
	Classname  *string `qmap:"classname,neat_sentence"`
	Origin     valve.Position
	TargetName string `qmap:"targetname"`

	MessageFlags int `qmap:"messageflags"`
	Target       string
	KillTarget   string `qmap:"killtarget"`
	Delay        float32

	Sentence      string
	Entity        string
	Listener      string
	Radius        float32
	Refire        float32
	Duration      float32
	Volume        float32            `qmap:"messagevolume"`
	SentenceFlags int                `qmap:"sentenceflags"`
	Attenuation   valve.Attenuation  `qmap:"messageattenuation"`
	TriggerState  valve.TriggerState `qmap:"triggerstate"`
}

func (ent Sentence) Validate(titles map[string]goldsrc.Title) error {
	if ent.TargetName == "" {
		return errors.New("empty targetname")
	}

	if ent.Sentence == "" {
		return errors.New("empty sentence")
	}

	if _, ok := titles[ent.Sentence]; !ok {
		return fmt.Errorf("message name '%s' not found in titles.txt", ent.Sentence)
	}

	return nil
}
