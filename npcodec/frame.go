package npcodec

import (
	"bufio"

	"sigmaos/fcall"
	"sigmaos/frame"
	np "sigmaos/sigmap"
)

type FcallWireCompat struct {
	Type fcall.Tfcall
	Tag  np.Ttag
	Msg  np.Tmsg
}

func ToInternal(fcallWC *FcallWireCompat) *np.FcallMsg {
	fm := np.MakeFcallMsgNull()
	fm.Fc.Type = uint32(fcallWC.Type)
	fm.Fc.Tag = uint32(fcallWC.Tag)
	fm.Fc.Session = uint64(fcall.NoSession)
	fm.Fc.Seqno = uint64(np.NoSeqno)
	fm.Msg = fcallWC.Msg
	return fm
}

func ToWireCompatible(fm *np.FcallMsg) *FcallWireCompat {
	fcallWC := &FcallWireCompat{}
	fcallWC.Type = fcall.Tfcall(fm.Fc.Type)
	fcallWC.Tag = np.Ttag(fm.Fc.Tag)
	fcallWC.Msg = fm.Msg
	return fcallWC
}

func MarshalFcallMsg(fc fcall.Fcall, b *bufio.Writer) *fcall.Err {
	fcm := fc.(*np.FcallMsg)
	f, error := marshal1(true, ToWireCompatible(fcm))
	if error != nil {
		return fcall.MkErr(fcall.TErrBadFcall, error.Error())
	}
	dataBuf := false
	var data []byte
	switch fcm.Type() {
	case fcall.TTwrite:
		msg := fcm.Msg.(*np.Twrite)
		data = msg.Data
		dataBuf = true
	case fcall.TTwriteV:
		msg := fcm.Msg.(*np.TwriteV)
		data = msg.Data
		dataBuf = true
	case fcall.TRread:
		msg := fcm.Msg.(*np.Rread)
		data = msg.Data
		dataBuf = true
	case fcall.TTsetfile:
		msg := fcm.Msg.(*np.Tsetfile)
		data = msg.Data
		dataBuf = true
	case fcall.TTputfile:
		msg := fcm.Msg.(*np.Tputfile)
		data = msg.Data
		dataBuf = true
	case fcall.TTwriteread:
		msg := fcm.Msg.(*np.Twriteread)
		data = msg.Data
		dataBuf = true
	case fcall.TRwriteread:
		msg := fcm.Msg.(*np.Rread)
		data = msg.Data
		dataBuf = true
	default:
	}
	if dataBuf {
		return frame.WriteFrameAndBuf(b, f, data)
	} else {
		return frame.WriteFrame(b, f)
	}
}

func UnmarshalFcallWireCompat(frame []byte) (fcall.Fcall, *fcall.Err) {
	fcallWC := &FcallWireCompat{}
	if err := unmarshal(frame, fcallWC); err != nil {
		return nil, fcall.MkErr(fcall.TErrBadFcall, err)
	}
	return ToInternal(fcallWC), nil
}
