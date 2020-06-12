package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"strconv"

	"github.com/deepch/vdk/av"
)

//Config global
var Config = loadConfig()

type Camera struct {
	Id   uint32 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

//ConfigST struct
type ConfigST struct {
	Server  ServerST            `json:"server"`
	Streams map[string]StreamST `json:"streams"`
}

//ServerST struct
type ServerST struct {
	HTTPPort string `json:"http_port"`
}

//StreamST struct
type StreamST struct {
	URL    string `json:"url"`
	Status bool   `json:"status"`
	Codecs []av.CodecData
	Cl     map[string]viwer
}
type viwer struct {
	c chan av.Packet
}

func loadConfig() *ConfigST {

	db, err := connectPg(getPgParam())
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Pg connected")
	}

	// var tmp ConfigST
	// data, err := ioutil.ReadFile("config.json")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// err = json.Unmarshal(data, &tmp)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	tmp := ConfigST{
		Server:  ServerST{HTTPPort: ":8083"},
		Streams: make(map[string]StreamST),
	}

	var cameras []Camera

	err = db.Select(&cameras, `select id, "name", url from smt_define_camera`)
	if err != nil {
		log.Println(err)
	}

	for _, c := range cameras {
		cl := make(map[string]viwer)
		tmp.Streams[strconv.Itoa(int(c.Id))] = StreamST{
			URL: c.Url,
			Cl:  cl,
		}
	}
	return &tmp
}

func (element *ConfigST) cast(uuid string, pck av.Packet) {
	for _, v := range element.Streams[uuid].Cl {
		if len(v.c) < cap(v.c) {
			v.c <- pck
		}
	}
}

func (element *ConfigST) ext(suuid string) bool {
	_, ok := element.Streams[suuid]
	return ok
}

func (element *ConfigST) coAd(suuid string, codecs []av.CodecData) {
	t := element.Streams[suuid]
	t.Codecs = codecs
	element.Streams[suuid] = t
}

func (element *ConfigST) coGe(suuid string) []av.CodecData {
	return element.Streams[suuid].Codecs
}

func (element *ConfigST) clAd(suuid string) (string, chan av.Packet) {
	cuuid := pseudoUUID()
	ch := make(chan av.Packet, 100)
	element.Streams[suuid].Cl[cuuid] = viwer{c: ch}
	return cuuid, ch
}

func (element *ConfigST) list() (string, []string) {
	var res []string
	var fist string
	for k := range element.Streams {
		if fist == "" {
			fist = k
		}
		res = append(res, k)
	}
	return fist, res
}
func (element *ConfigST) clDe(suuid, cuuid string) {
	delete(element.Streams[suuid].Cl, cuuid)
}

func pseudoUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}
