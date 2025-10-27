package smpp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"smppizdez/coding"
	"smppizdez/sender"
	"sync/atomic"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

var MessageTooLong = errors.New("Message is too long")

var supportedCodings = map[coding.Coding]encoding{
	coding.GSM7: data.GSM7BITPACKED.(encoding),
	coding.GSM8: gsm8{},
	coding.UCS2: ucs2f{},
}

var ref uint32

const (
	udhSize     = 6
	maxSegments = 255
)

const (
	JISCoding         byte = 0x05
	PictogramCoding        = 0x09
	MusicCodesCoding       = 0x0a
	ExtendedJISCoding      = 0x0d
	KSC5601Coding          = 0x0e
)

type encoding interface {
	data.Encoding
	data.Splitter
}

type deceptiveEncoding struct {
	effective encoding
	deceptive byte
}

func (c deceptiveEncoding) DataCoding() byte {
	return c.deceptive
}

func (c deceptiveEncoding) Encode(str string) ([]byte, error) {
	return c.effective.Encode(str)
}

func (c deceptiveEncoding) Decode(data []byte) (string, error) {
	return c.effective.Decode(data)
}

func (c deceptiveEncoding) ShouldSplit(text string, octetLimit uint) bool {
	return c.effective.ShouldSplit(text, octetLimit)
}

func (c deceptiveEncoding) EncodeSplit(text string, octetLimit uint) ([][]byte, error) {
	return c.effective.EncodeSplit(text, octetLimit)
}

type segment struct {
	pd    *pdu.SubmitSM
	ref   uint16
	total byte
	seq   byte
}

func getSegments(req *sender.Request) ([]segment, error) {
	messages, enc, err := getSegmentMessages(req)
	if err != nil {
		return nil, err
	}

	pd, err := submitSmFromRequest(req)
	if err != nil {
		return nil, err
	}

	if len(messages) == 1 {
		if req.SplitMode == sender.SplitMessagePayload {
			messagePayload := pdu.Field{
				Tag:  pdu.TagMessagePayload,
				Data: messages[0],
			}
			pd.RegisterOptionalParam(messagePayload)
		} else {
			err = pd.Message.SetMessageDataWithEncoding(messages[0], enc)
			if err != nil {
				return nil, err
			}
		}

		segments := make([]segment, 1)
		segments[0] = segment{pd: pd}
		return segments, nil
	}

	if req.SplitMode == sender.SplitUDH {
		return getSegmentsUDH(pd, messages, enc)
	} else {
		return getSegmentsSAR(pd, messages, enc)
	}
}

func getSegmentsUDH(orig *pdu.SubmitSM, messages [][]byte, enc encoding) ([]segment, error) {
	total := byte(len(messages))
	curRef := byte(atomic.AddUint32(&ref, 1))
	seq := byte(1)

	segments := make([]segment, 0, len(messages))
	for _, message := range messages {
		udh := pdu.NewIEConcatMessage(total, seq, curRef)
		pd := new(pdu.SubmitSM)
		*pd = *orig
		pd.EsmClass |= data.SM_UDH_GSM

		pd.AssignSequenceNumber()
		pd.Message.SetUDH([]pdu.InfoElement{udh})
		err := pd.Message.SetMessageDataWithEncoding(message, enc)
		if err != nil {
			return nil, err
		}

		seg := segment{
			pd:    pd,
			ref:   uint16(curRef),
			total: total,
			seq:   seq,
		}

		seq++
		segments = append(segments, seg)
	}

	return segments, nil
}

