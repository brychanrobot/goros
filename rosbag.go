package goros

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

func parseRecordHeader(buffer []byte) map[string]interface{} {
	log.Println(string(buffer))
	reader := bufio.NewReader(bytes.NewReader(buffer))
	log.Println(reader.Buffered())

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
			opCode, _ := reader.ReadByte()
			valueMap[fieldName] = opCode
			//log.Printf("%s%d", fieldName, opCode)
		case "compression":
			value := make([]byte, valueLength)
			reader.Read(value)
			valueMap[fieldName] = string(value)
			//log.Printf("%s%s", fieldName, string(value))
		case "conn":
			var connectionId uint32
			binary.Read(reader, binary.LittleEndian, &connectionId)
			valueMap[fieldName] = connectionId
			//log.Printf("%s%d", fieldName, connectionId)
		case "size":
			var uncompressedSize uint32
			binary.Read(reader, binary.LittleEndian, &uncompressedSize)
			valueMap[fieldName] = uncompressedSize
			//log.Printf("%s%d", fieldName, uncompressedSize)
		default:
			value := make([]byte, valueLength)
			reader.Read(value)
			valueMap[fieldName] = string(value)
			log.Printf("Congratulations! I don't know what to do with %s", fieldName)
		}
	}

	return valueMap
}

func parseRecord(reader *bufio.Reader, skip bool) {
	var headerLength uint32
	binary.Read(reader, binary.LittleEndian, &headerLength)
	log.Println(headerLength)
	if skip {
		reader.Discard(int(headerLength))
	} else {
		buffer := make([]byte, int(headerLength))
		reader.Read(buffer)
		valueMap := parseRecordHeader(buffer)
		log.Println(valueMap)
		//log.Println(string(buffer))
	}

	var dataLength uint32
	binary.Read(reader, binary.LittleEndian, &dataLength)
	log.Println(dataLength)
	if skip {
		reader.Discard(int(dataLength))
	} else {
		buffer := make([]byte, int(dataLength))
		reader.Read(buffer)
		log.Println(string(buffer[:5000]))
	}

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

	parseRecord(reader, true)
	parseRecord(reader, false)
	/*chunkHeader := make([]byte, 4096)
	_, _ = file.Read(chunkHeader)
	log.Println(string(buffer))
	*/
}
