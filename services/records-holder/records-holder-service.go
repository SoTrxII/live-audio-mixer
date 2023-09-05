package records_holder

import (
	"fmt"
	rt_wav_encoder "live-audio-mixer/internal/rt-wav-encoder"
	stream_handler "live-audio-mixer/internal/stream-handler"
	"live-audio-mixer/pkg/recorder"
	pb "live-audio-mixer/proto"
	"os"
	"path/filepath"
)

const (
	baseDir = "./rec/"
	dstName = "rec.wav"
)

type RecordsHolder struct {
	records map[string]*Record
}

type Record struct {
	rec *recorder.Recorder
	dir string
	dst *os.File
}

func NewRecordsHolder() *RecordsHolder {
	return &RecordsHolder{
		records: map[string]*Record{},
	}
}

func (rh *RecordsHolder) Record(id string) error {
	if rh.hasRecord(id) {
		return fmt.Errorf("record with id %s already exists", id)
	}
	dir := filepath.Join(baseDir, id)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	dst, err := os.Create(filepath.Join(dir, dstName))
	if err != nil {
		return err
	}

	rh.records[id] = &Record{
		rec: recorder.NewRecorder(stream_handler.NewHandler(), rt_wav_encoder.Encode),
		dir: dir,
		dst: dst,
	}
	rh.records[id].rec.Start(dst)
	return nil
}

func (rh *RecordsHolder) Stop(id string) error {
	if !rh.hasRecord(id) {
		return fmt.Errorf("Record with id %s does not exist", id)
	}
	record := rh.records[id]
	record.rec.Stop()
	err := record.dst.Close()
	if err != nil {
		return err
	}
	delete(rh.records, id)
	return nil
	// TODO : Upload to object storage
}

func (rh *RecordsHolder) Update(event *pb.Event) error {
	if !rh.hasRecord(event.RecordId) {
		return fmt.Errorf("Record with id %s does not exist", event.RecordId)
	}
	rh.records[event.RecordId].rec.Update(event)
	return nil
}

func (rh *RecordsHolder) hasRecord(id string) bool {
	_, ok := rh.records[id]
	return ok
}
