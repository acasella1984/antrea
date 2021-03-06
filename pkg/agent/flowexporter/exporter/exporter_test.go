// Copyright 2020 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	ipfixentities "github.com/vmware/go-ipfix/pkg/entities"
	ipfixregistry "github.com/vmware/go-ipfix/pkg/registry"

	"github.com/vmware-tanzu/antrea/pkg/agent/flowexporter"
	ipfixtest "github.com/vmware-tanzu/antrea/pkg/agent/flowexporter/ipfix/testing"
)

const (
	testTemplateID          = 256
	testFlowExportFrequency = 12
)

func TestFlowExporter_sendTemplateRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIPFIXExpProc := ipfixtest.NewMockIPFIXExportingProcess(ctrl)
	mockTempRec := ipfixtest.NewMockIPFIXRecord(ctrl)
	mockIPFIXRegistry := ipfixtest.NewMockIPFIXRegistry(ctrl)
	flowExp := &flowExporter{
		nil,
		mockIPFIXExpProc,
		nil,
		testFlowExportFrequency,
		0,
		testTemplateID,
		mockIPFIXRegistry,
	}
	// Following consists of all elements that are in IANAInfoElements and AntreaInfoElements (globals)
	// Only the element name is needed, other arguments have dummy values.
	elemList := make([]*ipfixentities.InfoElement, 0)
	for _, ie := range IANAInfoElements {
		elemList = append(elemList, ipfixentities.NewInfoElement(ie, 0, 0, ipfixregistry.IANAEnterpriseID, 0))
	}
	for _, ie := range IANAReverseInfoElements {
		elemList = append(elemList, ipfixentities.NewInfoElement(ie, 0, 0, ipfixregistry.ReverseEnterpriseID, 0))
	}
	for _, ie := range AntreaInfoElements {
		elemList = append(elemList, ipfixentities.NewInfoElement(ie, 0, 0, ipfixregistry.AntreaEnterpriseID, 0))
	}
	// Expect calls for different mock objects
	tempBytes := uint16(0)
	var templateRecord ipfixentities.Record

	mockTempRec.EXPECT().PrepareRecord().Return(tempBytes, nil)
	for i, ie := range IANAInfoElements {
		mockIPFIXRegistry.EXPECT().GetInfoElement(ie, ipfixregistry.IANAEnterpriseID).Return(elemList[i], nil)
		mockTempRec.EXPECT().AddInfoElement(elemList[i], nil).Return(tempBytes, nil)
	}
	for i, ie := range IANAReverseInfoElements {
		mockIPFIXRegistry.EXPECT().GetInfoElement(ie, ipfixregistry.ReverseEnterpriseID).Return(elemList[i+len(IANAInfoElements)], nil)
		mockTempRec.EXPECT().AddInfoElement(elemList[i+len(IANAInfoElements)], nil).Return(tempBytes, nil)
	}
	for i, ie := range AntreaInfoElements {
		mockIPFIXRegistry.EXPECT().GetInfoElement(ie, ipfixregistry.AntreaEnterpriseID).Return(elemList[i+len(IANAInfoElements)+len(IANAReverseInfoElements)], nil)
		mockTempRec.EXPECT().AddInfoElement(elemList[i+len(IANAInfoElements)+len(IANAReverseInfoElements)], nil).Return(tempBytes, nil)
	}
	mockTempRec.EXPECT().GetRecord().Return(templateRecord)
	mockTempRec.EXPECT().GetTemplateElements().Return(elemList)
	// Passing 0 for sentBytes as it is not used anywhere in the test. If this not a call to mock, the actual sentBytes
	// above elements: IANAInfoElements, IANAReverseInfoElements and AntreaInfoElements.
	mockIPFIXExpProc.EXPECT().AddRecordAndSendMsg(ipfixentities.Template, templateRecord).Return(0, nil)

	_, err := flowExp.sendTemplateRecord(mockTempRec)
	if err != nil {
		t.Errorf("Error in sending templated record: %v", err)
	}

	assert.Equal(t, len(IANAInfoElements)+len(IANAReverseInfoElements)+len(AntreaInfoElements), len(flowExp.elementsList), flowExp.elementsList, "flowExp.elementsList and template record should have same number of elements")
}

