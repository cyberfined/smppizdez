package smpp

import "github.com/linxGnu/gosmpp/data"

const gsm7EscapeSequence byte = 0x1B

type gsm8 struct{}

func (gsm8) DataCoding() byte {
	return data.GSM7BITCoding
}

func (gsm8) Encode(str string) ([]byte, error) {
	return data.GSM7BIT.Encode(str)
}

func (gsm8) Decode(bytes []byte) (string, error) {
	return data.GSM7BIT.Decode(bytes)
}

func (gsm8) ShouldSplit(text string, octetLimit uint) bool {
	runeSlice := []rune(text)
	tLen := uint(len(runeSlice))
	escCharsLen := uint(len(data.GetEscapeChars(runeSlice)))
	regCharsLen := tLen - escCharsLen
	bytesLen := regCharsLen + escCharsLen*2
	return bytesLen > octetLimit
}

func (gsm8) EncodeSplit(text string, octetLimit uint) ([][]byte, error) {
	if octetLimit < 64 {
		octetLimit = 134
	}

	bytes, err := data.GSM7BIT.Encode(text)
	if err != nil {
		return nil, err
	}

	fr := uint(0)
	to := uint(0)
	var segments [][]byte
	for {
		for to < uint(len(bytes)) {
			inc := uint(1)
			if bytes[to] == gsm7EscapeSequence {
				inc = 2
			}

			if to-fr+inc > octetLimit {
				break
			}

			to += inc
		}

		if fr == to {
			break
		}

		segments = append(segments, bytes[fr:to])
		fr = to
	}

	return segments, nil
}