func getSegmentsSAR(orig *pdu.SubmitSM, messages [][]byte, enc encoding) ([]segment, error) {
	var refData [2]byte
	curRef := uint16(atomic.AddUint32(&ref, 1))
	binary.BigEndian.PutUint16(refData[:], curRef)
	refTLV := pdu.Field{
		Tag:  pdu.TagSarMsgRefNum,
		Data: refData[:],
	}
	totalTLV := pdu.Field{
		Tag:  pdu.TagSarTotalSegments,
		Data: []byte{byte(len(messages))},
	}

	seq := byte(1)
	segments := make([]segment, 0, len(messages))
	for _, message := range messages {
		pd := new(pdu.SubmitSM)
		*pd = *orig
		pd.AssignSequenceNumber()
		pd.OptionalParameters = make(map[pdu.Tag]pdu.Field, len(orig.OptionalParameters)+3)

		for _, tlv := range orig.OptionalParameters {
			pd.RegisterOptionalParam(tlv)
		}

		seqTLV := pdu.Field{
			Tag:  pdu.TagSarSegmentSeqnum,
			Data: []byte{seq},
		}
		pd.RegisterOptionalParam(refTLV)
		pd.RegisterOptionalParam(totalTLV)
		pd.RegisterOptionalParam(seqTLV)

		err := pd.Message.SetMessageDataWithEncoding(message, enc)
		if err != nil {
			return nil, err
		}

		seg := segment{
			pd:    pd,
			ref:   curRef,
			total: byte(len(messages)),
			seq:   seq,
		}

		seq++
		segments = append(segments, seg)
	}

	return segments, nil
}

func submitSmFromRequest(req *sender.Request) (*pdu.SubmitSM, error) {
	var err error

	pd := pdu.NewSubmitSM().(*pdu.SubmitSM)
	pd.SourceAddr, err = convertAddress(req.Source)
	if err != nil {
		return nil, err
	}

	pd.DestAddr, err = convertAddress(req.Destination)
	if err != nil {
		return nil, err
	}

	pd.ValidityPeriod = req.ValidityPeriod
	pd.RegisteredDelivery = 0

	if req.RegisteredDelivery&sender.RdRequested != 0 {
		pd.RegisteredDelivery |= data.SM_SMSC_RECEIPT_REQUESTED
	}

	if req.RegisteredDelivery&sender.RdOnFailure != 0 {
		pd.RegisteredDelivery |= data.SM_SMSC_RECEIPT_ON_FAILURE
	}

	if req.RegisteredDelivery&sender.RdIntermediate != 0 {
		pd.RegisteredDelivery |= data.SM_NOTIF_REQUESTED
	}

	for _, tlv := range req.Optional {
		pd.RegisterOptionalParam(pdu.Field{
			Tag:  pdu.Tag(tlv.Tag),
			Data: tlv.Value,
		})
	}

	return pd, nil
}

func getSegmentMessages(req *sender.Request) ([][]byte, encoding, error) {
	var enc encoding
	effEnc, err := getCoding(req.EffectiveCoding)
	if err != nil {
		return nil, nil, err
	}
	decByte, err := getCodingByte(req.DeceptiveCoding)
	if err != nil {
		enc = effEnc
	} else {
		enc = deceptiveEncoding{deceptive: decByte, effective: effEnc}
	}

	var messages [][]byte
	if req.SplitMode == sender.SplitMessagePayload {
		msg, err := enc.Encode(req.Message)
		if err != nil {
			return nil, nil, err
		}
		messages = make([][]byte, 1)
		messages[0] = msg

		if len(messages) > 1 {
			return nil, nil, MessageTooLong
		}

		return messages, enc, nil
	} else if enc.ShouldSplit(req.Message, uint(req.BytePerSegment)) {
		var bytesPerMultiSegment uint
		switch req.SplitMode {
		case sender.SplitUDH:
			if req.BytePerSegment < udhSize+1 {
				bytesPerMultiSegment = udhSize + 1
			} else {
				bytesPerMultiSegment = uint(req.BytePerSegment - udhSize)
			}
		case sender.SplitNone:
			return nil, nil, MessageTooLong
		default:
			bytesPerMultiSegment = uint(req.BytePerSegment)
		}

		messages, err = enc.EncodeSplit(req.Message, bytesPerMultiSegment)
	} else {
		msg, err := enc.Encode(req.Message)
		if err == nil {
			messages = make([][]byte, 1)
			messages[0] = msg
		}
	}

	if err != nil {
		return nil, nil, err
	}

	if len(messages) > maxSegments {
		return nil, nil, MessageTooLong
	}

	return messages, enc, err
}

