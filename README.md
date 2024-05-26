# Teams Voice Input

Backend of the Teams Voice Input project. 

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


## Generate Protobuf definitions (optional)

In case of API changes (`*.proto` files in `apis` directory),
you may regenerate Protobuf definitions for Tinkoff VoiceKit by running the following script:

```
./sh/generate_tinkoff_protos.sh
```
