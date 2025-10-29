package smpp

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"smppizdez/account"
	"smppizdez/coding"
	"smppizdez/sender"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

var (
	commandMappings = map[data.CommandIDType]sender.Command{
		data.BIND_TRANSCEIVER:      sender.BindTransceiver,
		data.BIND_TRANSCEIVER_RESP: sender.BindTransceiverResp,
		data.BIND_RECEIVER:         sender.BindReceiver,
		data.BIND_RECEIVER_RESP:    sender.BindReceiverResp,
		data.BIND_TRANSMITTER:      sender.BindTransmitter,
		data.BIND_TRANSMITTER_RESP: sender.BindTransmitterResp,
		data.SUBMIT_SM:             sender.SubmitSM,
		data.SUBMIT_SM_RESP:        sender.SubmitSMResp,
		data.DELIVER_SM:            sender.DeliverSM,
		data.DELIVER_SM_RESP:       sender.DeliverSMResp,
		data.ENQUIRE_LINK:          sender.EnquireLink,
		data.ENQUIRE_LINK_RESP:     sender.EnquireLinkResp,
		data.UNBIND:                sender.Unbind,
		data.UNBIND_RESP:           sender.UnbindResp,
		data.GENERIC_NACK:          sender.GenericNack,
	}

	statusMappings = map[data.CommandStatusType]sender.CommandStatus{
		data.ESME_ROK:              sender.ESME_ROK,
		data.ESME_RINVMSGLEN:       sender.ESME_RINVMSGLEN,
		data.ESME_RINVCMDLEN:       sender.ESME_RINVCMDLEN,
		data.ESME_RINVCMDID:        sender.ESME_RINVCMDID,
		data.ESME_RINVBNDSTS:       sender.ESME_RINVBNDSTS,
		data.ESME_RALYBND:          sender.ESME_RALYBND,
		data.ESME_RINVPRTFLG:       sender.ESME_RINVPRTFLG,
		data.ESME_RINVREGDLVFLG:    sender.ESME_RINVREGDLVFLG,
		data.ESME_RSYSERR:          sender.ESME_RSYSERR,
		data.ESME_RINVSRCADR:       sender.ESME_RINVSRCADR,
		data.ESME_RINVDSTADR:       sender.ESME_RINVDSTADR,
		data.ESME_RINVMSGID:        sender.ESME_RINVMSGID,
		data.ESME_RBINDFAIL:        sender.ESME_RBINDFAIL,
		data.ESME_RINVPASWD:        sender.ESME_RINVPASWD,
		data.ESME_RINVSYSID:        sender.ESME_RINVSYSID,
		data.ESME_RCANCELFAIL:      sender.ESME_RCANCELFAIL,
		data.ESME_RREPLACEFAIL:     sender.ESME_RREPLACEFAIL,
		data.ESME_RMSGQFUL:         sender.ESME_RMSGQFUL,
		data.ESME_RINVSERTYP:       sender.ESME_RINVSERTYP,
		data.ESME_RINVNUMDESTS:     sender.ESME_RINVNUMDESTS,
		data.ESME_RINVDLNAME:       sender.ESME_RINVDLNAME,
		data.ESME_RINVDESTFLAG:     sender.ESME_RINVDESTFLAG,
		data.ESME_RINVSUBREP:       sender.ESME_RINVSUBREP,
		data.ESME_RINVESMCLASS:     sender.ESME_RINVESMCLASS,
		data.ESME_RCNTSUBDL:        sender.ESME_RCNTSUBDL,
		data.ESME_RSUBMITFAIL:      sender.ESME_RSUBMITFAIL,
		data.ESME_RINVSRCTON:       sender.ESME_RINVSRCTON,
		data.ESME_RINVSRCNPI:       sender.ESME_RINVSRCNPI,
		data.ESME_RINVDSTTON:       sender.ESME_RINVDSTTON,
		data.ESME_RINVDSTNPI:       sender.ESME_RINVDSTNPI,
		data.ESME_RINVSYSTYP:       sender.ESME_RINVSYSTYP,
		data.ESME_RINVREPFLAG:      sender.ESME_RINVREPFLAG,
		data.ESME_RINVNUMMSGS:      sender.ESME_RINVNUMMSGS,
		data.ESME_RTHROTTLED:       sender.ESME_RTHROTTLED,
		data.ESME_RINVSCHED:        sender.ESME_RINVSCHED,
		data.ESME_RINVEXPIRY:       sender.ESME_RINVEXPIRY,
		data.ESME_RINVDFTMSGID:     sender.ESME_RINVDFTMSGID,
		data.ESME_RX_T_APPN:        sender.ESME_RX_T_APPN,
		data.ESME_RX_P_APPN:        sender.ESME_RX_P_APPN,
		data.ESME_RX_R_APPN:        sender.ESME_RX_R_APPN,
		data.ESME_RQUERYFAIL:       sender.ESME_RQUERYFAIL,
		data.ESME_RINVOPTPARSTREAM: sender.ESME_RINVOPTPARSTREAM,
		data.ESME_ROPTPARNOTALLWD:  sender.ESME_ROPTPARNOTALLWD,
		data.ESME_RINVPARLEN:       sender.ESME_RINVPARLEN,
		data.ESME_RMISSINGOPTPARAM: sender.ESME_RMISSINGOPTPARAM,
		data.ESME_RINVOPTPARAMVAL:  sender.ESME_RINVOPTPARAMVAL,
		data.ESME_RDELIVERYFAILURE: sender.ESME_RDELIVERYFAILURE,
		data.ESME_RUNKNOWNERR:      sender.ESME_RUNKNOWNERR,
	}
)

