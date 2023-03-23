# msc-osc

receives msc (sysex) midi messages from an etc express and sends it out as an osc message. I made ths just to control scenes in QLC+ with the etc express. Tested on Windows and MacOS

## output msc message format:
```
address = /msc/<command>/<cue number>
message = <cue number> true <command>
```

## example:
midi input: `F0 7F 01 02 01 01 32 36 35 00 31 00 F7`  (cue 265 A/B fader)

osc output: /msc/go/265 ,iTs 265 true go