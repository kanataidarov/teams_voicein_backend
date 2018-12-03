# Tinkoff VoiceKit examples

Based on https://github.com/Tinkoff/voicekit-examples \
You will need go with modules support and opus to run this examples.

## Usage

### Install the requirements

On MacOS:

```
$ brew install opus opusfile
```

On Ubuntu:

```
$ apt install libopus-dev libopusfile-dev
```

### Setup environment

Set `VOICEKIT_API_KEY` and `VOICEKIT_SECRET_KEY` environment variables to your API key and secret key to authenticate
your requests to VoiceKit:

```bash
export TINKOFF_API_KEY="API key from Tinkoff Software account"
export TINKOFF_SECRET_KEY="Secret key from Tinkoff Software account"
```


### Build example binaries

Build binaries:

```bash
go build cmd/recognize/recognize.go
go build cmd/recognize_stream/recognize_stream.go
```

### Basic recognition examples

Run basic (non-streaming) speech recognition example:

```bash
./recognize -e MPEG_AUDIO -r 48000 -c 1 -i ./audio/test1.mp3
```

To disable automatic punctuation and get up to 3 recognition alternatives:

```bash
./recognize -e MPEG_AUDIO -r 48000 -c 1 --disable-automatic-punctuation --max-alternatives 3 -i ./audio/test1.mp3
```

To disable profanity filter:

```bash
./recognize -e MPEG_AUDIO -r 48000 -c 1 --disable-profanity-filter -i ./audio/test1.mp3
```

Run streaming speech recognition with interim results:

```bash
./recognize_stream -e MPEG_AUDIO -r 48000 -c 1 --interim-results -i ./audio/test1.mp3
```

Specify longer silence timeout for voice activity detection (you will probably need longer audio to actually see the difference):

```bash
./recognize_stream -e MPEG_AUDIO -r 48000 -c 1 --interim-results --silence-duration-threshold 1.2 -i ./audio/test1.mp3
```

Return just the first recognized utterance and halt (you will probably need longer audio to actually see the difference):

```bash
./recognize_stream -e MPEG_AUDIO -r 48000 -c 1 --interim-results --single-utterance -i ./audio/test1.mp3
```


## Generate Protobuf and gRPC definitions (optional)

In case of API changes (`*.proto` files in `apis` directory),
you may regenerate Protobuf and gRPC definitions by simply running the following script:

```
./sh/generate_protobuf.sh
```
