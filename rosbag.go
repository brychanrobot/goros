package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"log"
	"os"
)

func parseRecordHeader(reader *bufio.Reader) {
	var headerLength int32
	binary.Read(reader, binary.LittleEndian, &headerLength)
	//oneByte, err := reader.ReadByte()
	log.Println(headerLength)

	var fieldLength int32

	binary.Read(reader, binary.LittleEndian, &fieldLength)
	fieldName, _ := reader.ReadString('=')
	log.Println(fieldName)

}

func parseRosbag(path string) {
	log.Println(path)
	file, err := os.Open(path) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	line, _ := reader.ReadString('\n')
	log.Print(line)

	file.Seek(4096, 0) //offsets by the size of the chunk header

	reader = bufio.NewReader(file)

	parseRecordHeader(reader)
	/*chunkHeader := make([]byte, 4096)
	_, _ = file.Read(chunkHeader)
	log.Println(string(bytes))
	*/
}

func main() {
	flag.Parse()
	parseRosbag(flag.Arg(0))
}
