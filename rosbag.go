package goros

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	MessageData      = 0x02
	BagHeader        = 0x03
	IndexData        = 0x04
	ChunkRecord      = 0x05
	ChunkInfo        = 0x06
	ConnectionRecord = 0x07
)

var (
	typeMap = map[string]reflect.Type{
		"bool":    reflect.TypeOf((*bool)(nil)).Elem(),
		"float32": reflect.TypeOf((*float32)(nil)).Elem(),
		"float64": reflect.TypeOf((*float64)(nil)).Elem(),
		"uint32":  reflect.TypeOf((*uint32)(nil)).Elem(),
		"uint64":  reflect.TypeOf((*uint32)(nil)).Elem(),
		"int32":   reflect.TypeOf((*int32)(nil)).Elem(),
		"int64":   reflect.TypeOf((*int64)(nil)).Elem(),
	}
)

func parseSubMessageDefinition(subMessage string, superMessageTypeString string) {
	commentRegexp, _ := regexp.Compile("\\s*#[^\n]*")
	subMessageWithoutComments := commentRegexp.ReplaceAllString(subMessage, "")
	subMessageMap := make(map[string]reflect.Type)
	var subMessageTypeString string
	for _, line := range strings.Split(subMessageWithoutComments, "\n") {
		if len(line) != 0 {
			if strings.HasPrefix(line, "MSG") {
				subMessageTypeString = strings.TrimPrefix(line, "MSG: ")
				continue
			}

			split := strings.Split(line, " ")
			typeString := split[0]
			fieldName := split[1]

			var fieldType reflect.Type
			if strings.HasSuffix(typeString, "[]") {
				elementType := typeMap[typeString[:len(typeString)-2]]
				fieldType = reflect.SliceOf(elementType)
			} else {
				fieldType = typeMap[typeString]
			}

			subMessageMap[fieldName] = fieldType
			//fmt.Println(line)
		}
	}

	//return subMessageTypeString, subMessageMap
	if len(subMessageTypeString) == 0 {
		subMessageTypeString = superMessageTypeString
	}

	//typeString := strings.Split(subMessageTypeString, " ")[1]
	//typeMap[typeString] = subMessageMap
}

func parseMessageDefinition(dataMap map[string]interface{}) {
	//fmt.Println(dataMap["message_definition"])

	messageSplitterRegex, _ := regexp.Compile("[=]+\n")

	subMessages := messageSplitterRegex.Split(dataMap["message_definition"].(string), -1)

	//fmt.Println(subMessages)
	/*
		for _, line := range strings.Split(dataMap["message_definition"].(string), "\n") {
			fmt.Println(line)
		}
	*/
	for _, subMessage := range subMessages {
		fmt.Println(subMessage)
		//defer parseSubMessageDefinition(subMessage, dataMap["type"].(string)) //log.Println(subMessage)
	}
}

func parseRecordHeader(buffer []byte) map[string]interface{} {
	//log.Println(string(buffer))
	reader := bufio.NewReader(bytes.NewReader(buffer))
	//log.Println(reader.Buffered())

	valueMap := make(map[string]interface{})

	var fieldLength uint32
	for binary.Read(reader, binary.LittleEndian, &fieldLength) == nil {
		fieldName, _ := reader.ReadString('=')
		fieldName = fieldName[:len(fieldName)-1]
		//log.Println(fieldName)
		//reader.Discard(int(fieldLength) - len(fieldName))
		valueLength := int(fieldLength) - len(fieldName) - 1

		switch fieldName {
		case "op":
			value, _ := reader.ReadByte()
			valueMap[fieldName] = value
		case "conn", "size", "conn_count", "chunk_count", "count", "ver":
			var value uint32
			binary.Read(reader, binary.LittleEndian, &value)
			valueMap[fieldName] = value
		case "index_pos":
			var value uint64
			binary.Read(reader, binary.LittleEndian, &value)
			valueMap[fieldName] = value
		case "time", "start_time", "end_time", "chunk_pos":
			var secs int32
			var nsecs int32
			binary.Read(reader, binary.LittleEndian, &secs)
			binary.Read(reader, binary.LittleEndian, &nsecs)
			valueMap[fieldName] = time.Unix(int64(secs), int64(nsecs))
		case "compression", "topic":
			value := make([]byte, valueLength)
			reader.Read(value)
			valueMap[fieldName] = string(value)
		default:
			value := make([]byte, valueLength)
			reader.Read(value)
			valueMap[fieldName] = string(value)
			log.Printf("Congratulations! I don't know what to do with %s", fieldName)
		}
	}

	return valueMap
}

func parseRecord(reader *bufio.Reader) error {
	var headerLength uint32
	err := binary.Read(reader, binary.LittleEndian, &headerLength)
	if err != nil {
		return err
	}
	//log.Println(headerLength)
	buffer := make([]byte, int(headerLength))
	reader.Read(buffer)
	valueMap := parseRecordHeader(buffer)
	fmt.Println(valueMap)

	var dataLength uint32
	binary.Read(reader, binary.LittleEndian, &dataLength)
	//log.Println(dataLength)
	//if skip {
	//reader.Discard(int(dataLength))
	dataBuffer := make([]byte, int(dataLength))
	bytesRead, _ := io.ReadFull(reader, dataBuffer)
	//log.Println(bytesRead)

	if bytesRead != int(dataLength) {
		panic(fmt.Sprintf("expected %d bytes, got %d", dataLength, bytesRead))
	}
	/*} else {
		buffer := make([]byte, int(dataLength))
		reader.Read(buffer)
		log.Println(string(buffer))
	}
	*/
	switch valueMap["op"].(byte) {
	case BagHeader:
	case ChunkRecord:
		chunkReader := bufio.NewReaderSize(bytes.NewReader(dataBuffer), 8000000)
		var chunkErr error
		for chunkErr == nil {
			chunkErr = parseRecord(chunkReader)
		}
	case ConnectionRecord:
		//log.Println(string(dataBuffer))
		dataMap := parseRecordHeader(dataBuffer)
		//log.Println(dataMap)
		parseMessageDefinition(dataMap)
	case MessageData:
	case IndexData:
	case ChunkInfo:
	}

	return nil

}

func ParseRosbag(path string) {
	log.Println(path)
	file, err := os.Open(path) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 8000000)

	line, _ := reader.ReadString('\n')
	log.Print(line)

	//parseRecord(reader)
	//parseRecord(reader)
	var recordError error
	for recordError == nil {
		recordError = parseRecord(reader)
	}
	/*chunkHeader := make([]byte, 4096)
	_, _ = file.Read(chunkHeader)
	log.Println(string(buffer))
	*/
}