func getCoding(cod coding.Coding) (encoding, error) {
	enc, ok := supportedCodings[cod]
	if !ok {
		return nil, fmt.Errorf("Coding %s is unsupported", cod.String())
	}
	return enc, nil
}

func getCodingByte(cod coding.Coding) (byte, error) {
	switch cod {
	case coding.GSM7:
		fallthrough
	case coding.GSM8:
		return data.GSM7BITCoding, nil
	case coding.ASCII:
		return data.ASCIICoding, nil
	case coding.Octet1:
		return data.BINARY8BIT1Coding, nil
	case coding.Latin1:
		return data.LATIN1Coding, nil
	case coding.Octet2:
		return data.BINARY8BIT2Coding, nil
	case coding.JIS:
		return JISCoding, nil
	case coding.Cyrillic:
		return data.CYRILLICCoding, nil
	case coding.Hebrew:
		return data.HEBREWCoding, nil
	case coding.UCS2:
		return data.UCS2Coding, nil
	case coding.Pictogram:
		return PictogramCoding, nil
	case coding.MusicCodes:
		return MusicCodesCoding, nil
	case coding.ExtendedJIS:
		return ExtendedJISCoding, nil
	case coding.KSC5601:
		return KSC5601Coding, nil
	default:
		return 0, fmt.Errorf("Unknown coding enum value %d", cod)
	}
}

func convertAddress(addr sender.Address) (pdu.Address, error) {
	var result pdu.Address

	tonByte, err := tonToByte(addr.TON)
	if err != nil {
		return result, err
	}

	npiByte, err := npiToByte(addr.NPI)
	if err != nil {
		return result, err
	}

	result.SetTon(tonByte)
	result.SetNpi(npiByte)
	err = result.SetAddress(addr.Addr)
	return result, err
}

func tonToByte(t sender.TON) (byte, error) {
	switch t {
	case sender.TONUnknown:
		return data.GSM_TON_UNKNOWN, nil
	case sender.TONInternational:
		return data.GSM_TON_INTERNATIONAL, nil
	case sender.TONNational:
		return data.GSM_TON_NATIONAL, nil
	case sender.TONNetworkSpecific:
		return data.GSM_TON_NETWORK, nil
	case sender.TONSubscriberNumber:
		return data.GSM_TON_SUBSCRIBER, nil
	case sender.TONAlphanumeric:
		return data.GSM_TON_ALPHANUMERIC, nil
	case sender.TONAbbreviated:
		return data.GSM_TON_ABBREVIATED, nil
	default:
		return 0, fmt.Errorf("Unsupported TON value %s", t.String())
	}
}

func npiToByte(n sender.NPI) (byte, error) {
	switch n {
	case sender.NPIUnknown:
		return data.GSM_NPI_UNKNOWN, nil
	case sender.NPIISDN:
		return data.GSM_NPI_ISDN, nil
	case sender.NPIData:
		return data.GSM_NPI_X121, nil
	case sender.NPITelex:
		return data.GSM_NPI_TELEX, nil
	case sender.NPILandMobile:
		return data.GSM_NPI_LAND_MOBILE, nil
	case sender.NPINational:
		return data.GSM_NPI_NATIONAL, nil
	case sender.NPIPrivate:
		return data.GSM_NPI_PRIVATE, nil
	case sender.NPIERMES:
		return data.GSM_NPI_ERMES, nil
	case sender.NPIInternet:
		return data.GSM_NPI_INTERNET, nil
	case sender.NPIWAP:
		return data.GSM_NPI_WAP_CLIENT_ID, nil
	default:
		return 0, fmt.Errorf("Unsupported NPI value %s", n.String())
	}
}
