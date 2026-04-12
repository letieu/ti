# ti

`ti` is a terminal-based coding assistant agent. It provides an interactive chat interface with tool-calling capabilities, allowing it to read, write, and edit files, as well as execute bash commands.

## Features

- **Interactive CLI**: A simple, terminal-based interface for chatting with the agent.
- **Tool Calling**: The agent can perform real-world tasks including:
  - Reading files (`read`)
  - Writing files (`write`)
  - Editing files with diff-based updates (`edit`)
  - Executing bash commands (`bash`)
- **Thinking Process**: Displays the agent's internal reasoning (thinking) in real-time.
- **File Selection**: Integration with `fzf` for quick file selection (triggered by `@`).
- **Authentication**: Built-in support for Google Antigravity/Vertex AI authentication.

## Installation

Ensure you have Go installed (version 1.25.8 or later).

```bash
go build -o ti ./cmd/ti
```

## Usage

### 1. Login

Before using the agent, you need to authenticate with the backend (Antigravity):

```bash
./ti
# Inside the app:
> /login
```

Follow the instructions in your browser to complete the authentication.

### 2. Chatting

Once logged in, you can start chatting:

```bash
> build a simple go web server
```

### 3. Special Commands & Shortcuts

- `/login`: Initiates the OAuth login flow.
- `@`: Triggers a file picker (fzf) to insert file paths into your message.
- `Ctrl+C`: Exit the application.

## Development

### Project Structure

- `cmd/ti`: Main entry point for the CLI application.
- `internal/agent`: Core agent logic and event handling.
- `internal/llm`: LLM providers (currently supports Antigravity/Gemini).
- `internal/tool`: Implementations of tools the agent can use (bash, edit, read, write).
- `internal/cli`: Terminal UI and command handling.
- `internal/auth`: Authentication logic.

### Environment Variables

- `TI_LOG_LEVEL`: Set the logging level (e.g., `debug`, `info`, `error`). Defaults to `error`.
- `TI_LOG_FORMAT`: Set to `json` for JSON formatted logs.
- `TI_MEMORY_DUMP`: Set to `true` to enable memory dumping for debugging.
- `TI_MEMORY_DUMP_PATH`: Path where the agent's memory is dumped. Defaults to `memory_dump.json` if `TI_MEMORY_DUMP` is `true`.

## License

[MIT](LICENSE) (or specify your license)
