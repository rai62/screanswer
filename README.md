# Screanswer

Screanswer is a command line tool designed to provide a convenient way of answering text on the screen for macOS users.
It is an easy-to-use tool that can be used for a variety of purposes, such as searching sentences or terms, copying texts to the clipboard, and more.

![screanswer](https://user-images.githubusercontent.com/83504221/236134459-532089b5-3e02-4a59-8ae4-044509163c95.gif)

## Features

Screanswer is designed to be fast, efficient, and user-friendly. It includes the following features:

- Capturing text on the screen and copying it to the clipboard
- Answering the text on the screen automatically using chat AI
- Reading the output aloud in a system voice

## Installation

### Homebrew

You can install Screanswer using `brew`:

```sh
brew tap rai62/screanswer
brew install screanswer
```

### Go install

Alternatively, you can install Screanswer using `go install`:

```sh
go install github.com/rai62/screanswer@latest
```

## Usage

Before using Screanswer, you need to set up the following environment settings for OpenAI API and GCP.

### Setting up OpenAI API Key for Answering Text

To enable text answering functionality, register your [API key](https://platform.openai.com/account/api-keys) for OpenAI API in the environment variable `OPENAI_API_KEY`.

```sh
export OPENAI_API_KEY=your_api_key
```

### Setting up GCP for Image-to-Text and Text-to-Speech

To enable image-to-text and text-to-speech functionality, create a GCP project and enable the following APIs:

- [Cloud Vision API](https://console.cloud.google.com/apis/api/vision.googleapis.com)
- [Cloud Text-to-Speech API](https://console.cloud.google.com/apis/api/texttospeech.googleapis.com)

Please refer to the [official documentation](https://cloud.google.com/vision/docs/setup) for more information.
You only need to complete the [Create a project](https://cloud.google.com/vision/docs/setup#project) and [Enable the API](https://cloud.google.com/vision/docs/setup#api) sections.

After setting up GCP, register your credentials for GCP in the environment variable `GOOGLE_APPLICATION_CREDENTIALS`.
Please see the [official documentation](https://cloud.google.com/docs/authentication/application-default-credentials).

```sh
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/credentials.json
```

### Commands

Here are some commands to use Screanswer:

- To capture, answer, and read text on the screen: `screanswer`
- To capture and copy text on the screen: `screanswer capture`
- To get an answer to text: `screanswer answer <text>`
- To read text aloud: `screanswer speak <languageCode> <text>`

### Examples

Here are some examples of how to use Screanswer:

- To capture, answer, and read text on the screen:

```sh
screanswer
```

This will capture text on the screen, answer it using chat AI, and then read the answer aloud in a system voice.

- To capture and copy text on the screen:

```sh
screanswer capture
```

This will capture text on the screen and copy it to the clipboard.

- To get an answer to text:

```sh
screanswer answer 'What is the capital of France?'
```

This will answer the specified text using chat AI and print the answer to the terminal.

- To read text aloud:

```sh
screanswer speak en-US 'Hello, world!'
```

This will read the specified text "Hello, world!" in English using a system voice.

## Supported Languages (Language Codes)

Screanswer currently supports the following languages:

- English (`en-US`)
- Japanese (`ja-JP`)
