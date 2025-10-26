package main

import (
	"fmt"
	"slices"
	"smppizdez/account"
	"smppizdez/coding"
	"smppizdez/sender"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var submitSmTons = []sender.TON{
	sender.TONUnknown,
	sender.TONInternational,
	sender.TONNational,
	sender.TONNetworkSpecific,
	sender.TONSubscriberNumber,
	sender.TONAlphanumeric,
	sender.TONAbbreviated,
}

var submitSmNpis = []sender.NPI{
	sender.NPIUnknown,
	sender.NPIISDN,
	sender.NPIData,
	sender.NPITelex,
	sender.NPILandMobile,
	sender.NPINational,
	sender.NPIPrivate,
	sender.NPIERMES,
	sender.NPIInternet,
	sender.NPIWAP,
}

var submitSmStartSessionCallback func(*account.Account)

type radioBtnSplitMode struct {
	btn  *gtk.RadioButton
	mode sender.SplitMode
}

type submitSmContext struct {
	sender            sender.Sender
	session           sender.Session
	submitSmForm      *gtk.Grid
	srcTonSelector    *gtk.ComboBox
	srcNpiSelector    *gtk.ComboBox
	srcAddrEntry      *gtk.Entry
	dstTonSelector    *gtk.ComboBox
	dstNpiSelector    *gtk.ComboBox
	dstAddrEntry      *gtk.Entry
	validityEntry     *gtk.Entry
	effCodingSelector *gtk.ComboBox
	decCodingSelector *gtk.ComboBox
	rdReqCheck        *gtk.CheckButton
	rdFailCheck       *gtk.CheckButton
	rdInterCheck      *gtk.CheckButton
	messageEntry      *gtk.TextView
	spltRadios        []radioBtnSplitMode
	segmentBytesEntry *gtk.Entry
	messageLabel      *gtk.Label
	logsArea          *gtk.TextView
	logsScroller      *gtk.ScrolledWindow
	effectiveCoding   coding.Coding
}

func codingToString(cod coding.Coding) string {
	switch cod {
	case coding.GSM7:
		return "GSM7"
	case coding.GSM8:
		return "GSM8"
	case coding.ASCII:
		return "ASCII"
	case coding.Octet1:
		return "Octet1"
	case coding.Latin1:
		return "Latin1"
	case coding.Octet2:
		return "Octet2"
	case coding.JIS:
		return "JIS"
	case coding.Cyrillic:
		return "Cyrillic"
	case coding.Hebrew:
		return "Hebrew"
	case coding.UCS2:
		return "UCS2"
	case coding.Pictogram:
		return "Pictogram"
	case coding.MusicCodes:
		return "Music Codes"
	case coding.ExtendedJIS:
		return "Extended JIS"
	case coding.KSC5601:
		return "KSC5601"
	default:
		return "Unknown"
	}
}

func (ctx *submitSmContext) initCodingSelectors() {
	effStore, _ := gtk.ListStoreNew(glib.TYPE_STRING)
	supportedCodings := ctx.sender.SupportedCodings()
	slices.Sort(supportedCodings)
	for _, cod := range supportedCodings {
		iter := effStore.Append()
		effStore.Set(iter, []int{0}, []any{codingToString(cod)})
	}
	ctx.effCodingSelector.SetModel(effStore)
	effStore.Unref()
	column, _ := gtk.CellRendererTextNew()
	ctx.effCodingSelector.CellLayout.PackStart(column, true)
	ctx.effCodingSelector.CellLayout.AddAttribute(column, "text", 0)
	ctx.effCodingSelector.SetActive(0)
	ctx.effectiveCoding = supportedCodings[0]
	ctx.effCodingSelector.Connect("changed", func() {
		effIdx := getComboIndex(ctx.effCodingSelector)
		cod := supportedCodings[effIdx]
		ctx.effectiveCoding = cod

		for i := range coding.All {
			if coding.All[i] == cod {
				ctx.decCodingSelector.SetActive(i)
				break
			}
		}
	})

	decStore, _ := gtk.ListStoreNew(glib.TYPE_STRING)
	decActiveIdx := 0
	for i, cod := range coding.All {
		if cod == supportedCodings[0] {
			decActiveIdx = i
		}
		iter := decStore.Append()
		decStore.Set(iter, []int{0}, []any{codingToString(cod)})
	}
	ctx.decCodingSelector.SetModel(decStore)
	decStore.Unref()
	column, _ = gtk.CellRendererTextNew()
	ctx.decCodingSelector.CellLayout.PackStart(column, true)
	ctx.decCodingSelector.CellLayout.AddAttribute(column, "text", 0)
	ctx.decCodingSelector.SetActive(decActiveIdx)
}

func initSubmitSmForm(builder *gtk.Builder, s sender.Sender) {
	ctx := submitSmContext{
		sender:            s,
		srcTonSelector:    getComboById(builder, "source_ton_selector"),
		srcNpiSelector:    getComboById(builder, "source_npi_selector"),
		srcAddrEntry:      getEntryById(builder, "source_addr_input"),
		dstTonSelector:    getComboById(builder, "dest_ton_selector"),
		dstNpiSelector:    getComboById(builder, "dest_npi_selector"),
		dstAddrEntry:      getEntryById(builder, "dest_addr_input"),
		validityEntry:     getEntryById(builder, "validity_input"),
		effCodingSelector: getComboById(builder, "effective_coding_selector"),
		decCodingSelector: getComboById(builder, "deceptive_coding_selector"),
		rdReqCheck:        getCheckById(builder, "rd_requested"),
		rdFailCheck:       getCheckById(builder, "rd_failure"),
		rdInterCheck:      getCheckById(builder, "rd_intermediate"),
		messageEntry:      getTextViewById(builder, "message_input"),
		spltRadios: []radioBtnSplitMode{
			{btn: getRadioById(builder, "splt_udh_radio"), mode: sender.SplitUDH},
			{btn: getRadioById(builder, "splt_sar_radio"), mode: sender.SplitSAR},
			{
				btn:  getRadioById(builder, "splt_message_payload_radio"),
				mode: sender.SplitMessagePayload,
			},
			{btn: getRadioById(builder, "splt_none_radio"), mode: sender.SplitNone},
		},
		segmentBytesEntry: getEntryById(builder, "segment_bytes_input"),
		logsArea:          getTextViewById(builder, "logs_area"),
		messageLabel:      getLabelById(builder, "submit_sm_message_label"),
	}

	msgBuf, _ := ctx.messageEntry.GetBuffer()
	msgBuf.Connect("changed", func() {
		msgStart, msgEnd := msgBuf.GetBounds()
		message, _ := msgBuf.GetText(msgStart, msgEnd, true)
		var labelText string
		if len(message) == 1 {
			labelText = "Message (1 character)"
		} else {
			labelText = fmt.Sprintf("Message (%d characters)", len(message))
		}
		ctx.messageLabel.SetText(labelText)
	})

	gridI, _ := builder.GetObject("submit_sm_grid")
	ctx.submitSmForm = gridI.(*gtk.Grid)

	scrollerI, _ := builder.GetObject("logs_scroller")
	ctx.logsScroller = scrollerI.(*gtk.ScrolledWindow)

	ctx.initCodingSelectors()

	submitSmStartSessionCallback = ctx.startSession

	sendBtn := getButtonById(builder, "send_button")
	sendBtn.Connect("pressed", func() {
		ctx.resetStyles()
		if req := ctx.getRequest(); req != nil {
			err := ctx.session.SendMessage(req)
			if err != nil {
				errorDialog("Message submission error: %v", err)
			}
		}
	})

	unbindBtn := getButtonById(builder, "unbind_button")
	unbindBtn.Connect("pressed", func() {
		ctx.submitSmForm.SetSensitive(false)
		if ctx.session == nil {
			return
		}
		err := ctx.session.Close()
		if err != nil {
			errorDialog("Session closing error: %v", err)
		}
	})
}

func (ctx *submitSmContext) startSession(acc *account.Account) {
	glib.IdleAdd(func() {
		buf, err := ctx.logsArea.GetBuffer()
		if err != nil {
			return
		}
		logsStart, logsEnd := buf.GetBounds()
		buf.Delete(logsStart, logsEnd)
	})

	var err error
	if ctx.session != nil {
		err = ctx.session.Close()
		if err != nil {
			errorDialog("Session closing error: %v", err)
		}
	}

	ctx.session, err = ctx.sender.StartSession(acc, ctx.pduHandler, ctx.sessionCloseHandler)
	if err != nil {
		errorDialog("Failed to start SMPP session: %v", err)
		return
	}
	ctx.submitSmForm.SetSensitive(true)
}

func (ctx *submitSmContext) pduHandler(dir sender.Direction, pdu sender.PDU) {
	var dirStr string
	if dir == sender.Inbound {
		dirStr = "Incoming PDU"
	} else {
		dirStr = "Outcoming PDU"
	}

	var log string
	switch req := pdu.(type) {
	case *sender.SubmitSMPDU:
		if req.IsMultiSegment {
			log = fmt.Sprintf(
				"%s\n    Command: %v\n    Status: %v\n   Sequence: %d\n    "+
					"Ref: %d\n    Total: %d\n    Seq: %d\n",
				dirStr,
				req.Command,
				req.Status,
				req.Sequence,
				req.Ref,
				req.Total,
				req.Seq,
			)
		} else {
			log = fmt.Sprintf(
				"%s\n    Command: %v\n    Status: %v\n   Sequence: %d\n",
				dirStr,
				req.Command,
				req.Status,
				req.Sequence,
			)
		}
	case *sender.SubmitSMRespPDU:
		log = fmt.Sprintf(
			"%s\n    Command: %v\n    Status: %v\n   Sequence: %d\n    "+
				"Message ID: %s\n",
			dirStr,
			req.Command,
			req.Status,
			req.Sequence,
			req.MessageID,
		)
	case *sender.DeliverSMPDU:
		log = fmt.Sprintf(
			"%s\n    Command: %v\n    Status: %v\n   Sequence: %d\n    "+
				"Source TON: %v\n    Source NPI: %v\n    Source: %s\n    "+
				"Destination TON: %v\n    Destination NPI: %v\n    "+
				"Destination: %s\n    ESM Class: %d\n    Coding: %v\n    "+
				"Message: %s\n    Message ID: %s\n",
			dirStr,
			req.Command,
			req.Status,
			req.Sequence,
			req.Source.TON,
			req.Source.NPI,
			req.Source.Addr,
			req.Destination.TON,
			req.Destination.NPI,
			req.Destination.Addr,
			req.EsmClass,
			req.Coding,
			req.Message,
			req.MessageID,
		)
	case *sender.GenericPDU:
		log = fmt.Sprintf(
			"%s\n    Command: %v\n    Status: %v\n   Sequence: %d\n",
			dirStr,
			req.Command,
			req.Status,
			req.Sequence,
		)
	}

	glib.IdleAdd(func() {
		buf, err := ctx.logsArea.GetBuffer()
		if err != nil {
			return
		}
		buf.InsertAtCursor(log)
		vadg := ctx.logsScroller.GetVAdjustment()
		vadg.SetValue(vadg.GetUpper())
	})
}

func (ctx *submitSmContext) sessionCloseHandler(err error) {
	ctx.session = nil
	ctx.submitSmForm.SetSensitive(false)
	if err != nil {
		errorDialog("SMPP session error: %v", err)
	}
}

func (ctx *submitSmContext) getRequest() *sender.Request {
	req := &sender.Request{}

	var ok bool
	isValid := true
	req.Source.TON = ctx.getTON(ctx.srcTonSelector)
	req.Source.NPI = ctx.getNPI(ctx.srcNpiSelector)
	req.Source.Addr, ok = checkEntryPresence(ctx.srcAddrEntry, "Source Address")
	isValid = isValid && ok

	req.Destination.TON = ctx.getTON(ctx.dstTonSelector)
	req.Destination.NPI = ctx.getNPI(ctx.dstNpiSelector)
	req.Destination.Addr, ok = checkEntryPresence(ctx.dstAddrEntry, "Destination Address")
	isValid = isValid && ok
	req.ValidityPeriod, _ = ctx.validityEntry.GetText()
	req.EffectiveCoding = ctx.effectiveCoding
	req.DeceptiveCoding = ctx.getDeceptiveCoding()
	req.RegisteredDelivery = ctx.getRegisteredDelivery()
	req.SplitMode = ctx.getSplitMode()

	msgBuf, _ := ctx.messageEntry.GetBuffer()
	msgStart, msgEnd := msgBuf.GetBounds()
	req.Message, _ = msgBuf.GetText(msgStart, msgEnd, true)
	if len(req.Message) == 0 {
		markInvalidEntry(&ctx.messageEntry.Widget, "Message must be set")
		isValid = false
	}

	segmentBytesU64, ok := checkEntryNumerical(ctx.segmentBytesEntry, 8, "Bytes per segment")
	isValid = isValid && ok
	req.BytePerSegment = int(segmentBytesU64)

	if isValid {
		return req
	}

	return nil
}

func (ctx *submitSmContext) resetStyles() {
	widgets := []*gtk.Widget{
		&ctx.srcAddrEntry.Widget,
		&ctx.dstAddrEntry.Widget,
		&ctx.segmentBytesEntry.Widget,
	}
	for _, widget := range widgets {
		widget.SetTooltipText("")
		ctx, _ := widget.GetStyleContext()
		ctx.RemoveClass("invalid-entry")
	}
}

func (ctx *submitSmContext) getSplitMode() sender.SplitMode {
	for _, splt := range ctx.spltRadios {
		if splt.btn.GetActive() {
			return splt.mode
		}
	}
	return sender.SplitUDH
}

func (ctx *submitSmContext) getRegisteredDelivery() sender.RegisteredDelivery {
	rd := sender.RdNotRequested

	if ctx.rdReqCheck.GetActive() {
		rd |= sender.RdRequested
	}

	if ctx.rdFailCheck.GetActive() {
		rd |= sender.RdOnFailure
	}

	if ctx.rdInterCheck.GetActive() {
		rd |= sender.RdIntermediate
	}

	return rd
}

func (ctx *submitSmContext) getTON(c *gtk.ComboBox) sender.TON {
	idx := getComboIndex(c)
	return submitSmTons[idx]
}

func (ctx *submitSmContext) getNPI(c *gtk.ComboBox) sender.NPI {
	idx := getComboIndex(c)
	return submitSmNpis[idx]
}

func (ctx *submitSmContext) getDeceptiveCoding() coding.Coding {
	idx := getComboIndex(ctx.decCodingSelector)
	return coding.All[idx]
}
