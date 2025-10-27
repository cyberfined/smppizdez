package sender

import (
	"fmt"
	"smppizdez/account"
	"smppizdez/coding"
)

type Request struct {
	Source             Address
	Destination        Address
	ValidityPeriod     string
	RegisteredDelivery RegisteredDelivery
	Message            string
	DeceptiveCoding    coding.Coding
	EffectiveCoding    coding.Coding
	SplitMode          SplitMode
	Optional           []TLV
	BytePerSegment     int
}

type Address struct {
	TON  TON
	NPI  NPI
	Addr string
}

type TON int

const (
	TONUnknown TON = iota + 1
	TONInternational
	TONNational
	TONNetworkSpecific
	TONSubscriberNumber
	TONAlphanumeric
	TONAbbreviated
)

func (t TON) String() string {
	switch t {
	case TONUnknown:
		return "Unknown"
	case TONInternational:
		return "International"
	case TONNational:
		return "National"
	case TONNetworkSpecific:
		return "Network Specific"
	case TONSubscriberNumber:
		return "Subscriber Number"
	case TONAlphanumeric:
		return "Alphanumeric"
	case TONAbbreviated:
		return "Abbreviated"
	default:
		return fmt.Sprintf("Unknown TON enum value (%d)", t)
	}
}

type NPI int

const (
	NPIUnknown NPI = iota + 1
	NPIISDN
	NPIData
	NPITelex
	NPILandMobile
	NPINational
	NPIPrivate
	NPIERMES
	NPIInternet
	NPIWAP
)

func (n NPI) String() string {
	switch n {
	case NPIUnknown:
		return "Unknown"
	case NPIISDN:
		return "ISDN"
	case NPIData:
		return "Data"
	case NPITelex:
		return "Telex"
	case NPILandMobile:
		return "Land Mobile"
	case NPINational:
		return "National"
	case NPIPrivate:
		return "Private"
	case NPIERMES:
		return "ERMES"
	case NPIInternet:
		return "Internet"
	case NPIWAP:
		return "WAP"
	default:
		return fmt.Sprintf("Unknown NPI enum value (%d)", n)
	}
}

type RegisteredDelivery int

const (
	RdNotRequested RegisteredDelivery = 0
	RdRequested                       = 1
	RdOnFailure                       = 2
	RdIntermediate                    = 4
)

func (rd RegisteredDelivery) String() string {
	var result string

	if rd&RdRequested != 0 {
		result = "Requested"
	}

	if rd&RdOnFailure != 0 {
		if result == "" {
			result = "On Failure"
		} else {
			result += ",On Failure"
		}
	}

	if rd&RdIntermediate != 0 {
		if result == "" {
			result = "Intermediate"
		} else {
			result += ",Intermediate"
		}
	}

	if result == "" {
		result = "Not requested"
	}

	return result
}

type SplitMode int

const (
	SplitUDH SplitMode = iota + 1
	SplitSAR
	SplitMessagePayload
	SplitNone
)

func (s SplitMode) String() string {
	switch s {
	case SplitUDH:
		return "UDH"
	case SplitSAR:
		return "SAR"
	case SplitMessagePayload:
		return "Message Payload"
	case SplitNone:
		return "None"
	default:
		return fmt.Sprintf("Unknown SplitMode enum value (%d)", s)
	}
}

type TLV struct {
	Tag   uint16
	Value []byte
}

type CommandStatus int

