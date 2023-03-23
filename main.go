package main

import (
	"fmt"
	"time"

	"github.com/hypebeast/go-osc/osc"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

func main() {
	fmt.Println("hello world")

	client := osc.NewClient("127.0.0.1", 8765)
	msg := osc.NewMessage("/msc/cue/111")
	// msg.Append(int32(111))
	// msg.Append(true)
	// msg.Append("hello")
	client.Send(msg)

	msg = osc.NewMessage("/msc/cue/112")
	// msg.Append(int32(112))
	// msg.Append(true)
	// msg.Append("hello")
	client.Send(msg)

	msg = osc.NewMessage("/msc/cue/113")
	// msg.Append(int32(113))
	// msg.Append(true)
	// msg.Append("hello")
	client.Send(msg)

	defer midi.CloseDriver()

	in, err := midi.FindInPort("Keyboard")
	if err != nil {
		fmt.Println("can't find VMPK")
		return
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
		case msg.GetNoteEnd(&ch, &key):
			fmt.Printf("ending note %s on channel %v\n", midi.Note(key), ch)
		default:
			// ignore
		}
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	time.Sleep(time.Second * 5)

	stop()
}
