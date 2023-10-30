# Live audio Mixer

![CI](https://github.com/SoTrxII/live-audio-mixer/actions/workflows/publish-coverage.yml/badge.svg)
[![codecov](https://codecov.io/gh/SoTrxII/live-audio-mixer/graph/badge.svg?token=E1YZKGK9IT)](https://codecov.io/gh/SoTrxII/live-audio-mixer)

## Description

This mixer allows the user to mix multiple audio sources in real time. 

## Usage

There are tree commands available, all of which must be called using the gRPC API:

```protobuf
  service EventStream {
  // Start a new record with an ID
  rpc Start(RecordRequest) returns (RecordReply);
  // Stop a record by ID
  rpc Stop(StopRequest) returns (StopReply);
  // Send new events to the mixer
  rpc StreamEvents(stream Event) returns (EventReply) ;
  }
```

![Sequence](./resources/images/sequence.png)

Once a new recording is started, the mixer will start streaming silence to the output and will accept new events for this record.

If there is no audio, silence is used. The
user can send events to the mixer to change the audio sources. The mixer will then change the audio sources and send the
new audio to the output.

### Events

During recording, you can send events to the mixer to change the audio sources. Events are sent using the gRPC API.
Events are sent using the following message

```protobuf
enum EventType {
  UNSPECIFIED = 0;
  PLAY = 1;
  PAUSE = 2;
  RESUME = 3;
  STOP = 4;
  SEEK = 5;
  VOLUME = 6;
  OTHER = 7;
}

message Event {
  // UUID of the record
  string recordId = 1;
    // UUID of the event
  string evtId = 2;
  // Type of event, see below
  EventType type = 3;
  // Direct URL to the resource to play
  string assetUrl = 4;
  // Whether to loop the audio when it ends
  bool loop = 5;
  // Volume change in decibels
  double volumeDeltaDb = 6;
  // Seek position in seconds
  int64 seekPositionSec = 7;
}
```

The mixer supports those types of events :

- PLAY: adds a new audio source to the mixer
- STOP: Removes an audio source from the Mixer
- PAUSE: Pauses an audio source that is currently playing in the mixer
- RESUME: Resumes an audio source that is currently paused in the mixer
- VOLUME : Changes the volume of an audio source in the mixer (in dB)
- SEEK: Seeks an audio source in the mixer to a specific position (in seconds)

## Example 

```bash
# t=0, Start a new record, the mixer streams silence to the output
grpcurl -plaintext -d '{"recordId": "my-record-id"}' localhost:50001 liveaudiomixer.EventStream/Start

# t=2, start playing a new audio source with id "demo-audio"
grpcurl -plaintext -d '{"recordId": "my-record-id", "evtId": "demo-audio", "type": "PLAY", "assetUrl": "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"}' localhost:50001 liveaudiomixer.EventStream/StreamEvents

# t=5, reduce the volume of the audio source "demo-audio" by 10dB
grpcurl -plaintext -d '{"recordId": "my-record-id", "evtId": "demo-audio", "type": "VOLUME", "volumeDeltaDb": -10}' localhost:50001 liveaudiomixer.EventStream/StreamEvents

# t=10, seek the audio source "demo-audio" to 30 seconds
grpcurl -plaintext -d '{"recordId": "my-record-id", "evtId": "demo-audio", "type": "SEEK", "seekPositionSec": 30}' localhost:50001 liveaudiomixer.EventStream/StreamEvents

# t=15, stop the audio source "demo-audio"
grpcurl -plaintext -d '{"recordId": "my-record-id", "evtId": "demo-audio", "type": "STOP"}' localhost:50001 liveaudiomixer.EventStream/StreamEvents

# t=20, stop the record
grpcurl -plaintext -d '{"recordId": "my-record-id"}' localhost:50001 liveaudiomixer.EventStream/Stop
```

```grpc
````
## Installation

### Docker

The easiest way to run the mixer is to use the docker image. You can find the image on [Docker Hub](https://hub.docker.com/r/sotrxii/live-audio-mixer).

```bash 
docker pull sotrxii/live-audio-mixer:latest
```

### From source

You can also build the mixer from source. You will need to have [Go](https://golang.org/) installed on your machine.

```bash
git clone <url>
cd live-audio-mixer
go build
```

## Configuration

The mixer can be configured using environment variables. The following variables are available:

| Variable name | Description                                                                                                                                               | Required | Default value  |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|----------|----------------|
| `SERVER_PORT` | Port the app is listening to                                                                                                                              |          | `4096`         |
| `DAPR_GRPC_PORT` | Port to connect to Dapr gRPC server. This variable is set automatically when running the app with dapr run.                                               | False    | `50001`        |
| `DAPR_MAX_REQUEST_SIZE_MB` | Maximum size for a payload in a Dapr request. This must be at least 4/3 of the max record size. 100MB should be enough for at least 8 to 10h of recording | False    | `100`          |
| `OBJECT_STORE_NAME` | Name of the Dapr component to use as an external object store                                                                                             | False    | `object-store` |
| `OBJECT_STORE_B64` | Whether to encode files to B64 before sending them to the object store component. This depend on which component is used. For S3, it's true               |          | `true`         |


## Development

### Run the app

```bash
make run
```