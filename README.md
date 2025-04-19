# MetamorphLLM - Code Rewriting for IT Security Research

This project is a framework for IT security research focusing on code rewrite and demonstration. It consists of three main modules:

1. **Suspicious Module**: A program that appears suspicious but is actually harmless. This serves as an example of code that might trigger security tools despite being benign.

2. **Rewriter Module**: A placeholder for a future engine that will analyze and rewrite Go code, enabling metamorphic code research.

3. **Manager Module**: An automation tool that orchestrates the rewriting, testing, and deployment process of metamorphic code.

## Project Structure

```
/
├── cmd/
│   ├── suspicious/     # CLI for the suspicious program
│   ├── rewriter/       # CLI for the rewriting engine
│   └── manager/        # CLI for the automation manager
├── internal/
│   ├── suspicious/     # Implementation of suspicious functionality
│   ├── rewriter/       # Implementation of rewriting engine (placeholder)
│   └── manager/        # Implementation of automation manager
```

## Building the Project

To build all packages:

```bash
go build ./...
```

## Environment Setup

The rewriter component requires a Gemini API key for LLM-based code transformation. You can set up your environment variables easily:

1. Create a copy of the example environment file:
   ```bash
   make setup-env
   # or manually:
   cp env.example .env
   ```

2. Edit the `.env` file with your actual API key:
   ```
   GEMINI_API_KEY=your_actual_key_here
   ```

3. The Makefile will automatically load these variables when running commands

## Running Tests

To run all tests:

```bash
go test ./internal/...
```

## Continuous Integration

This project uses GitHub Actions for continuous integration. Whenever code is pushed to the main branch or a pull request is created, the following checks are automatically run:

- Building the project
- Running all tests
- Linting the code for quality assurance

You can see the status of these checks in the GitHub repository.

## Usage

### Running the Suspicious Program

This program demonstrates code that appears malicious but is actually harmless. It contains functions with suspicious-sounding names that actually perform benign operations:

```bash
go run cmd/suspicious/main.go
```

The suspicious program includes examples of:
- Functions that appear to scan for vulnerabilities
- Code that looks like it's creating persistence mechanisms
- Functions that look like they're hiding or obfuscating data
- Operations that resemble command execution

All of these are implemented in a completely harmless way, making this a valuable tool for research into security tool behavior.

### Running the Rewriting Engine (Placeholder)

The rewriting engine is currently a placeholder that doesn't perform actual code transformation:

```bash
# Basic usage
go run cmd/rewriter/main.go -input path/to/file.go

# Specify output file
go run cmd/rewriter/main.go -input path/to/file.go -output path/to/output.go
```

### Running the Manager Tool

The manager tool automates the process of rewriting, testing, and deploying metamorphic code. By default, it targets the `internal/suspicious/suspicious.go` file for rewriting and builds the binary in `cmd/suspicious`:

```bash
# Basic usage (uses default paths)
# Rewrites internal/suspicious/suspicious.go
# Builds cmd/suspicious/suspicious
go run cmd/manager/main.go

# Specify custom source file and target binary directory
go run cmd/manager/main.go -suspicious ./src/my_code.go -target-dir ./bin/my_app -rewriter ./build/rewriter

# Dry run (no deployment)
go run cmd/manager/main.go -dry-run

# Don't keep rewritten files after deployment (by default they are kept)
go run cmd/manager/main.go -keep=false

# Force rewrite even if rewritten file exists
go run cmd/manager/main.go -force-rewrite
```

**Key Features:**
- By default, rewritten source files are kept and reused in subsequent runs
- The rewriter will only be invoked when a rewritten file doesn't exist yet, or with `-force-rewrite`
- Tests are run against the rewritten code, not the original
- The binary is built from the rewritten code

With the Makefile, you can simply run:

```bash
# Normal operation (uses defaults)
make run-manager

# Dry run
make run-manager-dry

# Force rewrite
make run-manager-force
```

## Scientific Research Context

This project is intended for academic research in the following areas:

1. **Code Obfuscation Techniques**: Demonstrating methods to transform code while preserving semantics
2. **Security Tool Testing**: Providing examples of code that might trigger false positives
3. **Metamorphic Code Analysis**: Studying self-modifying code patterns
4. **Resilience Against Static Analysis**: Exploring techniques that complicate static code analysis

## Note

This project is for academic and research purposes only. The "suspicious" code is deliberately designed to look suspicious while being harmless. It should not be used as a template for actual malicious software.

## License

See the [LICENSE](LICENSE) file for license rights and limitations.