package records_holder

import (
	"fmt"
	rt_encoder "live-audio-mixer/internal/rt-encoder"
	stream_handler "live-audio-mixer/internal/stream-handler"
	"live-audio-mixer/pkg/recorder"
	pb "live-audio-mixer/proto"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	baseDir = "./rec/"
	dstName = "rec.wav"
)

type ObjectStorage interface {
	Upload(path string, id string) error
}
type RecordsHolder struct {
	records map[string]*Record
	store   ObjectStorage
}

type Record struct {
	rec *recorder.Recorder
	dir string
	dst *os.File
}

func NewRecordsHolder(store ObjectStorage) *RecordsHolder {
	return &RecordsHolder{
		records: map[string]*Record{},
		store:   store,
	}
}

func (rh *RecordsHolder) Record(id string) error {
	if rh.hasRecord(id) {
		return fmt.Errorf("record with id %s already exists", id)
	}
	absPath, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}
	dir := filepath.Join(absPath, id)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	dst, err := os.Create(filepath.Join(dir, dstName))
	if err != nil {
		return err
	}

	rh.records[id] = &Record{
		rec: recorder.NewRecorder(stream_handler.NewHandler(), rt_encoder.FFEncode),
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
	// Optionally, upload the file to the object storage
	if rh.store != nil {
		recordPath := filepath.Join(record.dir, dstName)
		err = rh.store.Upload(recordPath, fmt.Sprintf("%s.ogg", id))
		if err != nil {
			return err
		}
		err = os.RemoveAll(record.dir)
		if err != nil {
			slog.Warn(fmt.Sprintf("[RecordsHolder] :: Error while removing record dir %s : %v", record.dir, err))
		}
	}
	return nil
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
