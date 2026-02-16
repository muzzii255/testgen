# TestGen

TestGen is a CLI tool for generating and recording tests for Go web applications. It provides a proxy-based recording mechanism to capture HTTP traffic and automatically generates comprehensive test suites.

## Features

- **HTTP Traffic Recording**: Start a proxy server to intercept and record all HTTP requests and responses
- **Automatic Test Generation**: Generate Go test files from recorded traffic
- **Code Annotation Support**: Use `@testgen` comments to map endpoints to structs
- **CRUD Test Support**: Automatically generates tests for Create, Read, Update, and Delete operations
- **Struct Mapping**: Automatically maps JSON payloads to Go struct definitions

## Installation

```bash
go install github.com/yourusername/testgen@latest
```

Or clone and build from source:

```bash
git clone https://github.com/yourusername/testgen.git
cd testgen
go build -o testgen .
```

## Usage

### Recording HTTP Traffic

Start the proxy server to record HTTP traffic:

```bash
testgen record --port 9000 --target 8080
```

Options:
- `--port, -p`: Port to run the proxy server on (default: 9000)
- `--target, -t`: Target backend URL port (default: 8080)

The proxy will intercept requests and save recordings to the `./recordings` directory.

### Generating Tests

Generate test files from recorded JSON data:

```bash
testgen gen --file recordings/your-recording.json
```

Options:
- `--file, -f`: Path to the recorded JSON file (required)

### Code Annotations

Annotations are optional and can be placed anywhere in your codebase - not just directly above the struct definition. You can write them in the same file as your struct, a separate file, or even in a dedicated annotations file. TestGen will scan all `.go` files in your project to find them.

To map endpoints to structs for test generation, add `@testgen` comments in your Go code:

```go
// @testgen router=/api/v1/users struct=User
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

The annotation format is:
```
// @testgen router=<endpoint-path> struct=<struct-name>
```

For external packages:
```
// @testgen router=/api/v1/users struct=package.StructName
```

Example of placing annotations in a separate file:

```go
// annotations.go

// @testgen router=/api/v1/users struct=User
// @testgen router=/api/v1/products struct=Product
// @testgen router=/api/v1/orders struct=Order
```

## Project Structure

```
testgen/
├── cmd/                # CLI command implementations
│   ├── generate.go     # Generate command
│   ├── record.go       # Record command
│   └── root.go         # Root command
├── generator/          # Test generation logic
│   ├── codegen.go     # Code generation from recordings
│   └── generator.go   # Tag scanning and processing
├── proxy/             # HTTP proxy and recording
│   └── proxy.go       # Proxy server implementation
├── structgen/         # Struct parsing and mapping
│   └── structgen.go   # AST-based struct analysis
└── main.go            # Entry point
```

## Workflow

1. **Start your backend server** on port 8080 (or your preferred port)

2. **Start the recording proxy**:
   ```bash
   testgen record --port 9000 --target 8080
   ```

3. **Configure your application** to use the proxy (set `HTTP_PROXY=http://localhost:9000`)

4. **Make requests** to your application through the proxy

5. **Stop the proxy** - recordings will be saved automatically

6. **Add annotations** to your code structs:
   ```go
   // @testgen router=/api/v1/users struct=models.User
   type User struct {
       // ...
   }
   ```

7. **Generate tests**:
   ```bash
   testgen gen --file recordings/2026-01-01-users.json
   ```

8. **Find generated tests** in the `./gentest` directory

## Generated Test Structure

The generated tests include:
- Setup functions for test initialization
- Test cases for each HTTP method (GET, POST, PUT, DELETE)
- Status code assertions
- Payload mapping from JSON to Go structs

## Requirements

- Go 1.25 or later
- A Go web application to test

## Dependencies

- github.com/spf13/cobra - CLI framework
- golang.org/x/tools - Go tools for code analysis

## License

This project is licensed under the MIT License - see the LICENSE file for details.
