// Copyright (c) 2015 Datacratic. All rights reserved.

package trace

import (
	"time"

	"github.com/datacratic/gometrics/metric"
)

// Metrics creates metrics from the trace of events.
type Metrics struct {
	Prefix string
	metric.Summary
	metric.Reporter
}

// HandleTrace updates the summary of metrics from the captured trace.
// Metrics are named after their context.
// It keeps track of the number of occurences when entering and leaving a context.
// The duration is also evaluated for each exit point.
func (h *Metrics) HandleTrace(events []Event) {
	path := make([]string, len(events))

	for i, n := 1, len(events); i < n; i++ {
		item := &events[i]
		from := &events[item.From]

		switch item.Kind {
		case CountEvent:
			h.Summary.Count(path[item.From]+item.What, item.Data)
		case SetEvent:
			h.Summary.Set(path[item.From]+item.What, item.Data)
		case RecordEvent:
			h.Summary.Record(path[item.From]+item.What, item.Data)
		case LogEvent:
			h.Summary.Log(path[item.From]+item.What, item.Data)
		case StartEvent:
			path[i] = item.What + "."
			h.Summary.Count(item.What+".Count", 1)
		case EnterEvent:
			name := path[item.From] + item.What + "."
			path[i] = name
			h.Summary.Count(name+"Count", 1)
		case LeaveEvent:
			name := path[item.From] + item.What
			h.Summary.Count(name+".Count", 1)

			ns := int64(item.When) - int64(from.When)
			dt := time.Duration(ns)
			h.Summary.Record(name+".Latency", dt)
		}
	}
}

func (h *Metrics) Report(dt time.Duration) {
	if h.Reporter == nil {
		return
	}

	h.Summary.Name = h.Prefix
	h.Summary.Time = time.Now().UTC()
	h.Summary.Step = dt
	h.Summary.Write(h.Reporter)
	h.Summary.Reset()
}

func (h *Metrics) Close() {
	h.Report(time.Since(h.Summary.Time))
}
