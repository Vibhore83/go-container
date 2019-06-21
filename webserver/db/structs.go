/*
 * Structures and functions defined in this file are used to store/fetch data in MongoDB
 *
 * API version: 1.0.0
 * Contact: Arun K
 */


package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

//ContainerProp is the container struct
type ContainerProp struct {
	Image    string `json:"image" bson:"image"`
	CID      string `json:"cid" bson:"cid"`
	HostName string `json:"hostname" bson:"hostname"`
	IP       string `json:"ip" bson:"ip"`
	SvcPort  int    `json:"svc_port" bson:"svc_port"`
	RestPort int    `json:"rest_port" bson:"rest_port"`
}

//TestBed is the test bed struct
type TestBed struct {
	ID        string          `json:"_id" bson:"_id"`
	CTS       int             `json:"_cts" bson:"_cts"`
	Name      string          `json:"name" bson:"name"`
	Container []ContainerProp `json:"container" bson:"container"`
	Status    string          `json:"status" bson:"status"`
}

// TestBedMeta is the TestBedMeta collection struct
type TestBedMeta struct {
	ID             string `json:"_id" bson:"_id"`
	AllocatedPorts []int  `json:"allocatedPorts,omitempty" bson:"allocatedPorts,omitempty"`
}

//NewTestBed creates a new TestBed
func NewTestBed() *TestBed {
	testbedID := uuid.New().String()
	return &TestBed{
		ID:     fmt.Sprintf("%v", testbedID),
		CTS:    int(time.Now().Unix()),
		Status: "initiated",
	}
}

// NewTestBedMeta creates a new TestBedMeta document
func NewTestBedMeta() *TestBedMeta {
	id := uuid.New().String()
	return &TestBedMeta{
		ID: fmt.Sprintf("%v", id),
	}
}
