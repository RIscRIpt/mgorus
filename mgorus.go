package mgorus

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Hook is a wrapper around mgo.Collection, which implements
// the Fire method, so it can be used as a logrus.Hook
type Hook struct {
	collection *mgo.Collection
	origin     string
}

type doc struct {
	Id      bson.ObjectId `bson:"_id,omitempty"`
	Level   string        `bson:"level"`
	Origin  string        `bson:"origin,omitempty"`
	Message string        `bson:"message"`
	Data    bson.M        `bson:"data"`
}

// New returns a ready-to-use logrus.Hook
//
// Example:
//     logrus.AddHook(mgorus.New("system", mgo.Database("test").C("logs")))
func New(origin string, collection *mgo.Collection) *Hook {
	return &Hook{
		collection: collection,
		origin:     origin,
	}
}

func timedObjectId(t time.Time) bson.ObjectId {
	id := []byte(bson.NewObjectId())
	binary.BigEndian.PutUint32(id, uint32(t.Unix()))
	return bson.ObjectId(id)
}

// Fire implements one of logrus.Hook interface methods.
func (h *Hook) Fire(entry *logrus.Entry) error {
	doc := &doc{
		Id:      timedObjectId(entry.Time),
		Level:   entry.Level.String(),
		Origin:  h.origin,
		Message: entry.Message,
		Data:    bson.M(entry.Data),
	}
	if errData, ok := entry.Data[logrus.ErrorKey]; ok {
		if err, ok := errData.(error); ok {
			doc.Data[logrus.ErrorKey] = err.Error()
		} else {
			doc.Data[logrus.ErrorKey] = errData
		}
	}
	err := h.collection.Insert(doc)
	if err != nil {
		return fmt.Errorf("Failed to send log entry to mongodb: %s", err)
	}
	return nil
}

// Levels implements one of logrus.Hook interface methods.
// Returns a slice of levels, which represent a log level
// when this Hook must be triggered.
func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
