package valve

type Color string // "255 255 255"
type TriggerState int
type Position string // "0 0 0"
type Attenuation int

const (
	AttenuationSmallRadius Attenuation = iota
	AttenuationMediumRadius
	AttenuationLargeRadius
	AttenuationPlayEverywhere
)

const (
	TriggerStateOff TriggerState = iota
	TriggerStateOn
	TriggerStateToggle
)

type RenderMode uint8

const (
	RenderModeNormal RenderMode = iota
	RenderModeColor
	RenderModeTexture
	RenderModeGlow
	RenderModeSolid
	RenderModeAdditive
)

type MultiSource struct {
	ClassName *string `qmap:"classname,multisource"`
	Origin    Position

	GlobalState string `qmap:"globalstate"`
	Target      string
	TargetName  string `qmap:"targetname"`
}

type TriggerRelay struct {
	ClassName *string `qmap:"classname,trigger_relay"`
	Origin    Position

	Delay        float32
	Flags        uint8  `qmap:"spawnflags"`
	KillTarget   string `qmap:"killtarget"`
	Target       string
	TargetName   string       `qmap:"targetname"`
	TriggerState TriggerState `qmap:"triggerstate"`
}

// TODO: see if a button_target can function without a brush.
type ButtonTarget struct {
	ClassName *string `qmap:"classname,button_target"`
	Origin    Position

	Target       string
	TargetName   string `qmap:"targetname"`
	Master       string
	RenderFX     *uint8      `qmap:"renderfx,0"`
	RenderMode   *RenderMode `qmap:"rendermode,0"`
	RenderAmount *uint8      `qmap:"renderamt,255"`
	RenderColor  *Color      `qmap:"rendercolor,255 255 255"`
	Flags        uint8       `qmap:"spawnflags"`
}
