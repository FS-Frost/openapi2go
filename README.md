# openapi2go

CLI tool to generate Go code from OpenAPI specifications.

-   For servers:

    -   Structs with JSON tags
    -   Functions using [Gin](https://github.com/gin-gonic/gin) for HTTP interactions

-   For clients:
    -   Structs with JSON tags

# Install

```sh
go install github.com/FS-Frost/openapi2go@latest
```

# Usage

```sh
openapi2go spec1.yml [specN.yml] outDir
```

