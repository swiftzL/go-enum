// Code generated by go-enum
// DO NOT EDIT!

package example

import (
	"fmt"
	"strings"
)

const (
	// ColorBlack is a Color of type Black
	ColorBlack Color = iota
	// ColorWhite is a Color of type White
	ColorWhite
	// ColorRed is a Color of type Red
	ColorRed
	// ColorGreen is a Color of type Green
	ColorGreen Color = iota + 30
	// ColorBlue is a Color of type Blue
	ColorBlue
	// ColorGrey is a Color of type Grey
	ColorGrey
	// ColorYellow is a Color of type Yellow
	ColorYellow
	// ColorBlueGreen is a Color of type Blue-Green
	ColorBlueGreen
	// ColorRedOrange is a Color of type Red-Orange
	ColorRedOrange
)

const _ColorName = "BlackWhiteRedGreenBluegreyyellowblue-greenred-orange"

var _ColorMap = map[Color]string{
	0:  _ColorName[0:5],
	1:  _ColorName[5:10],
	2:  _ColorName[10:13],
	33: _ColorName[13:18],
	34: _ColorName[18:22],
	35: _ColorName[22:26],
	36: _ColorName[26:32],
	37: _ColorName[32:42],
	38: _ColorName[42:52],
}

// String implements the Stringer interface.
func (x Color) String() string {
	if str, ok := _ColorMap[x]; ok {
		return str
	}
	return fmt.Sprintf("Color(%d)", x)
}

var _ColorValue = map[string]Color{
	_ColorName[0:5]:                    0,
	strings.ToLower(_ColorName[0:5]):   0,
	_ColorName[5:10]:                   1,
	strings.ToLower(_ColorName[5:10]):  1,
	_ColorName[10:13]:                  2,
	strings.ToLower(_ColorName[10:13]): 2,
	_ColorName[13:18]:                  33,
	strings.ToLower(_ColorName[13:18]): 33,
	_ColorName[18:22]:                  34,
	strings.ToLower(_ColorName[18:22]): 34,
	_ColorName[22:26]:                  35,
	strings.ToLower(_ColorName[22:26]): 35,
	_ColorName[26:32]:                  36,
	strings.ToLower(_ColorName[26:32]): 36,
	_ColorName[32:42]:                  37,
	strings.ToLower(_ColorName[32:42]): 37,
	_ColorName[42:52]:                  38,
	strings.ToLower(_ColorName[42:52]): 38,
}

// ParseColor attempts to convert a string to a Color
func ParseColor(name string) (Color, error) {
	if x, ok := _ColorValue[name]; ok {
		return x, nil
	}
	return Color(0), fmt.Errorf("%s is not a valid Color", name)
}

// MarshalText implements the text marshaller method
func (x *Color) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method
func (x *Color) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseColor(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
