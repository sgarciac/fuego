package main

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"io"
	"log"
)

// wrapper to stream the json serialized results
type displayItemWriter struct {
	isFirst bool
	writer  *io.Writer
}

func newDisplayItemWriter(writer *io.Writer) displayItemWriter {
	return displayItemWriter{true, writer}
}

func (d *displayItemWriter) Write(doc *firestore.DocumentSnapshot, extendedJson bool) error {
	if d.isFirst {
		_, err := fmt.Fprintln(*d.writer, "[")
		if err != nil {
			return err
		}
		d.isFirst = false
	} else {
		_, err := fmt.Fprintln(*d.writer, ",")
		if err != nil {
			return err
		}
	}

	return writeSnapshot(*d.writer, doc, extendedJson)
}

func (d *displayItemWriter) Close() {
	if !d.isFirst {
		_, err := fmt.Fprintln(*d.writer, "]")
		if err != nil {
			log.Panicf("Could not write finishing part of results. %v", err)
		}
	}
}

func writeSnapshot(writer io.Writer, doc *firestore.DocumentSnapshot, extendedJson bool) error {
	var displayItem = make(map[string]interface{})

	displayItem["ID"] = doc.Ref.ID
	displayItem["CreateTime"] = doc.CreateTime
	displayItem["ReadTime"] = doc.ReadTime
	displayItem["UpdateTime"] = doc.UpdateTime
	displayItem["Data"] = doc.Data()

	jsonString, err := marshallData(displayItem, extendedJson)

	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, jsonString)

	if err != nil {
		return err
	}
	return nil
}