// TestFlowExporter_sendDataRecord tests essentially if element names in the switch-case matches globals
// IANAInfoElements and AntreaInfoElements.
func TestFlowExporter_sendDataRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Values in the connection are not important. Initializing with 0s.
	flow1 := flowexporter.Connection{
		StartTime:       time.Time{},
		StopTime:        time.Time{},
		OriginalPackets: 0,
		OriginalBytes:   0,
		ReversePackets:  0,
		ReverseBytes:    0,
		TupleOrig: flowexporter.Tuple{
			SourceAddress:      nil,
			DestinationAddress: nil,
			Protocol:           0,
			SourcePort:         0,
			DestinationPort:    0,
		},
		TupleReply: flowexporter.Tuple{
			SourceAddress:      nil,
			DestinationAddress: nil,
			Protocol:           0,
			SourcePort:         0,
			DestinationPort:    0,
		},
		SourcePodNamespace:      "",
		SourcePodName:           "",
		DestinationPodNamespace: "",
		DestinationPodName:      "",
	}
	record1 := flowexporter.FlowRecord{
		Conn:               &flow1,
		PrevPackets:        0,
		PrevBytes:          0,
		PrevReversePackets: 0,
		PrevReverseBytes:   0,
	}
	// Following consists of all elements that are in IANAInfoElements and AntreaInfoElements (globals)
	// Need only element name and other are dummys
	elemList := make([]*ipfixentities.InfoElement, len(IANAInfoElements)+len(IANAReverseInfoElements)+len(AntreaInfoElements))
	for i, ie := range IANAInfoElements {
		elemList[i] = ipfixentities.NewInfoElement(ie, 0, 0, 0, 0)
	}
	for i, ie := range IANAReverseInfoElements {
		elemList[i+len(IANAInfoElements)] = ipfixentities.NewInfoElement(ie, 0, 0, ipfixregistry.ReverseEnterpriseID, 0)
	}
	for i, ie := range AntreaInfoElements {
		elemList[i+len(IANAInfoElements)+len(IANAReverseInfoElements)] = ipfixentities.NewInfoElement(ie, 0, 0, 0, 0)
	}

	mockIPFIXExpProc := ipfixtest.NewMockIPFIXExportingProcess(ctrl)
	mockDataRec := ipfixtest.NewMockIPFIXRecord(ctrl)
	mockIPFIXRegistry := ipfixtest.NewMockIPFIXRegistry(ctrl)
	flowExp := &flowExporter{
		nil,
		mockIPFIXExpProc,
		elemList,
		testFlowExportFrequency,
		0,
		testTemplateID,
		mockIPFIXRegistry,
	}
	// Expect calls required
	var dataRecord ipfixentities.Record
	tempBytes := uint16(0)
	for _, ie := range flowExp.elementsList {
		switch ieName := ie.Name; ieName {
		case "flowStartSeconds", "flowEndSeconds":
			mockDataRec.EXPECT().AddInfoElement(ie, time.Time{}.Unix()).Return(tempBytes, nil)
		case "sourceIPv4Address", "destinationIPv4Address":
			mockDataRec.EXPECT().AddInfoElement(ie, nil).Return(tempBytes, nil)
		case "destinationClusterIP":
			mockDataRec.EXPECT().AddInfoElement(ie, net.IP{0, 0, 0, 0}).Return(tempBytes, nil)
		case "sourceTransportPort", "destinationTransportPort":
			mockDataRec.EXPECT().AddInfoElement(ie, uint16(0)).Return(tempBytes, nil)
		case "protocolIdentifier":
			mockDataRec.EXPECT().AddInfoElement(ie, uint8(0)).Return(tempBytes, nil)
		case "packetTotalCount", "octetTotalCount", "packetDeltaCount", "octetDeltaCount", "reverse_PacketTotalCount", "reverse_OctetTotalCount", "reverse_PacketDeltaCount", "reverse_OctetDeltaCount":
			mockDataRec.EXPECT().AddInfoElement(ie, uint64(0)).Return(tempBytes, nil)
		case "sourcePodName", "sourcePodNamespace", "sourceNodeName", "destinationPodName", "destinationPodNamespace", "destinationNodeName", "destinationServicePortName":
			mockDataRec.EXPECT().AddInfoElement(ie, "").Return(tempBytes, nil)
		}
	}
	mockDataRec.EXPECT().GetRecord().Return(dataRecord)
	mockIPFIXExpProc.EXPECT().AddRecordAndSendMsg(ipfixentities.Data, dataRecord).Return(0, nil)

	err := flowExp.sendDataRecord(mockDataRec, record1)
	if err != nil {
		t.Errorf("Error in sending data record: %v", err)
	}
}
