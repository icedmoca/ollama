# Ollama API Examples

Run the examples in this directory with:

```shell
go run example_name/main.go
```

## Chat - Chat with a model
- [chat/main.go](chat/main.go)

## Generate - Generate text from a model
- [generate/main.go](generate/main.go)
- [generate-streaming/main.go](generate-streaming/main.go)

## Multimodal - Work with images and text
- [multimodal/main.go](multimodal/main.go)

## Tokenize - Tokenize and detokenize text
- [tokenize/main.go](tokenize/main.go) - Basic tokenization example with keep_alive
- [tokenize/bench.go](tokenize/bench.go) - Performance benchmark tool with near-context-limit testing

## Pull - Pull a model
- [pull-progress/main.go](pull-progress/main.go)