type Sender struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	EnquireLink  time.Duration
}

type Session struct {
	handler       sender.PDUHandler
	defaultCoding coding.Coding
	conn          *gosmpp.Session
	tr            gosmpp.Transmitter
	lastErr       error
}

func (s *Session) SendMessage(req *sender.Request) error {
	segments, err := getSegments(req)
	if err != nil {
		return err
	}

	isMultiSegment := len(segments) > 1
	for _, seg := range segments {
		err = s.tr.Submit(seg.pd)
		if err != nil {
			return err
		}

		pduInfo := &sender.SubmitSMPDU{
			Header: sender.Header{
				Command:  sender.SubmitSM,
				Status:   sender.ESME_ROK,
				Sequence: uint32(seg.pd.SequenceNumber),
			},
			Ref:            int(seg.ref),
			Total:          int(seg.total),
			Seq:            int(seg.seq),
			IsMultiSegment: isMultiSegment,
		}
		s.handler(sender.Outbound, pduInfo)
	}

	return nil
}

func (s *Session) Close() error {
	return s.conn.Close()
}

func (s Sender) StartSession(
	acc *account.Account,
	handler sender.PDUHandler,
	onClose sender.CloseHandler,
) (sender.Session, error) {
	auth := gosmpp.Auth{
		SMSC:       fmt.Sprintf("%s:%d", acc.Host, acc.Port),
		SystemID:   acc.SystemID,
		Password:   acc.Password,
		SystemType: acc.SystemType,
	}

	var dialer gosmpp.Dialer
	if acc.TLS {
		dialer = tlsDialer
	} else {
		dialer = gosmpp.NonTLSDialer
	}

	var connector gosmpp.Connector
	switch acc.BindType {
	case account.Transceiver:
		connector = gosmpp.TRXConnector(dialer, auth)
	case account.Transmitter:
		connector = gosmpp.TXConnector(dialer, auth)
	default:
		connector = gosmpp.RXConnector(dialer, auth)
	}

	session := &Session{
		handler:       handler,
		defaultCoding: acc.DefaultCoding,
	}

	settings := gosmpp.Settings{
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		EnquireLink:  s.EnquireLink,
		OnAllPDU:     session.pduHandler,

		OnReceivingError: func(err error) {
			session.lastErr = err
		},
		OnSubmitError: func(_ pdu.PDU, err error) {
			session.lastErr = err
		},
		OnClosed: func(st gosmpp.State) {
			onClose(session.lastErr)
		},
	}

	var err error
	session.conn, err = gosmpp.NewSession(connector, settings, -1)
	if err != nil {
		return nil, err
	}

	session.tr = session.conn.Transmitter()
	return session, nil
}

func tlsDialer(addr string) (net.Conn, error) {
	cfg := tls.Config{MinVersion: tls.VersionTLS12}
	return tls.Dial("tcp", addr, &cfg)
}

func (s Sender) SupportedCodings() []coding.Coding {
	result := make([]coding.Coding, 0, len(supportedCodings))
	for cod := range supportedCodings {
		result = append(result, cod)
	}
	return result
}