const (
	ESME_ROK CommandStatus = iota + 1
	ESME_RINVMSGLEN
	ESME_RINVCMDLEN
	ESME_RINVCMDID
	ESME_RINVBNDSTS
	ESME_RALYBND
	ESME_RINVPRTFLG
	ESME_RINVREGDLVFLG
	ESME_RSYSERR
	ESME_RINVSRCADR
	ESME_RINVDSTADR
	ESME_RINVMSGID
	ESME_RBINDFAIL
	ESME_RINVPASWD
	ESME_RINVSYSID
	ESME_RCANCELFAIL
	ESME_RREPLACEFAIL
	ESME_RMSGQFUL
	ESME_RINVSERTYP
	ESME_RINVNUMDESTS
	ESME_RINVDLNAME
	ESME_RINVDESTFLAG
	ESME_RINVSUBREP
	ESME_RINVESMCLASS
	ESME_RCNTSUBDL
	ESME_RSUBMITFAIL
	ESME_RINVSRCTON
	ESME_RINVSRCNPI
	ESME_RINVDSTTON
	ESME_RINVDSTNPI
	ESME_RINVSYSTYP
	ESME_RINVREPFLAG
	ESME_RINVNUMMSGS
	ESME_RTHROTTLED
	ESME_RINVSCHED
	ESME_RINVEXPIRY
	ESME_RINVDFTMSGID
	ESME_RX_T_APPN
	ESME_RX_P_APPN
	ESME_RX_R_APPN
	ESME_RQUERYFAIL
	ESME_RINVOPTPARSTREAM
	ESME_ROPTPARNOTALLWD
	ESME_RINVPARLEN
	ESME_RMISSINGOPTPARAM
	ESME_RINVOPTPARAMVAL
	ESME_RDELIVERYFAILURE
	ESME_RUNKNOWNERR
)

func (s CommandStatus) String() string {
	switch s {
	case ESME_ROK:
		return "ESME_ROK"
	case ESME_RINVMSGLEN:
		return "ESME_RINVMSGLEN"
	case ESME_RINVCMDLEN:
		return "ESME_RINVCMDLEN"
	case ESME_RINVCMDID:
		return "ESME_RINVCMDID"
	case ESME_RINVBNDSTS:
		return "ESME_RINVBNDSTS"
	case ESME_RALYBND:
		return "ESME_RALYBND"
	case ESME_RINVPRTFLG:
		return "ESME_RINVPRTFLG"
	case ESME_RINVREGDLVFLG:
		return "ESME_RINVREGDLVFLG"
	case ESME_RSYSERR:
		return "ESME_RSYSERR"
	case ESME_RINVSRCADR:
		return "ESME_RINVSRCADR"
	case ESME_RINVDSTADR:
		return "ESME_RINVDSTADR"
	case ESME_RINVMSGID:
		return "ESME_RINVMSGID"
	case ESME_RBINDFAIL:
		return "ESME_RBINDFAIL"
	case ESME_RINVPASWD:
		return "ESME_RINVPASWD"
	case ESME_RINVSYSID:
		return "ESME_RINVSYSID"
	case ESME_RCANCELFAIL:
		return "ESME_RCANCELFAIL"
	case ESME_RREPLACEFAIL:
		return "ESME_RREPLACEFAIL"
	case ESME_RMSGQFUL:
		return "ESME_RMSGQFUL"
	case ESME_RINVSERTYP:
		return "ESME_RINVSERTYP"
	case ESME_RINVNUMDESTS:
		return "ESME_RINVNUMDESTS"
	case ESME_RINVDLNAME:
		return "ESME_RINVDLNAME"
	case ESME_RINVDESTFLAG:
		return "ESME_RINVDESTFLAG"
	case ESME_RINVSUBREP:
		return "ESME_RINVSUBREP"
	case ESME_RINVESMCLASS:
		return "ESME_RINVESMCLASS"
	case ESME_RCNTSUBDL:
		return "ESME_RCNTSUBDL"
	case ESME_RSUBMITFAIL:
		return "ESME_RSUBMITFAIL"
	case ESME_RINVSRCTON:
		return "ESME_RINVSRCTON"
	case ESME_RINVSRCNPI:
		return "ESME_RINVSRCNPI"
	case ESME_RINVDSTTON:
		return "ESME_RINVDSTTON"
	case ESME_RINVDSTNPI:
		return "ESME_RINVDSTNPI"
	case ESME_RINVSYSTYP:
		return "ESME_RINVSYSTYP"
	case ESME_RINVREPFLAG:
		return "ESME_RINVREPFLAG"
	case ESME_RINVNUMMSGS:
		return "ESME_RINVNUMMSGS"
	case ESME_RTHROTTLED:
		return "ESME_RTHROTTLED"
	case ESME_RINVSCHED:
		return "ESME_RINVSCHED"
	case ESME_RINVEXPIRY:
		return "ESME_RINVEXPIRY"
	case ESME_RINVDFTMSGID:
		return "ESME_RINVDFTMSGID"
	case ESME_RX_T_APPN:
		return "ESME_RX_T_APPN"
	case ESME_RX_P_APPN:
		return "ESME_RX_P_APPN"
	case ESME_RX_R_APPN:
		return "ESME_RX_R_APPN"
	case ESME_RQUERYFAIL:
		return "ESME_RQUERYFAIL"
	case ESME_RINVOPTPARSTREAM:
		return "ESME_RINVOPTPARSTREAM"
	case ESME_ROPTPARNOTALLWD:
		return "ESME_ROPTPARNOTALLWD"
	case ESME_RINVPARLEN:
		return "ESME_RINVPARLEN"
	case ESME_RMISSINGOPTPARAM:
		return "ESME_RMISSINGOPTPARAM"
	case ESME_RINVOPTPARAMVAL:
		return "ESME_RINVOPTPARAMVAL"
	case ESME_RDELIVERYFAILURE:
		return "ESME_RDELIVERYFAILURE"
	case ESME_RUNKNOWNERR:
		return "ESME_RUNKNOWNERR"
	default:
		return fmt.Sprintf("CommandStatus(%d)", s)
	}
}

