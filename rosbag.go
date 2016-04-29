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
	"strconv"
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
	typeMap = map[string]interface{}{
		"bool":    reflect.TypeOf((*bool)(nil)).Elem(),
		"float32": reflect.TypeOf((*float32)(nil)).Elem(),
		"float64": reflect.TypeOf((*float64)(nil)).Elem(),
		"uint32":  reflect.TypeOf((*uint32)(nil)).Elem(),
		"uint64":  reflect.TypeOf((*uint32)(nil)).Elem(),
		"int32":   reflect.TypeOf((*int32)(nil)).Elem(),
		"int64":   reflect.TypeOf((*int64)(nil)).Elem(),
		"uint8":   reflect.TypeOf((*byte)(nil)).Elem(),
		"time":    reflect.TypeOf((*RosTime)(nil)).Elem(),
		"string":  reflect.TypeOf((*string)(nil)).Elem(),
	}
)

type CustomMessage map[string]interface{}
type CustomMessageSlice struct {
	CustomMessage
}

type RosTime struct {
	Secs  int32
	NSecs int32
}

func parseSubMessageDefinition(subMessage string, superMessageTypeString string) {
	commentRegexp, _ := regexp.Compile("\\s*#[^\n]*")
	subMessageWithoutComments := commentRegexp.ReplaceAllString(subMessage, "")
	subMessageMap := make(CustomMessage)
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

			if strings.HasSuffix(typeString, "[]") {
				elementType := typeMap[typeString[:len(typeString)-2]]
				switch elementType.(type) {
				case reflect.Type:
					subMessageMap[fieldName] = reflect.SliceOf(elementType.(reflect.Type))
				case CustomMessage:
					subMessageMap[fieldName] = elementType
				}
			} else {
				subMessageMap[fieldName] = typeMap[typeString]
			}
		}
	}

	//return subMessageTypeString, subMessageMap
	if len(subMessageTypeString) == 0 {
		subMessageTypeString = superMessageTypeString
	}

	typeString := strings.Split(subMessageTypeString, "/")[1]
	typeMap[typeString] = subMessageMap

	log.Printf("%s, %s", typeString, subMessageMap)
}

func parseMessageDefinition(dataMap map[string]interface{}) {
	//fmt.Println(dataMap)

	messageSplitterRegex, _ := regexp.Compile("[=]+\n")

	subMessages := messageSplitterRegex.Split(dataMap["message_definition"].(string), -1)

	//fmt.Println(subMessages)
	/*
		for _, line := range strings.Split(dataMap["message_definition"].(string), "\n") {
			fmt.Println(line)
		}
	*/
	for _, subMessage := range subMessages {
		//defer fmt.Println(subMessage)
		defer parseSubMessageDefinition(subMessage, dataMap["type"].(string)) //log.Println(subMessage)
	}
}

func parseRecordHeader(buffer []byte) map[string]interface{} {
	//log.Println(string(buffer))
	reader := bufio.NewReader(bytes.NewReader(buffer))
	//log.Println(reader.Buffered())
	valueMap := make(map[string]interface{})

	var fieldLength uint32
	for binary.Read(reader, binary.LittleEndian, &fieldLength) == nil {
		fieldName, err := reader.ReadString('=')
		fieldName = fieldName[:len(fieldName)-1]
		//log.Println(fieldName)
		//reader.Discard(int(fieldLength) - len(fieldName))
		valueLength := int(fieldLength) - len(fieldName) - 1

		switch fieldName {
		case "op":
			//var value byte
			//binary.Read(reader, binary.LittleEndian, value)
			value, _ := reader.ReadByte()
			valueMap[fieldName] = value
		case "conn", "size", "conn_count", "chunk_count", "count", "ver":
			var value uint32
			err = binary.Read(reader, binary.LittleEndian, &value)
			valueMap[fieldName] = value
		case "index_pos":
			var value uint64
			err = binary.Read(reader, binary.LittleEndian, &value)
			valueMap[fieldName] = value
		case "time", "start_time", "end_time", "chunk_pos":
			var timestamp RosTime
			err = binary.Read(reader, binary.LittleEndian, &timestamp)
			valueMap[fieldName] = time.Unix(int64(timestamp.Secs), int64(timestamp.NSecs))
		case "compression", "topic", "callerid", "type", "md5sum", "message_definition":
			value := make([]byte, valueLength)
			_, err = io.ReadFull(reader, value)
			valueMap[fieldName] = string(value)
		case "latching":
			value := make([]byte, valueLength)
			_, err = io.ReadFull(reader, value)
			intVal, _ := strconv.Atoi(string(value))
			valueMap[fieldName] = intVal
		default:
			value := make([]byte, valueLength)
			_, err = io.ReadFull(reader, value)
			valueMap[fieldName] = string(value)
			log.Printf("Congratulations! I don't know what to do with %s", fieldName)
		}

		if err != nil {
			panic(err)
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
	if headerLength == 0 {
		panic(fmt.Sprintf("cannot read header of size %d", headerLength))
	}
	//log.Println(headerLength)
	buffer := make([]byte, int(headerLength))
	io.ReadFull(reader, buffer)
	valueMap := parseRecordHeader(buffer)
	fmt.Println(valueMap)

	var dataLength uint32
	binary.Read(reader, binary.LittleEndian, &dataLength)
	//log.Println(dataLength)
	//if skip {
	//reader.Discard(int(dataLength))
	var dataBuffer []byte
	if valueMap["op"] != MessageData {
		dataBuffer = make([]byte, int(dataLength))
		bytesRead, _ := io.ReadFull(reader, dataBuffer)
		//log.Println(bytesRead)

		if bytesRead != int(dataLength) {
			panic(fmt.Sprintf("expected %d bytes, got %d", dataLength, bytesRead))
		}
	} else {
		reader.Discard(int(dataLength))
	}
	/*} else {
		buffer := make([]byte, int(dataLength))
		reader.Read(buffer)
		log.Println(string(buffer))
	}
	*/
	/*
		switch valueMap["op"].(byte) {
		case BagHeader:
		case ChunkRecord:
			chunkReader := bufio.NewReader(bytes.NewReader(dataBuffer))
			var chunkErr error
			for chunkErr == nil {
				chunkErr = parseRecord(chunkReader)
			}
		case ConnectionRecord:
			fmt.Println(valueMap)
			//log.Println(string(dataBuffer))
			dataMap := parseRecordHeader(dataBuffer)
			//log.Println(dataMap)
			parseMessageDefinition(dataMap)
		case MessageData:
			//fmt.Println(valueMap)
		case IndexData:
		case ChunkInfo:
		}
	*/

	return nil

}

func ParseRosbag(path string) {
	log.Println(path)
	file, err := os.Open(path) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

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