func pduAddressToSender(addr pdu.Address) sender.Address {
	var result sender.Address

	tonByte := addr.Ton()
	switch tonByte {
	case data.GSM_TON_UNKNOWN:
		result.TON = sender.TONUnknown
	case data.GSM_TON_INTERNATIONAL:
		result.TON = sender.TONInternational
	case data.GSM_TON_NATIONAL:
		result.TON = sender.TONNational
	case data.GSM_TON_NETWORK:
		result.TON = sender.TONNetworkSpecific
	case data.GSM_TON_SUBSCRIBER:
		result.TON = sender.TONSubscriberNumber
	case data.GSM_TON_ALPHANUMERIC:
		result.TON = sender.TONAlphanumeric
	case data.GSM_TON_ABBREVIATED:
		result.TON = sender.TONAbbreviated
	default:
		result.TON = sender.TON(tonByte)
	}

	npiByte := addr.Npi()
	switch npiByte {
	case data.GSM_NPI_UNKNOWN:
		result.NPI = sender.NPIUnknown
	case data.GSM_NPI_ISDN:
		result.NPI = sender.NPIISDN
	case data.GSM_NPI_X121:
		result.NPI = sender.NPIData
	case data.GSM_NPI_TELEX:
		result.NPI = sender.NPITelex
	case data.GSM_NPI_LAND_MOBILE:
		result.NPI = sender.NPILandMobile
	case data.GSM_NPI_NATIONAL:
		result.NPI = sender.NPINational
	case data.GSM_NPI_PRIVATE:
		result.NPI = sender.NPIPrivate
	case data.GSM_NPI_ERMES:
		result.NPI = sender.NPIERMES
	case data.GSM_NPI_INTERNET:
		result.NPI = sender.NPIInternet
	case data.GSM_NPI_WAP_CLIENT_ID:
		result.NPI = sender.NPIWAP
	default:
		result.NPI = sender.NPI(npiByte)
	}

	result.Addr = addr.Address()

	return result
}

func getCodingByByte(defaultCoding coding.Coding, codingByte byte) (coding.Coding, encoding) {
	switch codingByte {
	case data.GSM7BITCoding:
		if defaultCoding == coding.GSM7 {
			return defaultCoding, data.GSM7BITPACKED.(encoding)
		} else {
			return defaultCoding, gsm8{}
		}
	case data.ASCIICoding:
		return coding.ASCII, data.ASCII.(encoding)
	case data.BINARY8BIT1Coding:
		return coding.Octet1, nil
	case data.LATIN1Coding:
		return coding.Latin1, data.LATIN1.(encoding)
	case data.BINARY8BIT2Coding:
		return coding.Octet2, nil
	case data.CYRILLICCoding:
		return coding.Cyrillic, data.CYRILLIC.(encoding)
	case data.HEBREWCoding:
		return coding.Hebrew, data.HEBREW.(encoding)
	case data.UCS2Coding:
		return coding.UCS2, ucs2f{}
	default:
		return coding.Coding(codingByte), nil
	}
}

func getPduHeader(pd pdu.PDU) (sender.Header, bool) {
	var hdr sender.Header
	var ok bool
	hdr.Command, ok = commandMappings[pd.GetHeader().CommandID]
	if !ok {
		return hdr, false
	}

	origStatus := pd.GetHeader().CommandStatus
	hdr.Status, ok = statusMappings[origStatus]
	if !ok {
		hdr.Status = sender.CommandStatus(origStatus)
	}
	hdr.Sequence = uint32(pd.GetHeader().SequenceNumber)
	return hdr, true
}

func (s *Session) pduHandler(pd pdu.PDU) (pdu.PDU, bool) {
	var response pdu.PDU
	var shouldClose bool
	if pd.CanResponse() {
		response = pd.GetResponse()
	}

	_, shouldClose = pd.(*pdu.Unbind)
	if !shouldClose {
		_, shouldClose = pd.(*pdu.UnbindResp)
	}

	hdr, ok := getPduHeader(pd)
	if !ok {
		return response, shouldClose
	}

	var pduInfo sender.PDU
	switch req := pd.(type) {
	case *pdu.SubmitSMResp:
		pduInfo = &sender.SubmitSMRespPDU{
			Header:    hdr,
			MessageID: req.MessageID,
		}

	case *pdu.DeliverSM:
		cod, dec := getCodingByByte(s.defaultCoding, req.Message.Encoding().DataCoding())

		var messageID, message string
		if field, ok := req.OptionalParameters[pdu.TagReceiptedMessageID]; ok {
			messageID = string(field.Data)
		}

		if dec != nil {
			var err error
			message, err = req.Message.GetMessageWithEncoding(dec)
			if err != nil {
				data, err := req.Message.GetMessageData()
				if err == nil {
					message = base64.StdEncoding.EncodeToString(data)
				}
			}
		}

		pduInfo = &sender.DeliverSMPDU{
			Header:      hdr,
			Source:      pduAddressToSender(req.SourceAddr),
			Destination: pduAddressToSender(req.DestAddr),
			EsmClass:    int(req.EsmClass),
			Coding:      cod,
			MessageID:   messageID,
			Message:     message,
		}
	default:
		pduInfo = &sender.GenericPDU{
			Header: hdr,
		}
	}

	s.handler(sender.Inbound, pduInfo)

	if response != nil {
		hdr, ok = getPduHeader(response)
		if ok {
			pduInfo = &sender.GenericPDU{Header: hdr}
			s.handler(sender.Outbound, pduInfo)
		}
	}

	return response, shouldClose
}