type Command int

const (
	BindTransceiver Command = iota + 1
	BindTransceiverResp
	BindReceiver
	BindReceiverResp
	BindTransmitter
	BindTransmitterResp
	SubmitSM
	SubmitSMResp
	DeliverSM
	DeliverSMResp
	EnquireLink
	EnquireLinkResp
	Unbind
	UnbindResp
	GenericNack
)

func (c Command) String() string {
	switch c {
	case BindTransceiver:
		return "BIND_TRANSCEIVER"
	case BindTransceiverResp:
		return "BIND_TRANSCEIVER_RESP"
	case BindReceiver:
		return "BIND_RECEIVER"
	case BindReceiverResp:
		return "BIND_RECEIVER_RESP"
	case BindTransmitter:
		return "BIND_TRANSMITTER"
	case BindTransmitterResp:
		return "BIND_TRANSMITTER_RESP"
	case SubmitSM:
		return "SUBMIT_SM"
	case SubmitSMResp:
		return "SUBMIT_SM_RESP"
	case DeliverSM:
		return "DELIVER_SM"
	case DeliverSMResp:
		return "DELIVER_SM_RESP"
	case EnquireLink:
		return "ENQUIRE_LINK"
	case EnquireLinkResp:
		return "ENQUIRE_LINK_RESP"
	case Unbind:
		return "UNBIND"
	case UnbindResp:
		return "UNBIND_RESP"
	case GenericNack:
		return "GENERIC_NACK"
	default:
		return fmt.Sprintf("Command(%d)", c)
	}
}

type Header struct {
	Command  Command
	Status   CommandStatus
	Sequence uint32
}

type PDU interface {
	GetHeader() Header
}

type GenericPDU struct {
	Header
}

func (p *GenericPDU) GetHeader() Header {
	return p.Header
}

type SubmitSMPDU struct {
	Header
	Ref            int
	Total          int
	Seq            int
	IsMultiSegment bool
}

func (s *SubmitSMPDU) GetHeader() Header {
	return s.Header
}

type SubmitSMRespPDU struct {
	Header
	MessageID string
}

func (p *SubmitSMRespPDU) GetHeader() Header {
	return p.Header
}

type DeliverSMPDU struct {
	Header
	Source      Address
	Destination Address
	EsmClass    int
	Coding      coding.Coding
	Message     string
	MessageID   string
}

func (p *DeliverSMPDU) GetHeader() Header {
	return p.Header
}

type Session interface {
	SendMessage(req *Request) error
	Close() error
}

type Direction int

const (
	Outbound Direction = iota + 1
	Inbound
)

func (d Direction) String() string {
	switch d {
	case Outbound:
		return "Outbound"
	case Inbound:
		return "Inbound"
	default:
		return fmt.Sprintf("Direction(%d)", d)
	}
}

type PDUHandler func(Direction, PDU)

type CloseHandler func(err error)

type Sender interface {
	SupportedCodings() []coding.Coding
	StartSession(
		acc *account.Account,
		handler PDUHandler,
		onClose CloseHandler,
	) (Session, error)
}
