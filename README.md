# GitAI

GitAI is a Go-based command-line tool that uses AI to automatically generate meaningful commit messages for your Git repositories.

## Features

- Automatically detects changes in your Git repository
- Generates commit messages using the Groq API with LLaMA 3.1 70B model on Groq
- Follows Conventional Commits format for generated messages
- Handles both staged and unstaged changes, including new files
- Allows user confirmation before committing changes

## Prerequisites

- Go 1.23.0 or higher
- Git installed and configured on your system
- A Groq API key

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/arunsathiya/gitai.git
   ```

2. Change to the project directory:
   ```
   cd gitai
   ```

3. Install dependencies:
   ```
   go mod download
   ```

4. Build the project:
   ```
   go build
   ```

## Configuration

1. Create a `.gitai.env` file in your home directory:
   ```
   touch ~/.gitai.env
   ```

2. Add your Groq API key to the `.gitai.env` file:
   ```
   GROQ_API_KEY=your_api_key_here
   ```

## Usage

Run the GitAI tool from your Git repository:

```
/path/to/gitai
```

The tool will:
1. Detect changes in your repository
2. Generate a commit message using AI
3. Show you the generated message and ask for confirmation
4. If confirmed, stage all changes and create a commit with the generated message

## How It Works

1. GitAI uses the go-git library to interact with your Git repository.
2. It generates a diff of the changes in your working directory.
3. The diff is sent to the Groq API, which uses a large language model to generate a commit message.
4. The generated message is presented to you for confirmation.
5. If you approve, GitAI stages all changes and creates a commit with the generated message.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Disclaimer

This tool uses AI to generate commit messages. While it aims to produce meaningful and accurate messages, always review the generated messages before confirming the commit to ensure they accurately represent your changes.