package coding

import "fmt"

type Coding int

const (
	GSM7 Coding = iota + 1
	GSM8
	ASCII
	Octet1
	Latin1
	Octet2
	JIS
	Cyrillic
	Hebrew
	UCS2
	Pictogram
	MusicCodes
	ExtendedJIS
	KSC5601
)

var All = []Coding{
	GSM7,
	GSM8,
	ASCII,
	Octet1,
	Latin1,
	Octet2,
	JIS,
	Cyrillic,
	Hebrew,
	UCS2,
	Pictogram,
	MusicCodes,
	ExtendedJIS,
	KSC5601,
}

func (c Coding) String() string {
	switch c {
	case GSM7:
		return "GSM7"
	case GSM8:
		return "GSM8"
	case ASCII:
		return "ASCII"
	case Octet1:
		return "Octet1"
	case Latin1:
		return "Latin1"
	case Octet2:
		return "Octet2"
	case JIS:
		return "JIS"
	case Cyrillic:
		return "Cyrillic"
	case Hebrew:
		return "Hebrew"
	case UCS2:
		return "UCS2"
	case Pictogram:
		return "Pictogram"
	case MusicCodes:
		return "MusicCodes"
	case ExtendedJIS:
		return "ExtendedJIS"
	case KSC5601:
		return "KSC5601"
	default:
		return fmt.Sprintf("Unknown Coding enum value (%d)", c)
	}
}
