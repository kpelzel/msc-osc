package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/hypebeast/go-osc/osc"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver

	yaml "gopkg.in/yaml.v3"
)

const (
	DefaultMidiIn     = "Keyboard"
	DefaultOSCOutIP   = "127.0.0.1"
	DefaultOSCOutPort = 8765
)

type MSCOSC struct {
	OSCClient *osc.Client
}

type conf struct {
	MidiIn     string `yaml:"midiIn"`
	OSCOutIP   net.IP `yaml:"oscOutIP"`
	OSCOutPort int    `yaml:"oscOutPort"`
}

func main() {
	defer midi.CloseDriver()

	fmt.Println("hello world")

	confBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	conf := &conf{}
	err = yaml.Unmarshal(confBytes, conf)
	if err != nil {
		log.Fatalf("failed to unmarshal config file: %v", err)
	}

	// setup osc client
	mscOSC := &MSCOSC{
		OSCClient: osc.NewClient(conf.OSCOutIP.String(), conf.OSCOutPort),
	}

	// connect to midi input
	in, err := midi.FindInPort(conf.MidiIn)
	if err != nil {
		fmt.Printf("can't find midi %v\n", conf.MidiIn)
		return
	}

	// listen for midi sysex commands from etc
	stop, err := midi.ListenTo(in, mscOSC.midiListenFunc, midi.UseSysEx())
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Printf("listening for midi from %v(%v) and outputting to %s:%d\n", in.String(), in.Number(), conf.OSCOutIP, conf.OSCOutPort)

	// listen for ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		// sig is a ^C, handle it
		fmt.Println("quitting")
		break
	}

	stop()
}

func (m *MSCOSC) midiListenFunc(msg midi.Message, timestampms int32) {
	var bt []byte
	var ch, key, vel uint8
	switch {
	case msg.GetSysEx(&bt):
		fmt.Printf("got sysex: % X\n", bt)
		command, cue, err := parseMSC(bt)
		if err != nil {
			fmt.Printf("failed to parse msc: %v\n", err)
		} else {
			m.sendOSC(command, cue)
		}
	case msg.GetNoteStart(&ch, &key, &vel):
		fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
	case msg.GetNoteEnd(&ch, &key):
		fmt.Printf("ending note %s on channel %v\n", midi.Note(key), ch)
	default:
		// ignore
	}
}

func parseMSC(bt []byte) (command string, cue string, err error) {
	if len(bt) >= 9 && bt[0] == 0x7f {
		// get cue number
		btLen := len(bt)
		cue = string(bt[5 : btLen-3])

		// get command
		command = ""
		switch bt[4] {
		case 0x01:
			command = "go"
		case 0x02:
			command = "stop"
		case 0x03:
			command = "resume"
		case 0x07:
			command = "macro"
		default:
			return "", "", fmt.Errorf("unrecognized msc command: %x", bt[4])
		}

		return command, cue, nil
	}

	return "", "", fmt.Errorf("not an msc packet. len: %v bt[0]: %x\n", len(bt), bt[0])
}

func (m *MSCOSC) sendOSC(command, cue string) {
	cueInt, err := strconv.Atoi(cue)
	if err != nil {
		fmt.Printf("failed to convert %v to int: %v\n", cue, err)
	} else {
		msg := osc.NewMessage(fmt.Sprintf("/msc/%s/%s", command, cue))
		msg.Append(int32(cueInt))
		msg.Append(true)
		msg.Append(command)
		fmt.Printf("sending %v\n", msg.String())
		m.OSCClient.Send(msg)
	}
}
